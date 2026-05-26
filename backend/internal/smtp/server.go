package smtp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/mail"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/inst-dev/webhook/internal/config"
	"github.com/inst-dev/webhook/internal/database"
	"github.com/inst-dev/webhook/internal/models"
	"github.com/inst-dev/webhook/internal/redis"
	gosmtp "github.com/emersion/go-smtp"
	"github.com/rs/zerolog/log"
)

// Server is the SMTP email hook server
type Server struct {
	cfg    *config.Config
	db     *database.Pool
	rdb    *redis.Client
	server *gosmtp.Server
}

// NewServer creates a new SMTP server
func NewServer(cfg *config.Config, db *database.Pool, rdb *redis.Client) *Server {
	return &Server{cfg: cfg, db: db, rdb: rdb}
}

// Start starts the SMTP server
func (s *Server) Start(ctx context.Context) error {
	backend := &smtpBackend{
		cfg: s.cfg,
		db:  s.db,
		rdb: s.rdb,
	}

	s.server = gosmtp.NewServer(backend)
	s.server.Addr = fmt.Sprintf(":%d", s.cfg.SMTPPort)
	s.server.Domain = s.cfg.SMTPDomainName
	s.server.MaxMessageBytes = 10 * 1024 * 1024 // 10MB
	s.server.MaxRecipients = 50
	s.server.AllowInsecureAuth = true
	s.server.ReadTimeout = 30 * time.Second
	s.server.WriteTimeout = 30 * time.Second

	log.Info().Str("addr", s.server.Addr).Msg("SMTP server starting")
	return s.server.ListenAndServe()
}

// Stop stops the SMTP server
func (s *Server) Stop() {
	if s.server != nil {
		s.server.Close()
	}
}

// smtpBackend implements the SMTP backend
type smtpBackend struct {
	cfg *config.Config
	db  *database.Pool
	rdb *redis.Client
}

func (b *smtpBackend) NewSession(c *gosmtp.Conn) (gosmtp.Session, error) {
	return &smtpSession{
		cfg:      b.cfg,
		db:       b.db,
		rdb:      b.rdb,
		sourceIP: c.Conn().RemoteAddr().String(),
	}, nil
}

// smtpSession handles a single SMTP session
type smtpSession struct {
	cfg      *config.Config
	db       *database.Pool
	rdb      *redis.Client
	from     string
	to       []string
	sourceIP string
}

func (s *smtpSession) AuthPlain(username, password string) error {
	return nil // Accept all auth for inbound
}

func (s *smtpSession) Mail(from string, opts *gosmtp.MailOptions) error {
	s.from = from
	return nil
}

func (s *smtpSession) Rcpt(to string, opts *gosmtp.RcptOptions) error {
	s.to = append(s.to, to)
	return nil
}

func (s *smtpSession) Data(r io.Reader) error {
	// Read the full message
	rawData, err := io.ReadAll(io.LimitReader(r, 10*1024*1024))
	if err != nil {
		return err
	}

	// Parse the email
	msg, err := mail.ReadMessage(bytes.NewReader(rawData))
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse email")
		// Still store it raw
		return s.storeRaw(rawData)
	}

	subject := msg.Header.Get("Subject")
	contentType := msg.Header.Get("Content-Type")

	// Parse body
	body, htmlBody, attachments := s.parseBody(msg, contentType)

	// Parse headers
	headers := make(map[string]string)
	for key, values := range msg.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}
	headersJSON, _ := json.Marshal(headers)
	attachmentsJSON, _ := json.Marshal(attachments)

	// Extract token from recipient
	for _, to := range s.to {
		token := extractTokenFromEmail(to, s.cfg.SMTPDomainName)
		endpointID := s.resolveEndpoint(token)

		emailLog := &models.EmailLog{
			ID:          uuid.New(),
			EndpointID:  endpointID,
			From:        s.from,
			To:          to,
			Subject:     subject,
			Body:        body,
			HTMLBody:    htmlBody,
			RawMessage:  rawData,
			Headers:     headersJSON,
			Attachments: attachmentsJSON,
			SourceIP:    s.sourceIP,
			Size:        int64(len(rawData)),
			CreatedAt:   time.Now(),
		}

		s.storeEmail(emailLog)
	}

	return nil
}

func (s *smtpSession) Reset() {
	s.from = ""
	s.to = nil
}

func (s *smtpSession) Logout() error {
	return nil
}

func (s *smtpSession) storeRaw(data []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for _, to := range s.to {
		token := extractTokenFromEmail(to, s.cfg.SMTPDomainName)
		endpointID := s.resolveEndpoint(token)

		emailLog := &models.EmailLog{
			ID:         uuid.New(),
			EndpointID: endpointID,
			From:       s.from,
			To:         to,
			RawMessage: data,
			SourceIP:   s.sourceIP,
			Size:       int64(len(data)),
			CreatedAt:  time.Now(),
		}

		_, err := s.db.Exec(ctx, `
			INSERT INTO email_logs (id, endpoint_id, from_addr, to_addr, raw_message, source_ip, size, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`, emailLog.ID, emailLog.EndpointID, emailLog.From, emailLog.To,
			emailLog.RawMessage, emailLog.SourceIP, emailLog.Size, emailLog.CreatedAt)

		if err != nil {
			log.Error().Err(err).Msg("Failed to store raw email")
		}
	}
	return nil
}

func (s *smtpSession) storeEmail(emailLog *models.EmailLog) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := s.db.Exec(ctx, `
		INSERT INTO email_logs (id, endpoint_id, from_addr, to_addr, subject, body, html_body, 
			raw_message, headers, attachments, source_ip, size, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`, emailLog.ID, emailLog.EndpointID, emailLog.From, emailLog.To, emailLog.Subject,
		emailLog.Body, emailLog.HTMLBody, emailLog.RawMessage, emailLog.Headers,
		emailLog.Attachments, emailLog.SourceIP, emailLog.Size, emailLog.CreatedAt)

	if err != nil {
		log.Error().Err(err).Msg("Failed to store email")
		return
	}

	// Publish realtime event
	event, _ := json.Marshal(map[string]interface{}{
		"type":    "email_received",
		"from":    emailLog.From,
		"to":      emailLog.To,
		"subject": emailLog.Subject,
	})
	s.rdb.Publish(ctx, "events:email", string(event))
}

func (s *smtpSession) resolveEndpoint(token string) uuid.UUID {
	if token == "" {
		return uuid.Nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cached, err := s.rdb.Get(ctx, "endpoint:token:"+token).Result()
	if err == nil {
		id, _ := uuid.Parse(cached)
		return id
	}
	return uuid.Nil
}

func (s *smtpSession) parseBody(msg *mail.Message, contentType string) (string, string, []Attachment) {
	var textBody, htmlBody string
	var attachments []Attachment

	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		// Simple body
		body, _ := io.ReadAll(msg.Body)
		return string(body), "", nil
	}

	if strings.HasPrefix(mediaType, "multipart/") {
		mr := multipart.NewReader(msg.Body, params["boundary"])
		for {
			part, err := mr.NextPart()
			if err != nil {
				break
			}

			partType := part.Header.Get("Content-Type")
			data, _ := io.ReadAll(io.LimitReader(part, 5*1024*1024))

			if strings.Contains(partType, "text/plain") {
				textBody = string(data)
			} else if strings.Contains(partType, "text/html") {
				htmlBody = string(data)
			} else if part.FileName() != "" {
				attachments = append(attachments, Attachment{
					Filename:    part.FileName(),
					ContentType: partType,
					Size:        int64(len(data)),
				})
			}
		}
	} else {
		body, _ := io.ReadAll(msg.Body)
		if strings.Contains(mediaType, "html") {
			htmlBody = string(body)
		} else {
			textBody = string(body)
		}
	}

	return textBody, htmlBody, attachments
}

// Attachment represents an email attachment metadata
type Attachment struct {
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	Size        int64  `json:"size"`
}

// extractTokenFromEmail extracts the token from an email address
func extractTokenFromEmail(email, domain string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return ""
	}
	if !strings.Contains(parts[1], domain) {
		return ""
	}
	return parts[0]
}

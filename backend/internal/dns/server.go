package dns

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/inst-dev/webhook/internal/config"
	"github.com/inst-dev/webhook/internal/database"
	"github.com/inst-dev/webhook/internal/models"
	"github.com/inst-dev/webhook/internal/redis"
	mdns "github.com/miekg/dns"
	"github.com/rs/zerolog/log"
)

// Server is the DNS interaction logging server
type Server struct {
	cfg      *config.Config
	db       *database.Pool
	rdb      *redis.Client
	server   *mdns.Server
	serverV6 *mdns.Server
}

// NewServer creates a new DNS server
func NewServer(cfg *config.Config, db *database.Pool, rdb *redis.Client) *Server {
	return &Server{cfg: cfg, db: db, rdb: rdb}
}

// Start starts the DNS server
func (s *Server) Start(ctx context.Context) error {
	handler := &dnsHandler{
		cfg: s.cfg,
		db:  s.db,
		rdb: s.rdb,
	}

	addr := fmt.Sprintf(":%d", s.cfg.DNSPort)

	// UDP server
	s.server = &mdns.Server{
		Addr:    addr,
		Net:     "udp",
		Handler: handler,
	}

	// TCP server
	s.serverV6 = &mdns.Server{
		Addr:    addr,
		Net:     "tcp",
		Handler: handler,
	}

	go func() {
		log.Info().Str("addr", addr).Msg("DNS UDP server starting")
		if err := s.server.ListenAndServe(); err != nil {
			log.Error().Err(err).Msg("DNS UDP server error")
		}
	}()

	go func() {
		log.Info().Str("addr", addr).Msg("DNS TCP server starting")
		if err := s.serverV6.ListenAndServe(); err != nil {
			log.Error().Err(err).Msg("DNS TCP server error")
		}
	}()

	return nil
}

// Stop stops the DNS server
func (s *Server) Stop() {
	if s.server != nil {
		s.server.Shutdown()
	}
	if s.serverV6 != nil {
		s.serverV6.Shutdown()
	}
}

// dnsHandler handles DNS queries
type dnsHandler struct {
	cfg *config.Config
	db  *database.Pool
	rdb *redis.Client
}

func (h *dnsHandler) ServeDNS(w mdns.ResponseWriter, r *mdns.Msg) {
	msg := new(mdns.Msg)
	msg.SetReply(r)
	msg.Authoritative = true

	for _, q := range r.Question {
		log.Debug().
			Str("name", q.Name).
			Str("type", mdns.TypeToString[q.Qtype]).
			Msg("DNS query received")

		// Log the interaction
		go h.logQuery(q, w.RemoteAddr())

		// Respond based on query type
		switch q.Qtype {
		case mdns.TypeA:
			msg.Answer = append(msg.Answer, &mdns.A{
				Hdr: mdns.RR_Header{
					Name:   q.Name,
					Rrtype: mdns.TypeA,
					Class:  mdns.ClassINET,
					Ttl:    60,
				},
				A: net.ParseIP("127.0.0.1"),
			})
		case mdns.TypeAAAA:
			msg.Answer = append(msg.Answer, &mdns.AAAA{
				Hdr: mdns.RR_Header{
					Name:   q.Name,
					Rrtype: mdns.TypeAAAA,
					Class:  mdns.ClassINET,
					Ttl:    60,
				},
				AAAA: net.ParseIP("::1"),
			})
		case mdns.TypeTXT:
			msg.Answer = append(msg.Answer, &mdns.TXT{
				Hdr: mdns.RR_Header{
					Name:   q.Name,
					Rrtype: mdns.TypeTXT,
					Class:  mdns.ClassINET,
					Ttl:    60,
				},
				Txt: []string{"webhook.inst.lk dns interaction"},
			})
		case mdns.TypeMX:
			msg.Answer = append(msg.Answer, &mdns.MX{
				Hdr: mdns.RR_Header{
					Name:   q.Name,
					Rrtype: mdns.TypeMX,
					Class:  mdns.ClassINET,
					Ttl:    60,
				},
				Preference: 10,
				Mx:         "mail." + h.cfg.Domain + ".",
			})
		}
	}

	w.WriteMsg(msg)
}

// logQuery logs a DNS query to the database
func (h *dnsHandler) logQuery(q mdns.Question, addr net.Addr) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Extract token from query name
	// Expected format: <data>.<token>.dns.webhook.inst.lk
	name := strings.TrimSuffix(q.Name, ".")
	parts := strings.Split(name, ".")
	
	var endpointToken string
	dnsZone := h.cfg.DNSDomainName
	zoneParts := strings.Split(dnsZone, ".")
	
	if len(parts) > len(zoneParts) {
		// Token is the part before the zone
		endpointToken = parts[len(parts)-len(zoneParts)-1]
	}

	// Get source IP and port
	sourceIP := ""
	sourcePort := 0
	if udpAddr, ok := addr.(*net.UDPAddr); ok {
		sourceIP = udpAddr.IP.String()
		sourcePort = udpAddr.Port
	} else if tcpAddr, ok := addr.(*net.TCPAddr); ok {
		sourceIP = tcpAddr.IP.String()
		sourcePort = tcpAddr.Port
	}

	// Look up endpoint by token
	var endpointID uuid.UUID
	if endpointToken != "" {
		cached, err := h.rdb.Get(ctx, "endpoint:token:"+endpointToken).Result()
		if err == nil {
			endpointID, _ = uuid.Parse(cached)
		}
	}

	dnsLog := &models.DNSLog{
		ID:         uuid.New(),
		EndpointID: endpointID,
		QueryName:  q.Name,
		QueryType:  mdns.TypeToString[q.Qtype],
		SourceIP:   sourceIP,
		SourcePort: sourcePort,
		CreatedAt:  time.Now(),
	}

	_, err := h.db.Exec(ctx, `
		INSERT INTO dns_logs (id, endpoint_id, query_name, query_type, source_ip, source_port, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, dnsLog.ID, dnsLog.EndpointID, dnsLog.QueryName, dnsLog.QueryType,
		dnsLog.SourceIP, dnsLog.SourcePort, dnsLog.CreatedAt)

	if err != nil {
		log.Error().Err(err).Msg("Failed to log DNS query")
		return
	}

	// Publish event for realtime
	event, _ := json.Marshal(map[string]interface{}{
		"type":       "dns_query",
		"query_name": q.Name,
		"query_type": mdns.TypeToString[q.Qtype],
		"source_ip":  sourceIP,
		"timestamp":  time.Now().Format(time.RFC3339),
	})
	h.rdb.Publish(ctx, "events:dns", string(event))
}

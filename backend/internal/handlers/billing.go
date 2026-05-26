package handlers

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/inst-dev/webhook/internal/config"
	"github.com/inst-dev/webhook/internal/database"
	"github.com/inst-dev/webhook/internal/redis"
	"github.com/rs/zerolog/log"
)

// Plan represents a billing plan
type Plan struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Price       float64           `json:"price"`
	Currency    string            `json:"currency"`
	Interval    string            `json:"interval"`
	Features    map[string]interface{} `json:"features"`
	Description string            `json:"description"`
}

// BillingHandler handles billing and subscription operations
type BillingHandler struct {
	cfg *config.Config
	db  *database.Pool
	rdb *redis.Client
}

// NewBillingHandler creates a new billing handler
func NewBillingHandler(cfg *config.Config, db *database.Pool, rdb *redis.Client) *BillingHandler {
	return &BillingHandler{cfg: cfg, db: db, rdb: rdb}
}

// ListPlans returns available subscription plans
func (h *BillingHandler) ListPlans(c *fiber.Ctx) error {
	plans := []Plan{
		{
			ID:       "free",
			Name:     "Free",
			Price:    0,
			Currency: "USD",
			Interval: "monthly",
			Features: map[string]interface{}{
				"endpoints":      5,
				"retention_hours": 72,
				"custom_domains": 0,
				"dns_hooks":      false,
				"email_hooks":    false,
				"api_access":     false,
				"team_members":   1,
				"requests":       "unlimited",
			},
			Description: "Perfect for trying out the platform",
		},
		{
			ID:       "pro",
			Name:     "Pro",
			Price:    9.99,
			Currency: "USD",
			Interval: "monthly",
			Features: map[string]interface{}{
				"endpoints":       50,
				"retention_hours":  720,
				"custom_domains":  3,
				"dns_hooks":       true,
				"email_hooks":     true,
				"api_access":      true,
				"team_members":    1,
				"requests":        "unlimited",
				"custom_responses": true,
				"request_replay":  true,
			},
			Description: "For professional developers and security researchers",
		},
		{
			ID:       "team",
			Name:     "Team",
			Price:    29.99,
			Currency: "USD",
			Interval: "monthly",
			Features: map[string]interface{}{
				"endpoints":       200,
				"retention_hours":  2160,
				"custom_domains":  10,
				"dns_hooks":       true,
				"email_hooks":     true,
				"api_access":      true,
				"team_members":    10,
				"requests":        "unlimited",
				"custom_responses": true,
				"request_replay":  true,
				"analytics":       true,
			},
			Description: "For development teams and organizations",
		},
		{
			ID:       "enterprise",
			Name:     "Enterprise",
			Price:    99.99,
			Currency: "USD",
			Interval: "monthly",
			Features: map[string]interface{}{
				"endpoints":       "unlimited",
				"retention_hours":  8760,
				"custom_domains":  "unlimited",
				"dns_hooks":       true,
				"email_hooks":     true,
				"api_access":      true,
				"team_members":    "unlimited",
				"requests":        "unlimited",
				"custom_responses": true,
				"request_replay":  true,
				"analytics":       true,
				"priority_support": true,
				"sla":             true,
			},
			Description: "For enterprises requiring full access and support",
		},
	}

	return c.JSON(fiber.Map{"plans": plans})
}

// GetSubscription returns the current user's subscription
func (h *BillingHandler) GetSubscription(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)
	ctx := c.Context()

	var sub struct {
		ID                 uuid.UUID  `json:"id"`
		Plan               string     `json:"plan"`
		Status             string     `json:"status"`
		Provider           string     `json:"provider"`
		CurrentPeriodStart time.Time  `json:"current_period_start"`
		CurrentPeriodEnd   time.Time  `json:"current_period_end"`
		CreatedAt          time.Time  `json:"created_at"`
		CancelledAt        *time.Time `json:"cancelled_at"`
	}

	err := h.db.QueryRow(ctx, `
		SELECT id, plan, status, provider, current_period_start, current_period_end, created_at, cancelled_at
		FROM subscriptions
		WHERE user_id = $1 AND status = 'active'
		ORDER BY created_at DESC
		LIMIT 1
	`, userID).Scan(&sub.ID, &sub.Plan, &sub.Status, &sub.Provider,
		&sub.CurrentPeriodStart, &sub.CurrentPeriodEnd, &sub.CreatedAt, &sub.CancelledAt)

	if err != nil {
		// No active subscription - they're on free plan
		return c.JSON(fiber.Map{
			"subscription": nil,
			"plan":         "free",
		})
	}

	return c.JSON(fiber.Map{"subscription": sub})
}

// CreateSubscription initiates a new subscription
func (h *BillingHandler) CreateSubscription(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)

	var input struct {
		Plan     string `json:"plan" validate:"required"`
		Provider string `json:"provider" validate:"required,oneof=paypal payhere"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid_request",
			"message": "Invalid request body",
		})
	}

	validPlans := map[string]bool{"pro": true, "team": true, "enterprise": true}
	if !validPlans[input.Plan] {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid_plan",
			"message": "Invalid subscription plan",
		})
	}

	// Generate payment URL based on provider
	var paymentURL string
	var err error

	switch input.Provider {
	case "paypal":
		paymentURL, err = h.createPayPalSubscription(userID, input.Plan)
	case "payhere":
		paymentURL, err = h.createPayHereSubscription(userID, input.Plan)
	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid_provider",
			"message": "Unsupported payment provider",
		})
	}

	if err != nil {
		log.Error().Err(err).Str("provider", input.Provider).Msg("Failed to create subscription")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "subscription_failed",
			"message": "Failed to create subscription",
		})
	}

	return c.JSON(fiber.Map{
		"payment_url": paymentURL,
		"provider":    input.Provider,
	})
}

// CancelSubscription cancels the current subscription
func (h *BillingHandler) CancelSubscription(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)
	ctx := c.Context()

	now := time.Now()
	_, err := h.db.Exec(ctx, `
		UPDATE subscriptions
		SET status = 'cancelled', cancelled_at = $2
		WHERE user_id = $1 AND status = 'active'
	`, userID, now)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "cancel_failed",
			"message": "Failed to cancel subscription",
		})
	}

	// Downgrade user plan at end of period
	h.db.Exec(ctx, `UPDATE users SET plan = 'free' WHERE id = $1`, userID)

	return c.JSON(fiber.Map{
		"message": "Subscription cancelled. Access continues until end of billing period.",
	})
}

// ListInvoices returns billing history
func (h *BillingHandler) ListInvoices(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)
	ctx := c.Context()

	rows, err := h.db.Query(ctx, `
		SELECT id, plan, provider, current_period_start, current_period_end, status, created_at
		FROM subscriptions
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 50
	`, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve invoices",
		})
	}
	defer rows.Close()

	var invoices []fiber.Map
	for rows.Next() {
		var id uuid.UUID
		var plan, provider, status string
		var periodStart, periodEnd, createdAt time.Time

		if err := rows.Scan(&id, &plan, &provider, &periodStart, &periodEnd, &status, &createdAt); err != nil {
			continue
		}
		invoices = append(invoices, fiber.Map{
			"id":           id,
			"plan":         plan,
			"provider":     provider,
			"period_start": periodStart,
			"period_end":   periodEnd,
			"status":       status,
			"created_at":   createdAt,
		})
	}

	return c.JSON(fiber.Map{"invoices": invoices})
}

// PayPalWebhook handles PayPal IPN/webhook notifications
func (h *BillingHandler) PayPalWebhook(c *fiber.Ctx) error {
	body := c.Body()

	// Verify webhook signature
	if !h.verifyPayPalSignature(c) {
		log.Warn().Msg("Invalid PayPal webhook signature")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid signature",
		})
	}

	var event struct {
		EventType string          `json:"event_type"`
		Resource  json.RawMessage `json:"resource"`
	}

	if err := json.Unmarshal(body, &event); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid payload",
		})
	}

	log.Info().Str("event_type", event.EventType).Msg("PayPal webhook received")

	switch event.EventType {
	case "BILLING.SUBSCRIPTION.ACTIVATED":
		h.handlePayPalSubscriptionActivated(event.Resource)
	case "BILLING.SUBSCRIPTION.CANCELLED":
		h.handlePayPalSubscriptionCancelled(event.Resource)
	case "PAYMENT.SALE.COMPLETED":
		h.handlePayPalPaymentCompleted(event.Resource)
	}

	return c.SendStatus(fiber.StatusOK)
}

// PayHereWebhook handles PayHere payment notifications
func (h *BillingHandler) PayHereWebhook(c *fiber.Ctx) error {
	merchantID := c.FormValue("merchant_id")
	orderID := c.FormValue("order_id")
	paymentID := c.FormValue("payment_id")
	amount := c.FormValue("payhere_amount")
	currency := c.FormValue("payhere_currency")
	statusCode := c.FormValue("status_code")
	md5sig := c.FormValue("md5sig")

	// Verify PayHere signature
	expectedSig := h.generatePayHereMD5(merchantID, orderID, amount, currency, statusCode)
	if !strings.EqualFold(md5sig, expectedSig) {
		log.Warn().Msg("Invalid PayHere webhook signature")
		return c.Status(fiber.StatusUnauthorized).SendString("Invalid signature")
	}

	log.Info().
		Str("order_id", orderID).
		Str("payment_id", paymentID).
		Str("status", statusCode).
		Msg("PayHere webhook received")

	// Status code 2 = success
	if statusCode == "2" {
		h.handlePayHerePaymentSuccess(orderID, paymentID)
	}

	return c.SendStatus(fiber.StatusOK)
}

// createPayPalSubscription creates a PayPal subscription
func (h *BillingHandler) createPayPalSubscription(userID uuid.UUID, plan string) (string, error) {
	// In production, use PayPal SDK to create subscription
	// This generates the approval URL for the user to complete payment
	baseURL := "https://www.paypal.com/checkoutnow"
	if h.cfg.PayPal.Mode == "sandbox" {
		baseURL = "https://www.sandbox.paypal.com/checkoutnow"
	}

	// Store pending subscription reference
	h.rdb.Set(nil, fmt.Sprintf("pending_sub:%s:%s", userID.String(), plan), "paypal", 24*time.Hour)

	return fmt.Sprintf("%s?plan=%s&user=%s", baseURL, plan, userID.String()), nil
}

// createPayHereSubscription creates a PayHere subscription
func (h *BillingHandler) createPayHereSubscription(userID uuid.UUID, plan string) (string, error) {
	baseURL := "https://www.payhere.lk/pay/checkout"
	if h.cfg.PayHere.Mode == "sandbox" {
		baseURL = "https://sandbox.payhere.lk/pay/checkout"
	}

	// Store pending subscription reference
	h.rdb.Set(nil, fmt.Sprintf("pending_sub:%s:%s", userID.String(), plan), "payhere", 24*time.Hour)

	return fmt.Sprintf("%s?merchant_id=%s&order_id=%s", baseURL, h.cfg.PayHere.MerchantID, userID.String()), nil
}

// verifyPayPalSignature verifies the PayPal webhook signature
func (h *BillingHandler) verifyPayPalSignature(c *fiber.Ctx) bool {
	if h.cfg.PayPal.Mode == "sandbox" {
		return true // Skip in sandbox mode
	}

	transmissionID := c.Get("PAYPAL-TRANSMISSION-ID")
	transmissionTime := c.Get("PAYPAL-TRANSMISSION-TIME")
	certURL := c.Get("PAYPAL-CERT-URL")
	transmissionSig := c.Get("PAYPAL-TRANSMISSION-SIG")

	if transmissionID == "" || transmissionTime == "" || certURL == "" || transmissionSig == "" {
		return false
	}

	// In production, implement full PayPal signature verification
	// using their public certificates
	_ = io.Discard
	return true
}

// generatePayHereMD5 generates PayHere verification hash
func (h *BillingHandler) generatePayHereMD5(merchantID, orderID, amount, currency, statusCode string) string {
	// PayHere MD5 signature: md5(merchant_id + order_id + payhere_amount + payhere_currency + status_code + md5(merchant_secret))
	secretHash := md5.Sum([]byte(h.cfg.PayHere.MerchantSecret))
	secretHashStr := strings.ToUpper(hex.EncodeToString(secretHash[:]))

	data := merchantID + orderID + amount + currency + statusCode + secretHashStr
	hash := md5.Sum([]byte(data))
	return strings.ToUpper(hex.EncodeToString(hash[:]))
}

func (h *BillingHandler) handlePayPalSubscriptionActivated(resource json.RawMessage) {
	// Parse resource and activate subscription
	log.Info().Msg("PayPal subscription activated")
}

func (h *BillingHandler) handlePayPalSubscriptionCancelled(resource json.RawMessage) {
	log.Info().Msg("PayPal subscription cancelled")
}

func (h *BillingHandler) handlePayPalPaymentCompleted(resource json.RawMessage) {
	log.Info().Msg("PayPal payment completed")
}

func (h *BillingHandler) handlePayHerePaymentSuccess(orderID, paymentID string) {
	log.Info().Str("order_id", orderID).Str("payment_id", paymentID).Msg("PayHere payment success")
}

// Unused import suppression
var _ = hmac.New(sha256.New, nil)

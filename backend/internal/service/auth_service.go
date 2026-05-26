package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/inst-dev/webhook/internal/config"
	"github.com/inst-dev/webhook/internal/models"
	"github.com/inst-dev/webhook/internal/redis"
	"github.com/inst-dev/webhook/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrEmailTaken         = errors.New("email already registered")
	ErrUserNotFound       = errors.New("user not found")
	ErrTokenExpired       = errors.New("token expired")
	ErrTokenInvalid       = errors.New("invalid token")
)

// AuthService handles authentication logic
type AuthService struct {
	cfg      *config.Config
	userRepo *repository.UserRepository
	rdb      *redis.Client
}

// NewAuthService creates a new auth service
func NewAuthService(cfg *config.Config, userRepo *repository.UserRepository, rdb *redis.Client) *AuthService {
	return &AuthService{
		cfg:      cfg,
		userRepo: userRepo,
		rdb:      rdb,
	}
}

// RegisterInput holds registration data
type RegisterInput struct {
	Email       string `json:"email" validate:"required,email,max=255"`
	Password    string `json:"password" validate:"required,min=8,max=128"`
	DisplayName string `json:"display_name" validate:"required,min=2,max=100"`
}

// LoginInput holds login data
type LoginInput struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// TokenPair holds access and refresh tokens
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// Register creates a new user account
func (s *AuthService) Register(ctx context.Context, input RegisterInput) (*models.User, *TokenPair, error) {
	// Check if email is taken
	_, err := s.userRepo.GetByEmail(ctx, input.Email)
	if err == nil {
		return nil, nil, ErrEmailTaken
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &models.User{
		Email:        input.Email,
		PasswordHash: string(hashedPassword),
		DisplayName:  input.DisplayName,
		Plan:         "free",
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		if errors.Is(err, repository.ErrAlreadyExists) {
			return nil, nil, ErrEmailTaken
		}
		return nil, nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Generate tokens
	tokens, err := s.generateTokenPair(user.ID)
	if err != nil {
		return nil, nil, err
	}

	// Store refresh token in Redis
	s.storeRefreshToken(ctx, user.ID, tokens.RefreshToken)

	// Generate email verification token
	s.generateVerificationToken(ctx, user.ID)

	return user, tokens, nil
}

// Login authenticates a user
func (s *AuthService) Login(ctx context.Context, input LoginInput) (*models.User, *TokenPair, error) {
	user, err := s.userRepo.GetByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, nil, ErrInvalidCredentials
		}
		return nil, nil, err
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return nil, nil, ErrInvalidCredentials
	}

	// Update last login
	now := time.Now()
	user.LastLoginAt = &now
	s.userRepo.Update(ctx, user)

	// Generate tokens
	tokens, err := s.generateTokenPair(user.ID)
	if err != nil {
		return nil, nil, err
	}

	// Store refresh token
	s.storeRefreshToken(ctx, user.ID, tokens.RefreshToken)

	return user, tokens, nil
}

// RefreshToken issues a new token pair using a refresh token
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error) {
	// Validate refresh token exists in Redis
	userIDStr, err := s.rdb.Get(ctx, "refresh:"+refreshToken).Result()
	if err != nil {
		return nil, ErrTokenInvalid
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, ErrTokenInvalid
	}

	// Revoke old refresh token
	s.rdb.Del(ctx, "refresh:"+refreshToken)

	// Generate new token pair
	tokens, err := s.generateTokenPair(userID)
	if err != nil {
		return nil, err
	}

	// Store new refresh token
	s.storeRefreshToken(ctx, userID, tokens.RefreshToken)

	return tokens, nil
}

// Logout invalidates the current session
func (s *AuthService) Logout(ctx context.Context, accessToken, refreshToken string) error {
	// Blacklist access token
	s.rdb.Set(ctx, "blacklist:"+accessToken, "1", s.cfg.JWT.AccessExpiry)

	// Remove refresh token
	if refreshToken != "" {
		s.rdb.Del(ctx, "refresh:"+refreshToken)
	}

	return nil
}

// ValidateToken validates an access token and returns the user ID
func (s *AuthService) ValidateToken(tokenString string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrTokenInvalid
		}
		return []byte(s.cfg.JWT.Secret), nil
	})

	if err != nil || !token.Valid {
		return uuid.Nil, ErrTokenInvalid
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok {
		return uuid.Nil, ErrTokenInvalid
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, ErrTokenInvalid
	}

	return userID, nil
}

// GetUser gets a user by ID
func (s *AuthService) GetUser(ctx context.Context, id uuid.UUID) (*models.User, error) {
	return s.userRepo.GetByID(ctx, id)
}

// generateTokenPair creates access + refresh tokens
func (s *AuthService) generateTokenPair(userID uuid.UUID) (*TokenPair, error) {
	now := time.Now()
	expiresAt := now.Add(s.cfg.JWT.AccessExpiry)

	// Access token
	accessClaims := &jwt.RegisteredClaims{
		Subject:   userID.String(),
		Issuer:    s.cfg.JWT.Issuer,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(expiresAt),
		ID:        uuid.New().String(),
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(s.cfg.JWT.Secret))
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	// Refresh token (random string)
	refreshBytes := make([]byte, 32)
	if _, err := rand.Read(refreshBytes); err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}
	refreshTokenString := hex.EncodeToString(refreshBytes)

	return &TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		ExpiresAt:    expiresAt,
	}, nil
}

// storeRefreshToken stores a refresh token in Redis
func (s *AuthService) storeRefreshToken(ctx context.Context, userID uuid.UUID, token string) {
	s.rdb.Set(ctx, "refresh:"+token, userID.String(), s.cfg.JWT.RefreshExpiry)
}

// generateVerificationToken creates an email verification token
func (s *AuthService) generateVerificationToken(ctx context.Context, userID uuid.UUID) string {
	tokenBytes := make([]byte, 32)
	rand.Read(tokenBytes)
	token := hex.EncodeToString(tokenBytes)
	s.rdb.Set(ctx, "verify:"+token, userID.String(), 24*time.Hour)
	return token
}

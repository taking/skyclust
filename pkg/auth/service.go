package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"cmp/pkg/database"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// User represents a user in the system
type User = database.User

// Service defines the authentication service interface
type Service interface {
	Register(ctx context.Context, email, password, name string) (*User, error)
	Login(ctx context.Context, email, password string) (string, *User, error)
	ValidateToken(ctx context.Context, token string) (string, error)
	GetUser(ctx context.Context, userID string) (*User, error)
	HashPassword(password string) (string, error)
	GenerateToken(userID string) (string, error)
}

type service struct {
	db     database.Service
	secret string
}

// NewService creates a new authentication service
func NewService(db database.Service, secret string) Service {
	return &service{db: db, secret: secret}
}

// Register creates a new user
func (s *service) Register(ctx context.Context, email, password, name string) (*User, error) {
	// Check if user already exists
	existingUser, _ := s.db.GetUserByEmail(ctx, email)
	if existingUser != nil {
		return nil, fmt.Errorf("user already exists")
	}

	// Hash password
	hashedPassword, err := s.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &User{
		ID:        generateID(),
		Email:     email,
		Name:      name,
		Password:  hashedPassword,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.db.CreateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// Login authenticates a user
func (s *service) Login(ctx context.Context, email, password string) (string, *User, error) {
	// Get user by email
	user, err := s.db.GetUserByEmail(ctx, email)
	if err != nil {
		return "", nil, fmt.Errorf("user not found")
	}
	if user == nil {
		return "", nil, fmt.Errorf("user not found")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", nil, fmt.Errorf("invalid credentials")
	}

	// Generate token
	token, err := s.GenerateToken(user.ID)
	if err != nil {
		return "", nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return token, user, nil
}

// ValidateToken validates a JWT token
func (s *service) ValidateToken(ctx context.Context, token string) (string, error) {
	// Parse and validate JWT token
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.secret), nil
	})

	if err != nil {
		return "", fmt.Errorf("invalid token: %w", err)
	}

	if !parsedToken.Valid {
		return "", fmt.Errorf("invalid token")
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("invalid token claims")
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return "", fmt.Errorf("invalid user ID in token")
	}

	// Verify token exists in database
	dbUserID, err := s.db.ValidateToken(ctx, token)
	if err != nil || dbUserID != userID {
		return "", fmt.Errorf("token not found in database")
	}

	return userID, nil
}

// GetUser retrieves a user by ID
func (s *service) GetUser(ctx context.Context, userID string) (*User, error) {
	user, err := s.db.GetUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	return user, nil
}

// HashPassword hashes a password using bcrypt
func (s *service) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// GenerateToken generates a JWT token for a user
func (s *service) GenerateToken(userID string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.secret))
	if err != nil {
		return "", err
	}

	// Store token in database for validation
	if err := s.db.StoreToken(context.Background(), userID, tokenString); err != nil {
		return "", err
	}

	return tokenString, nil
}

// generateID generates a random ID
func generateID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

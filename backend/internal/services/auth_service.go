package services

import (
	"context"
	"fmt"
	"time"

	"github.com/fiducia/backend/internal/config"
	"github.com/fiducia/backend/internal/database"
	"github.com/fiducia/backend/internal/models"
	"github.com/fiducia/backend/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	db       *database.DB
	userRepo *repository.UserRepository
	cfg      *config.Config
}

func NewAuthService(db *database.DB, cfg *config.Config) *AuthService {
	return &AuthService{
		db:       db,
		userRepo: repository.NewUserRepository(db),
		cfg:      cfg,
	}
}

func (s *AuthService) Register(ctx context.Context, req models.RegisterRequest) (*models.LoginResponse, error) {
	// Check if user exists
	_, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err == nil {
		return nil, fmt.Errorf("email already registered")
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Start transaction
	tx, err := s.db.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Create Cabinet
	cabinetID := uuid.New()
	_, err = tx.Exec(ctx, "INSERT INTO cabinets (id, name, created_at, updated_at, settings) VALUES ($1, $2, $3, $4, '{}')",
		cabinetID, req.CabinetName, time.Now(), time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to create cabinet: %w", err)
	}

	// Create User
	userID := uuid.New()
	user := &models.User{
		ID:           userID,
		CabinetID:    cabinetID,
		Email:        req.Email,
		PasswordHash: string(hash),
		FullName:     req.FullName,
		Role:         models.RoleAdmin,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Use repo create logic but adapted for Tx?
	// To keep it simple, I'll direct execute query here since Repo uses Pool directly usually.
	// Or I should make repo accept DB interface.
	// For speed, just Exec the user insert here in the TX.
	_, err = tx.Exec(ctx, `
		INSERT INTO users (id, cabinet_id, email, password_hash, full_name, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, user.ID, user.CabinetID, user.Email, user.PasswordHash, user.FullName, user.Role, user.CreatedAt, user.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Generate Token
	token, err := s.GenerateToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &models.LoginResponse{
		Token: token,
		User:  user,
	}, nil
}

func (s *AuthService) Login(ctx context.Context, req models.LoginRequest) (*models.LoginResponse, error) {
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	token, err := s.GenerateToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &models.LoginResponse{
		Token: token,
		User:  user,
	}, nil
}

func (s *AuthService) GenerateToken(user *models.User) (string, error) {
	claims := models.AuthClaims{
		UserID:    user.ID,
		CabinetID: user.CabinetID,
		Role:      user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "fiducia-backend",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWTSecret))
}

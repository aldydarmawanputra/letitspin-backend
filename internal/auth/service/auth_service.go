package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	jwtpkg "let-it-spin/internal/auth/jwt"
	"let-it-spin/internal/auth/repository"
)

type AuthService struct {
	authRepo *repository.AuthRepository
	jwt      *jwtpkg.JWTService
}

func NewAuthService(authRepo *repository.AuthRepository, jwt *jwtpkg.JWTService) *AuthService {
	return &AuthService{
		authRepo: authRepo,
		jwt:      jwt,
	}
}

type LoginResult struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
}

func (s *AuthService) Login(ctx context.Context, email string, password string) (*LoginResult, error) {
	user, err := s.authRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	hash, err := s.authRepo.GetUserCredentialByUserID(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return nil, err
	}

	accessToken, err := s.jwt.GenerateAccessToken(user.ID.String(), user.Email, 15*time.Minute)
	if err != nil {
		return nil, err
	}

	refreshToken := uuid.NewString()

	err = s.authRepo.StoreRefreshToken(
		ctx,
		user.ID,
		refreshToken,
		time.Now().Add(7*24*time.Hour),
	)
	if err != nil {
		return nil, err
	}

	return &LoginResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    900,
	}, nil
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*LoginResult, error) {
	userID, err := s.authRepo.ValidateRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, err
	}

	user, err := s.authRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	accessToken, err := s.jwt.GenerateAccessToken(user.ID.String(), user.Email, 15*time.Minute)
	if err != nil {
		return nil, err
	}

	newRefreshToken := uuid.NewString()

	_ = s.authRepo.RevokeRefreshToken(ctx, refreshToken)

	err = s.authRepo.StoreRefreshToken(
		ctx,
		user.ID,
		newRefreshToken,
		time.Now().Add(7*24*time.Hour),
	)
	if err != nil {
		return nil, err
	}

	return &LoginResult{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    900,
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	if refreshToken == "" {
		return nil
	}
	return s.authRepo.RevokeRefreshToken(ctx, refreshToken)
}

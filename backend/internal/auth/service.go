package auth

import (
	"context"
	"fmt"

	"bitbucket.org/HyperEther/document-management-system/internal/platform/token"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	Register(ctx context.Context, req *RegisterRequest) (*UserResponse, error)
	Login(ctx context.Context, req *LoginRequest) (accessToken, refreshToken string, err error)
	Get(ctx context.Context, id uuid.UUID) (*UserResponse, error)
	RefreshTokens(ctx context.Context, refreshToken string) (accessToken, newRefreshToken string, err error)
}

type service struct {
	repo     Repository
	tokenCfg token.Config
}

func NewService(repo Repository, tokenCfg token.Config) Service {
	return &service{repo: repo, tokenCfg: tokenCfg}
}

func (s *service) Register(ctx context.Context, req *RegisterRequest) (*UserResponse, error) {
	exists, err := s.repo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("checking email existence: %w", err)
	}
	if exists {
		return nil, ErrEmailTaken
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hashing password: %w", err)
	}

	user := FromRegisterRequest(req, string(hash))
	if err := s.repo.Create(ctx, &user); err != nil {
		return nil, fmt.Errorf("creating user: %w", err)
	}

	return ToUserResponse(&user), nil
}

func (s *service) Get(ctx context.Context, id uuid.UUID) (*UserResponse, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return ToUserResponse(user), nil
}

func (s *service) Login(ctx context.Context, req *LoginRequest) (string, string, error) {
	user, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		return "", "", ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return "", "", ErrInvalidCredentials
	}

	return s.issueTokens(user)
}

func (s *service) RefreshTokens(ctx context.Context, refreshToken string) (string, string, error) {
	claims, err := token.ParseRefreshToken(refreshToken, s.tokenCfg.RefreshSecret)
	if err != nil {
		return "", "", ErrInvalidToken
	}

	user, err := s.repo.GetByID(ctx, claims.UserID)
	if err != nil {
		return "", "", ErrInvalidToken
	}

	return s.issueTokens(user)
}

func (s *service) issueTokens(user *User) (string, string, error) {
	accessToken, err := token.NewAccessToken(user.ID, user.Email, s.tokenCfg.AccessSecret)
	if err != nil {
		return "", "", fmt.Errorf("issue access token: %w", err)
	}

	refreshToken, err := token.NewRefreshToken(user.ID, s.tokenCfg.RefreshSecret)
	if err != nil {
		return "", "", fmt.Errorf("issue refresh token: %w", err)
	}

	return accessToken, refreshToken, nil
}

package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"bitbucket.org/HyperEther/document-management-system/internal/platform/token"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var testTokenCfg = token.Config{
	AccessSecret:  "test-access-secret",
	RefreshSecret: "test-refresh-secret",
}

func TestService_Register_Success(t *testing.T) {
	repo := &mockRepository{
		existsByEmailFn: func(ctx context.Context, email string) (bool, error) {
			return false, nil
		},
		createFn: func(ctx context.Context, user *User) error {
			return nil
		},
	}

	svc := NewService(repo, testTokenCfg)
	resp, err := svc.Register(context.Background(), &RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.Email != "test@example.com" {
		t.Errorf("expected test@example.com, got %s", resp.Email)
	}
	if resp.ID == (uuid.UUID{}) {
		t.Error("expected a non-zero UUID")
	}
}

func TestService_Register_EmailTaken(t *testing.T) {
	repo := &mockRepository{
		existsByEmailFn: func(ctx context.Context, email string) (bool, error) {
			return true, nil
		},
	}

	svc := NewService(repo, testTokenCfg)
	_, err := svc.Register(context.Background(), &RegisterRequest{
		Email:    "taken@example.com",
		Password: "password123",
	})

	if !errors.Is(err, ErrEmailTaken) {
		t.Errorf("expected ErrEmailTaken, got %v", err)
	}
}

func TestService_Register_PasswordIsHashed(t *testing.T) {
	var storedUser *User
	repo := &mockRepository{
		existsByEmailFn: func(ctx context.Context, email string) (bool, error) {
			return false, nil
		},
		createFn: func(ctx context.Context, user *User) error {
			storedUser = user
			return nil
		},
	}

	svc := NewService(repo, testTokenCfg)
	svc.Register(context.Background(), &RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
	})

	if storedUser.Password == "password123" {
		t.Error("password was stored in plaintext")
	}
	if len(storedUser.Password) == 0 {
		t.Error("hashed password is empty")
	}
}

func TestService_Register_ExistsByEmailFailure(t *testing.T) {
	repo := &mockRepository{
		existsByEmailFn: func(ctx context.Context, email string) (bool, error) {
			return false, errors.New("db connection lost")
		},
	}

	_, err := NewService(repo, testTokenCfg).Register(context.Background(), &RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
	})

	if err == nil {
		t.Error("expected an error, got nil")
	}
}

func TestService_Register_CreateFailure(t *testing.T) {
	repo := &mockRepository{
		existsByEmailFn: func(ctx context.Context, email string) (bool, error) {
			return false, nil
		},
		createFn: func(ctx context.Context, user *User) error {
			return errors.New("insert failed")
		},
	}

	_, err := NewService(repo, testTokenCfg).Register(context.Background(), &RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
	})

	if err == nil {
		t.Error("expected an error, got nil")
	}
}

func TestService_Login_Success(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	repo := &mockRepository{
		getByEmailFn: func(ctx context.Context, email string) (*User, error) {
			return &User{
				ID:       uuid.New(),
				Email:    email,
				Password: string(hash),
			}, nil
		},
	}

	svc := NewService(repo, testTokenCfg)
	accessToken, refreshToken, err := svc.Login(context.Background(), &LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if accessToken == "" {
		t.Error("expected non-empty access token")
	}
	if refreshToken == "" {
		t.Error("expected non-empty refresh token")
	}
}

func TestService_Login_WrongPassword(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	repo := &mockRepository{
		getByEmailFn: func(ctx context.Context, email string) (*User, error) {
			return &User{
				ID:       uuid.New(),
				Email:    email,
				Password: string(hash),
			}, nil
		},
	}

	_, _, err := NewService(repo, testTokenCfg).Login(context.Background(), &LoginRequest{
		Email:    "test@example.com",
		Password: "wrongpassword",
	})

	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestService_Login_UserNotFound(t *testing.T) {
	repo := &mockRepository{
		getByEmailFn: func(ctx context.Context, email string) (*User, error) {
			return nil, ErrNotFound
		},
	}

	_, _, err := NewService(repo, testTokenCfg).Login(context.Background(), &LoginRequest{
		Email:    "ghost@example.com",
		Password: "password123",
	})

	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestService_RefreshTokens_Success(t *testing.T) {
	userID := uuid.New()
	repo := &mockRepository{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*User, error) {
			return &User{ID: userID, Email: "test@example.com"}, nil
		},
	}

	refreshToken, err := token.NewRefreshToken(userID, testTokenCfg.RefreshSecret)
	if err != nil {
		t.Fatalf("could not create refresh token: %v", err)
	}

	svc := NewService(repo, testTokenCfg)
	accessToken, newRefreshToken, err := svc.RefreshTokens(context.Background(), refreshToken)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if accessToken == "" {
		t.Error("expected non-empty access token")
	}
	if newRefreshToken == "" {
		t.Error("expected non-empty refresh token")
	}
	if newRefreshToken == refreshToken {
		t.Error("expected a new refresh token, got the same one")
	}
}

func TestService_RefreshTokens_InvalidToken(t *testing.T) {
	repo := &mockRepository{}

	_, _, err := NewService(repo, testTokenCfg).RefreshTokens(
		context.Background(), "not.a.valid.token",
	)

	if !errors.Is(err, ErrInvalidToken) {
		t.Errorf("expected ErrInvalidToken, got %v", err)
	}
}

func TestService_RefreshTokens_UserNotFound(t *testing.T) {
	userID := uuid.New()
	repo := &mockRepository{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*User, error) {
			return nil, ErrNotFound
		},
	}

	refreshToken, _ := token.NewRefreshToken(userID, testTokenCfg.RefreshSecret)

	_, _, err := NewService(repo, testTokenCfg).RefreshTokens(context.Background(), refreshToken)

	// user deleted between token issue and refresh — still ErrInvalidToken to the caller
	if !errors.Is(err, ErrInvalidToken) {
		t.Errorf("expected ErrInvalidToken, got %v", err)
	}
}

func TestService_RefreshTokens_ExpiredToken(t *testing.T) {
	repo := &mockRepository{}

	expiredToken, err := token.NewRefreshTokenWithDuration(
		uuid.New(), testTokenCfg.RefreshSecret, -time.Hour,
	)
	if err != nil {
		t.Fatalf("could not create expired token: %v", err)
	}

	_, _, err = NewService(repo, testTokenCfg).RefreshTokens(context.Background(), expiredToken)

	if !errors.Is(err, ErrInvalidToken) {
		t.Errorf("expected ErrInvalidToken, got %v", err)
	}
}

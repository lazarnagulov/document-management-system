package auth

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestService_Register_Success(t *testing.T) {
	repo := &mockRepository{
		existsByEmailFn: func(ctx context.Context, email string) (bool, error) {
			return false, nil
		},
		createFn: func(ctx context.Context, user *User) error {
			return nil
		},
	}

	svc := NewService(repo)
	resp, err := svc.Register(context.Background(), &RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.Email != "test@example.com" {
		t.Errorf("expected email test@example.com, got %s", resp.Email)
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

	svc := NewService(repo)
	_, err := svc.Register(context.Background(), &RegisterRequest{
		Email:    "taken@example.com",
		Password: "password123",
	})

	if !errors.Is(err, ErrEmailTaken) {
		t.Errorf("expected ErrEmailTaken, got %v", err)
	}
}

func TestService_Register_ExistsByEmailFailure(t *testing.T) {
	repo := &mockRepository{
		existsByEmailFn: func(ctx context.Context, email string) (bool, error) {
			return false, errors.New("db connection lost")
		},
	}

	svc := NewService(repo)
	_, err := svc.Register(context.Background(), &RegisterRequest{
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

	svc := NewService(repo)
	_, err := svc.Register(context.Background(), &RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
	})

	if err == nil {
		t.Error("expected an error, got nil")
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

	svc := NewService(repo)
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

package auth

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type mockService struct {
	registerFn      func(ctx context.Context, req *RegisterRequest) (*UserResponse, error)
	loginFn         func(ctx context.Context, req *LoginRequest) (string, string, error)
	refreshTokensFn func(ctx context.Context, refreshToken string) (string, string, error)
	getFn           func(ctx context.Context, id uuid.UUID) (*UserResponse, error)
}

func (m *mockService) Register(ctx context.Context, req *RegisterRequest) (*UserResponse, error) {
	if m.registerFn == nil {
		return nil, fmt.Errorf("registerFn not set")
	}
	return m.registerFn(ctx, req)
}

func (m *mockService) Login(ctx context.Context, req *LoginRequest) (string, string, error) {
	if m.loginFn == nil {
		return "", "", fmt.Errorf("loginFn not set")
	}
	return m.loginFn(ctx, req)
}

func (m *mockService) RefreshTokens(ctx context.Context, refreshToken string) (string, string, error) {
	if m.refreshTokensFn == nil {
		return "", "", fmt.Errorf("refreshTokensFn not set")
	}
	return m.refreshTokensFn(ctx, refreshToken)
}

func (m *mockService) Get(ctx context.Context, id uuid.UUID) (*UserResponse, error) {
	if m.getFn == nil {
		return nil, fmt.Errorf("getFn not set")
	}
	return m.getFn(ctx, id)
}

type mockRepository struct {
	createFn        func(ctx context.Context, user *User) error
	getByEmailFn    func(ctx context.Context, email string) (*User, error)
	getByIDFn       func(ctx context.Context, id uuid.UUID) (*User, error)
	existsByEmailFn func(ctx context.Context, email string) (bool, error)
}

func (m *mockRepository) Create(ctx context.Context, user *User) error {
	if m.createFn == nil {
		return fmt.Errorf("createFn not set")
	}
	return m.createFn(ctx, user)
}

func (m *mockRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	if m.getByEmailFn == nil {
		return nil, fmt.Errorf("getByEmailFn not set")
	}
	return m.getByEmailFn(ctx, email)
}

func (m *mockRepository) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	if m.getByIDFn == nil {
		return nil, fmt.Errorf("getByIDFn not set")
	}
	return m.getByIDFn(ctx, id)
}

func (m *mockRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	if m.existsByEmailFn == nil {
		return false, fmt.Errorf("existsByEmailFn not set")
	}
	return m.existsByEmailFn(ctx, email)
}

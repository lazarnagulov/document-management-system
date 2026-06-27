package auth

import (
	"context"
)

type mockService struct {
	registerFn func(ctx context.Context, req *RegisterRequest) (*UserResponse, error)
}

func (m *mockService) Register(ctx context.Context, req *RegisterRequest) (*UserResponse, error) {
	return m.registerFn(ctx, req)
}

type mockRepository struct {
	createFn        func(ctx context.Context, user *User) error
	getByEmailFn    func(ctx context.Context, email string) (*User, error)
	existsByEmailFn func(ctx context.Context, email string) (bool, error)
}

func (m *mockRepository) Create(ctx context.Context, user *User) error {
	return m.createFn(ctx, user)
}

func (m *mockRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	return m.getByEmailFn(ctx, email)
}

func (m *mockRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	return m.existsByEmailFn(ctx, email)
}

package mocks

import (
	"context"
	"time"

	"github.com/Kostaaa1/tinylink/core/transactor"
	"github.com/Kostaaa1/tinylink/internal/domain/tinylink"
	"github.com/stretchr/testify/mock"
)

type MockDbRepository struct {
	mock.Mock
}

type MockCacheRepository struct {
	mock.Mock
}

var _ tinylink.DbRepository = (*MockDbRepository)(nil)

func (m *MockDbRepository) Insert(ctx context.Context, tl *tinylink.Tinylink) error {
	args := m.Called(ctx, tl)
	return args.Error(0)
}

func (m *MockDbRepository) Update(ctx context.Context, tl *tinylink.Tinylink) error {
	args := m.Called(ctx, tl)
	return args.Error(0)
}

func (m *MockDbRepository) Delete(ctx context.Context, userID uint64, alias string) error {
	args := m.Called(ctx, userID, alias)
	return args.Error(0)
}

func (m *MockDbRepository) Redirect(ctx context.Context, userID *uint64, alias string) (*tinylink.RedirectValue, error) {
	args := m.Called(ctx, userID, alias)

	if rv := args.Get(0); rv != nil {
		return rv.(*tinylink.RedirectValue), args.Error(1)
	}

	return nil, args.Error(1)
}

func (m *MockDbRepository) Get(ctx context.Context, rowID uint64) (*tinylink.Tinylink, error) {
	args := m.Called(ctx, rowID)

	if rv := args.Get(0); rv != nil {
		return rv.(*tinylink.Tinylink), args.Error(1)
	}

	return nil, args.Error(1)
}

func (m *MockDbRepository) ListByUserID(ctx context.Context, userID uint64) ([]*tinylink.Tinylink, error) {
	args := m.Called(ctx, userID)
	if rv := args.Get(0); rv != nil {
		return rv.([]*tinylink.Tinylink), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockDbRepository) ListByGuestUUID(ctx context.Context, guestUUID string) ([]*tinylink.Tinylink, error) {
	args := m.Called(ctx, guestUUID)
	if rv := args.Get(0); rv != nil {
		return rv.([]*tinylink.Tinylink), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockDbRepository) WithTx(tx transactor.Tx) tinylink.DbRepository {
	return m
}

var _ tinylink.CacheRepository = (*MockCacheRepository)(nil)

func (m *MockCacheRepository) Redirect(ctx context.Context, alias string) (*tinylink.RedirectValue, error) {
	args := m.Called(ctx, alias)

	if rv := args.Get(0); rv != nil {
		return rv.(*tinylink.RedirectValue), args.Error(1)
	}

	return nil, args.Error(1)
}

func (m *MockCacheRepository) Cache(ctx context.Context, value tinylink.RedirectValue, ttl time.Duration) error {
	args := m.Called(ctx, value, ttl)
	return args.Error(0)
}

func (m *MockCacheRepository) GenerateAlias(ctx context.Context) (string, error) {
	args := m.Called(ctx)
	if rv := args.Get(0); rv != nil {
		return rv.(string), args.Error(1)
	}
	return "", args.Error(1)
}

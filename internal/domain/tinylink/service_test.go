package tinylink_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Kostaaa1/tinylink/internal/constants"
	"github.com/Kostaaa1/tinylink/internal/domain/tinylink"
	mocks "github.com/Kostaaa1/tinylink/internal/mocks/tinylink"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// fuck yea
func TestTinylinkService_Redirect(t *testing.T) {
	userID := (*uint64)(nil)

	alias := "abc123"

	expected := &tinylink.RedirectValue{
		RowID: 42,
		URL:   "https://test_url.com",
		Alias: alias,
	}

	unknownErr := errors.New("some random unknown error from cache repo")

	defCacheTTL := time.Hour

	defRedirectValue := tinylink.RedirectValue{
		RowID: expected.RowID,
		Alias: expected.Alias,
		URL:   expected.URL,
	}

	type testCase struct {
		dbRedirectReturn    []interface{}
		cacheRedirectReturn []interface{}
		cacheCacheReturn    []interface{}
		assertFn            func(t *testing.T, ctx context.Context, rowID uint64, url string, err error)
		mockAssertions      func(t *testing.T, ctx context.Context, mdr *mocks.MockDbRepository, mcr *mocks.MockCacheRepository)
	}

	testCases := map[string]testCase{
		"redirect value from cache": {
			cacheRedirectReturn: []interface{}{expected, nil},
			assertFn: func(t *testing.T, ctx context.Context, rowID uint64, url string, err error) {
				require.NoError(t, err)
				require.Equal(t, expected.RowID, rowID)
				require.Equal(t, expected.URL, url)
			},
			mockAssertions: func(t *testing.T, ctx context.Context, mdr *mocks.MockDbRepository, mcr *mocks.MockCacheRepository) {
				mcr.AssertCalled(t, "Redirect", ctx, alias)
				mcr.AssertExpectations(t)
				mdr.AssertNotCalled(t, "Redirect")
				mcr.AssertNotCalled(t, "Cache")
			},
		},
		"redirect value from db and cache it": {
			cacheRedirectReturn: []interface{}{nil, constants.ErrNotFound},
			dbRedirectReturn:    []interface{}{expected, nil},
			cacheCacheReturn:    []interface{}{nil},
			assertFn: func(t *testing.T, ctx context.Context, rowID uint64, url string, err error) {
				require.NoError(t, err)
				require.Equal(t, expected.RowID, rowID)
				require.Equal(t, expected.URL, url)
			},
			mockAssertions: func(t *testing.T, ctx context.Context, mdr *mocks.MockDbRepository, mcr *mocks.MockCacheRepository) {
				mcr.AssertCalled(t, "Redirect", ctx, alias)
				mcr.AssertExpectations(t)

				mdr.AssertCalled(t, "Redirect", ctx, userID, alias)
				mdr.AssertExpectations(t)

				mcr.AssertCalled(t, "Cache", ctx, defRedirectValue, mock.Anything)
			},
		},
		"cache repo returning unknown error": {
			cacheRedirectReturn: []interface{}{nil, unknownErr},
			assertFn: func(t *testing.T, ctx context.Context, rowID uint64, url string, err error) {
				require.Error(t, err)
				require.Equal(t, err, unknownErr)
				require.Equal(t, rowID, uint64(0))
				require.Equal(t, url, "")
			},
			mockAssertions: func(t *testing.T, ctx context.Context, mdr *mocks.MockDbRepository, mcr *mocks.MockCacheRepository) {
				mcr.AssertCalled(t, "Redirect", ctx, alias)
				mdr.AssertNotCalled(t, "Redirect")
				mcr.AssertNotCalled(t, "Cache")
				mcr.AssertExpectations(t)
			},
		},
		"URL not found by alias": {
			cacheRedirectReturn: []interface{}{nil, constants.ErrNotFound},
			dbRedirectReturn:    []interface{}{nil, constants.ErrNotFound},
			assertFn: func(t *testing.T, ctx context.Context, rowID uint64, url string, err error) {
				require.Error(t, err)
				require.Equal(t, err, constants.ErrNotFound)
				require.Equal(t, rowID, uint64(0))
				require.Equal(t, url, "")
			},
			mockAssertions: func(t *testing.T, ctx context.Context, mdr *mocks.MockDbRepository, mcr *mocks.MockCacheRepository) {
				mcr.AssertCalled(t, "Redirect", ctx, alias)
				mdr.AssertCalled(t, "Redirect", ctx, userID, alias)
				mcr.AssertNotCalled(t, "Cache")

				mdr.AssertExpectations(t)
				mcr.AssertExpectations(t)
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()

			mockDb := new(mocks.MockDbRepository)
			mockCache := new(mocks.MockCacheRepository)
			svc := tinylink.NewService(mockDb, mockCache)

			if tc.cacheRedirectReturn != nil {
				mockCache.On("Redirect", ctx, expected.Alias).Return(tc.cacheRedirectReturn...)
			}

			if tc.cacheCacheReturn != nil {
				mockCache.On("Cache", ctx, defRedirectValue, defCacheTTL).Return(tc.cacheRedirectReturn...)
			}

			if tc.dbRedirectReturn != nil {
				mockDb.On("Redirect", ctx, userID, expected.Alias).Return(tc.dbRedirectReturn...)
			}

			rowID, url, err := svc.Redirect(ctx, userID, expected.Alias)
			tc.assertFn(t, ctx, rowID, url, err)
			tc.mockAssertions(t, ctx, mockDb, mockCache)
		})
	}
}

// func TestTinylinkService_Create(t *testing.T) {
// 	type testCase struct {
// 		params          tinylink.CreateTinylinkParams
// 		mockDbReturn    []interface{}
// 		mockCacheReturn []interface{}
// 		assertFn        func(t *testing.T, tl *tinylink.Tinylink, err error)
// 		mockAssertions  func(t *testing.T, ctx context.Context, mdr *mocks.MockDbRepository, mcr *mocks.MockCacheRepository)
// 	}

// 	alias := "abc123"
// 	url := "https://test_123.com"
// 	guestUUID := "random_string"

// 	testCases := map[string]testCase{
// 		"pass: create tinylink with alias": {
// 			params: tinylink.CreateTinylinkParams{
// 				Alias:     &alias,
// 				URL:       url,
// 				GuestUUID: guestUUID,
// 			},
// 			mockCacheReturn: nil,
// 			mockDbReturn:    []interface{}{nil},
// 			assertFn: func(t *testing.T, tl *tinylink.Tinylink, err error) {
// 				require.NoError(t, err)
// 				require.NotZero(t, tl.ID)
// 				require.NotZero(t, tl.Version)
// 				require.False(t, tl.CreatedAt.IsZero())
// 				require.Equal(t, tl.Alias, alias)
// 				require.Equal(t, tl.URL, url)
// 				require.Equal(t, tl.GuestUUID, guestUUID)
// 			},
// 			mockAssertions: func(t *testing.T, ctx context.Context, mdr *mocks.MockDbRepository, mcr *mocks.MockCacheRepository,
// 			) {
// 				mdr.AssertCalled(t, "Insert", ctx, mock.AnythingOfType("*tinylink.Tinylink"))
// 				mdr.AssertNumberOfCalls(t, "Insert", 1)
// 				mcr.AssertNotCalled(t, "GenerateAlias", ctx)
// 			},
// 		},
// 		"pass: create tinylink without alias": {
// 			params: tinylink.CreateTinylinkParams{
// 				Alias:     nil,
// 				URL:       url,
// 				GuestUUID: guestUUID,
// 			},
// 			mockCacheReturn: []interface{}{"generated_alias", nil},
// 			mockDbReturn:    []interface{}{nil},
// 			assertFn: func(t *testing.T, tl *tinylink.Tinylink, err error) {
// 				require.NoError(t, err)
// 				require.NotZero(t, tl.ID)
// 				require.NotZero(t, tl.Version)
// 				require.False(t, tl.CreatedAt.IsZero())
// 				require.Equal(t, tl.Alias, "generated_alias")
// 				require.Equal(t, tl.URL, url)
// 				require.Equal(t, tl.GuestUUID, guestUUID)
// 			},
// 			mockAssertions: func(t *testing.T, ctx context.Context, mdr *mocks.MockDbRepository, mcr *mocks.MockCacheRepository,
// 			) {
// 				mdr.AssertCalled(t, "Insert", ctx, mock.AnythingOfType("*tinylink.Tinylink"))
// 				mdr.AssertExpectations(t)
// 				mcr.AssertCalled(t, "GenerateAlias", ctx)
// 				mcr.AssertExpectations(t)
// 			},
// 		},
// 		"fail: create private tinylink as guest": {
// 			params: tinylink.CreateTinylinkParams{
// 				Alias:     nil,
// 				URL:       url,
// 				GuestUUID: guestUUID,
// 				Private:   true,
// 			},
// 			mockDbReturn:    nil,
// 			mockCacheReturn: nil,
// 			assertFn: func(t *testing.T, tl *tinylink.Tinylink, err error) {
// 				require.Error(t, err)
// 				require.Nil(t, tl)
// 			},
// 			mockAssertions: func(t *testing.T, ctx context.Context, mdr *mocks.MockDbRepository, mcr *mocks.MockCacheRepository,
// 			) {
// 				mdr.AssertNotCalled(t, "Insert", ctx, mock.AnythingOfType("*tinylink.Tinylink"))
// 				mdr.AssertNotCalled(t, "GenerateAlias", ctx)
// 			},
// 		},
// 		"fail: db insert failed": {
// 			params: tinylink.CreateTinylinkParams{
// 				Alias:     &alias,
// 				URL:       url,
// 				GuestUUID: guestUUID,
// 			},
// 			mockDbReturn:    []interface{}{errors.New("some db error")},
// 			mockCacheReturn: nil,
// 			assertFn: func(t *testing.T, tl *tinylink.Tinylink, err error) {
// 				require.Error(t, err)
// 				require.Nil(t, tl)
// 				require.Equal(t, err, errors.New("some db error"))
// 			},
// 			mockAssertions: func(t *testing.T, ctx context.Context, mdr *mocks.MockDbRepository, mcr *mocks.MockCacheRepository,
// 			) {
// 				mdr.AssertCalled(t, "Insert", ctx, mock.AnythingOfType("*tinylink.Tinylink"))
// 				mdr.AssertExpectations(t)
// 				mcr.AssertNotCalled(t, "GenerateAlias", ctx)
// 			},
// 		},
// 		"fail: guestUUID missing": {
// 			params: tinylink.CreateTinylinkParams{
// 				Alias: nil,
// 				URL:   url,
// 			},
// 			mockDbReturn:    nil,
// 			mockCacheReturn: nil,
// 			assertFn: func(t *testing.T, tl *tinylink.Tinylink, err error) {
// 				require.Error(t, err)
// 				require.Nil(t, tl)
// 			},
// 			mockAssertions: func(t *testing.T, ctx context.Context, mdr *mocks.MockDbRepository, mcr *mocks.MockCacheRepository,
// 			) {
// 				mdr.AssertNotCalled(t, "Insert", ctx, mock.AnythingOfType("*tinylink.Tinylink"))
// 				mcr.AssertNotCalled(t, "GenerateAlias", ctx)
// 			},
// 		},
// 	}

// 	for name, tc := range testCases {
// 		t.Run(name, func(t *testing.T) {
// 			ctx := context.Background()
// 			mockDb := new(mocks.MockDbRepository)
// 			mockCache := new(mocks.MockCacheRepository)

// 			svc := tinylink.NewService(mockDb, mockCache)

// 			if tc.mockCacheReturn != nil {
// 				mockCache.On("GenerateAlias", ctx).Return(tc.mockCacheReturn...)
// 			}

// 			if tc.mockDbReturn != nil {
// 				mockDb.On("Insert", ctx, mock.AnythingOfType("*tinylink.Tinylink")).
// 					Run(func(args mock.Arguments) {
// 						tl := args.Get(1).(*tinylink.Tinylink)
// 						tl.ID = 42
// 						tl.Version++
// 						tl.CreatedAt = time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
// 					}).
// 					Return(tc.mockDbReturn...)
// 			}

// 			tl, err := svc.Create(ctx, tc.params)
// 			tc.assertFn(t, tl, err)
// 			tc.mockAssertions(t, ctx, mockDb, mockCache)
// 		})
// 	}
// }

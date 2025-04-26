package tinylink_test

import (
	"context"
	"fmt"
	"math/rand"
	"testing"

	"github.com/Kostaaa1/tinylink/internal/common/data"
	"github.com/Kostaaa1/tinylink/internal/domain/tinylink"
	"github.com/Kostaaa1/tinylink/internal/domain/user"
	"github.com/Kostaaa1/tinylink/internal/infrastructure/db/redisdb"
	"github.com/Kostaaa1/tinylink/internal/infrastructure/db/sqlitedb"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

var (
	ctx = context.TODO()
)

func randomURL() string {
	return fmt.Sprintf("https://%s.com", uuid.NewString())
}

func randomAlias() string {
	return uuid.NewString()
}

func setupTinylinkSuite(t *testing.T) (tinylink.Adapters, *tinylink.Service, user.Adapters, *user.Service, string) {
	redis := redisdb.StartTest(t)
	db := sqlitedb.StartTest(t)

	provider := tinylink.NewRepositoryProvider(db, redis)
	adapters := provider.Adapters()
	service := tinylink.NewService(provider)

	userProvider := user.NewRepositoryProvider(db)
	userAdapters := userProvider.Adapters()
	userService := user.NewService(userProvider, nil)

	return adapters, service, userAdapters, userService, uuid.NewString()
}

func createMockUser(t *testing.T, ctx context.Context, userDb user.UserRepository) *user.User {
	mockUser := &user.User{
		Email: fmt.Sprintf("testuser%d@gmail.com", rand.Intn(100)),
		Name:  "TestUser",
	}
	mockUser.Password.Set("test123")
	require.Nil(t, userDb.Insert(ctx, mockUser))
	return mockUser
}

func mockTinylinkCreateRequest() tinylink.CreateTinylinkRequest {
	url := randomURL()
	alias := randomAlias()
	req := tinylink.CreateTinylinkRequest{
		URL:   url,
		Alias: alias,
	}
	return req
}

func TestTInylinkService_List(t *testing.T) {
	t.Parallel()

	adapters, service, userAdapters, _, sessionID := setupTinylinkSuite(t)

	// Store in redis
	req := mockTinylinkCreateRequest()
	tl, err := service.Create(ctx, nil, &sessionID, req)
	require.NoError(t, err)
	require.Equal(t, tl.ID, uint64(0))

	req = mockTinylinkCreateRequest()
	tl, err = service.Create(ctx, nil, &sessionID, req)
	require.NoError(t, err)
	require.Equal(t, tl.ID, uint64(0))

	tls, err := adapters.TinylinkRedisRepository.ListUserLinks(ctx, sessionID)
	require.NoError(t, err)
	require.Equal(t, len(tls), 2)

	// Store in sqlite
	mockUser := createMockUser(t, ctx, userAdapters.UserDbRepository)
	userID := mockUser.GetID()

	tl, err = service.Create(ctx, &userID, nil, mockTinylinkCreateRequest())
	require.NoError(t, err)
	require.NotNil(t, tl)
	require.Greater(t, tl.ID, uint64(0))
	require.Greater(t, tl.CreatedAt, int64(0))
	require.Greater(t, tl.CreatedAt, int64(0))

	tl2, err := service.Create(ctx, &userID, nil, mockTinylinkCreateRequest())
	require.NoError(t, err)
	require.NotNil(t, tl2)
	require.Greater(t, tl2.ID, uint64(0))
	require.Greater(t, tl2.CreatedAt, int64(0))
	require.Greater(t, tl2.CreatedAt, int64(0))

	tls, err = adapters.TinylinkDBRepository.ListUserLinks(ctx, userID)
	require.NoError(t, err)
	require.Greater(t, len(tls), 0)
	require.Equal(t, len(tls), 2)
	require.Greater(t, tls[0].ID, uint64(0))
	require.Greater(t, tls[1].ID, uint64(0))
}

func TestTinylinkService_CreatePublic(t *testing.T) {
	t.Parallel()

	adapters, service, userAdapters, _, sessionID := setupTinylinkSuite(t)
	mockUser := createMockUser(t, ctx, userAdapters.UserDbRepository)
	url := randomURL()
	alias := randomAlias()
	req := tinylink.CreateTinylinkRequest{
		URL:   url,
		Alias: alias,
	}

	userID := mockUser.GetID()

	// should insert in peristed db
	tl, err := service.Create(ctx, &userID, nil, req)
	require.NoError(t, err)
	require.NotNil(t, tl)
	require.Greater(t, tl.ID, uint64(0))
	require.Greater(t, tl.CreatedAt, int64(0))
	require.Greater(t, tl.CreatedAt, int64(0))

	ok, err := adapters.TinylinkDBRepository.AliasExistsWithID(ctx, mockUser.GetID(), alias)
	require.NoError(t, err)
	require.True(t, ok)

	// should insert in redis
	req2 := tinylink.CreateTinylinkRequest{
		URL:   randomURL(),
		Alias: randomAlias(),
	}
	tl2, err := service.Create(ctx, nil, &sessionID, req2)
	require.NoError(t, err)
	require.NotNil(t, tl2)
	require.Greater(t, tl.CreatedAt, int64(0))

	ok, err = adapters.TinylinkRedisRepository.AliasExists(ctx, req2.Alias)
	require.Nil(t, err)
	require.True(t, ok)
}

func TestTinylinkService_MigrateLinksFromRedisToDB(t *testing.T) {
	t.Parallel()

	adapters, service, userAdapters, _, sessionID := setupTinylinkSuite(t)
	mockUser := createMockUser(t, ctx, userAdapters.UserDbRepository)
	userID := mockUser.GetID()

	// create couple of links in redis
	tl1, err := service.Create(ctx, nil, &sessionID, tinylink.CreateTinylinkRequest{
		URL:   randomURL(),
		Alias: randomAlias(),
	})
	require.NoError(t, err)
	require.NotNil(t, tl1)
	require.Equal(t, tl1.ID, uint64(0))
	t.Log("Created first link: ", tl1.Alias)

	tl2, err := service.Create(ctx, nil, &sessionID, tinylink.CreateTinylinkRequest{
		URL:   randomURL(),
		Alias: randomAlias(),
	})
	require.NoError(t, err)
	require.NotNil(t, tl2)
	require.Equal(t, tl2.ID, uint64(0))
	t.Log("Created second link: ", tl2.Alias)

	// validate
	links, err := adapters.TinylinkRedisRepository.ListUserLinks(ctx, sessionID)
	require.NoError(t, err)
	require.NotNil(t, links)
	require.Equal(t, len(links), 2)

	// migrate - delete from redis and move them to sqlite
	err = service.MigrateLinksFromRedisToDB(ctx, userID, sessionID)
	require.NoError(t, err)

	// validate that redis is empty for this sessionID
	links, err = adapters.TinylinkRedisRepository.ListUserLinks(ctx, sessionID)
	require.Error(t, err)
	require.Equal(t, err, data.ErrNotFound)
	require.Zero(t, len(links))

	// validate that sqlite includes THOSE links
	dbLinks, err := adapters.TinylinkDBRepository.ListUserLinks(ctx, userID)
	require.NoError(t, err)
	require.Equal(t, len(dbLinks), 2)
}

func TestTinylinkService_CreatePrivate(t *testing.T) {
	t.Parallel()

	adapters, service, userAdapters, _, _ := setupTinylinkSuite(t)
	mockUser := createMockUser(t, ctx, userAdapters.UserDbRepository)
	userID := mockUser.GetID()

	req := tinylink.CreateTinylinkRequest{
		URL:     randomURL(),
		Alias:   randomAlias(),
		Private: true,
	}

	tl, err := service.Create(ctx, &userID, nil, req)
	require.NoError(t, err)
	require.NotNil(t, tl)
	require.Greater(t, tl.ID, uint64(0))
	require.Greater(t, tl.CreatedAt, int64(0))

	fetched, err := adapters.TinylinkDBRepository.Get(ctx, tl.ID)
	require.NoError(t, err)
	require.Equal(t, tl.ID, fetched.ID)
	require.Greater(t, tl.CreatedAt, int64(0))
	require.Equal(t, tl.CreatedAt, fetched.CreatedAt)
	require.Equal(t, tl.Alias, fetched.Alias)
	require.Equal(t, tl.URL, fetched.URL)
}

func TestTinylinkService_DuplicateAliasFailsAndSucceeds(t *testing.T) {
	t.Parallel()

	adapters, service, userAdapters, _, sessionID := setupTinylinkSuite(t)
	mockUser := createMockUser(t, ctx, userAdapters.UserDbRepository)
	userID := mockUser.GetID()

	req := tinylink.CreateTinylinkRequest{
		URL:   randomURL(),
		Alias: randomAlias(),
	}

	// should insert public tinylink in peristed db
	tl, err := service.Create(ctx, &userID, nil, req)
	require.NoError(t, err)
	require.NotNil(t, tl)
	require.Greater(t, tl.ID, uint64(0))
	require.Greater(t, tl.CreatedAt, int64(0))
	require.Greater(t, tl.CreatedAt, int64(0))
	require.False(t, tl.Private)

	// validate it in db
	ok, err := adapters.TinylinkDBRepository.AliasExistsWithID(ctx, mockUser.GetID(), tl.Alias)
	require.NoError(t, err)
	require.True(t, ok)

	// should fail becuase it uses the same alias
	tl2, err := service.Create(ctx, nil, &sessionID, req)
	require.Error(t, err)
	require.Nil(t, tl2)
	require.Equal(t, err, tinylink.ErrAliasExists)

	// but if we update user tinylink to be private, then it should become available to create public tinylink with the same alias.
	randURL := randomURL()
	randAlias := randomAlias()
	updatedTl, err := service.Update(ctx, mockUser.GetID(), tinylink.UpdateTinylinkRequest{
		ID:      tl.ID,
		URL:     &randURL,
		Alias:   &randAlias,
		Private: true,
	})
	require.NoError(t, err)
	require.True(t, updatedTl.Private)
	require.Equal(t, updatedTl.ID, tl.ID)
	require.Equal(t, updatedTl.Version, uint64(1))

	// now inserting will work again
	tl, err = service.Create(ctx, nil, &sessionID, req)
	require.NoError(t, err)
	require.NotNil(t, tl)
}

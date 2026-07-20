package handler_test

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/artalkjs/artalk/v2/internal/cache"
	"github.com/artalkjs/artalk/v2/internal/config"
	"github.com/artalkjs/artalk/v2/internal/dao"
	"github.com/artalkjs/artalk/v2/internal/entity"
	"github.com/artalkjs/artalk/v2/server/common"
	"github.com/artalkjs/artalk/v2/server/handler"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthMergeApplyTransaction(t *testing.T) {
	t.Run("commits all merged data", func(t *testing.T) {
		app, fiber := NewApiTestApp()
		defer app.Cleanup()

		target, source, comment, notify, vote, identity := createMergeFixtures(t, app.Dao(), false)
		handler.AuthMergeApply(app.App, fiber)

		resp := mergeUsersRequest(t, fiber, app.Conf().AppKey, source, target.Name)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		assertRecordUserID(t, app.Dao(), &comment, target.ID)
		assertRecordUserID(t, app.Dao(), &notify, target.ID)
		assertRecordUserID(t, app.Dao(), &vote, target.ID)
		assertRecordUserID(t, app.Dao(), &identity, target.ID)

		var sourceCount int64
		require.NoError(t, app.Dao().DB().Unscoped().Model(&entity.User{}).Where("id = ?", source.ID).Count(&sourceCount).Error)
		assert.Zero(t, sourceCount)
	})

	t.Run("rolls back every table and preserves caches", func(t *testing.T) {
		app, fiber := NewApiTestApp()
		defer app.Cleanup()

		target, source, comment, notify, vote, identity := createMergeFixtures(t, app.Dao(), true)

		cacheStore, err := cache.New(config.CacheConf{Type: config.CacheTypeBuiltin, Expires: 30})
		require.NoError(t, err)
		defer cacheStore.Close()
		daoCache := dao.NewCacheAdaptor(cacheStore)
		app.Dao().SetCache(daoCache)
		require.NoError(t, daoCache.UserCacheSave(&source))
		require.NoError(t, daoCache.UserCacheSave(&target))
		require.NoError(t, daoCache.CommentCacheSave(&comment))

		trigger := fmt.Sprintf(`
			CREATE TRIGGER fail_notify_merge
			BEFORE UPDATE OF user_id ON atk_notifies
			WHEN OLD.user_id = %d
			BEGIN
				SELECT RAISE(ABORT, 'injected merge failure');
			END;
		`, source.ID)
		require.NoError(t, app.Dao().DB().Exec(trigger).Error)

		handler.AuthMergeApply(app.App, fiber)
		resp := mergeUsersRequest(t, fiber, app.Conf().AppKey, source, target.Name)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

		assertRecordUserID(t, app.Dao(), &comment, source.ID)
		assertRecordUserID(t, app.Dao(), &notify, source.ID)
		assertRecordUserID(t, app.Dao(), &vote, source.ID)
		assertRecordUserID(t, app.Dao(), &identity, source.ID)

		var users []entity.User
		require.NoError(t, app.Dao().DB().Where("id IN ?", []uint{source.ID, target.ID}).Find(&users).Error)
		require.Len(t, users, 2)
		for _, user := range users {
			if user.ID == target.ID {
				assert.False(t, user.IsAdmin)
			}
		}

		var cachedUser entity.User
		require.NoError(t, daoCache.FindCache(fmt.Sprintf(dao.UserByIDKey, source.ID), &cachedUser))
		var cachedComment entity.Comment
		require.NoError(t, daoCache.FindCache(fmt.Sprintf(dao.CommentByIDKey, comment.ID), &cachedComment))
	})
}

func createMergeFixtures(t *testing.T, db *dao.Dao, sourceAdmin bool) (
	target entity.User,
	source entity.User,
	comment entity.Comment,
	notify entity.Notify,
	vote entity.Vote,
	identity entity.AuthIdentity,
) {
	t.Helper()

	email := fmt.Sprintf("merge-%t@example.com", sourceAdmin)
	target = entity.User{Name: "merge-target", Email: email}
	source = entity.User{Name: "merge-source", Email: email, IsAdmin: sourceAdmin}
	require.NoError(t, db.DB().Create(&target).Error)
	require.NoError(t, db.DB().Create(&source).Error)

	comment = entity.Comment{Content: "merge comment", UserID: source.ID}
	require.NoError(t, db.DB().Create(&comment).Error)
	notify = entity.Notify{UserID: source.ID, CommentID: comment.ID}
	require.NoError(t, db.DB().Create(&notify).Error)
	vote = entity.Vote{UserID: source.ID, TargetID: comment.ID, Type: entity.VoteTypeCommentUp}
	require.NoError(t, db.DB().Create(&vote).Error)
	identity = entity.AuthIdentity{Provider: "test", RemoteUID: fmt.Sprintf("merge-%d", source.ID), UserID: source.ID}
	require.NoError(t, db.DB().Create(&identity).Error)

	return target, source, comment, notify, vote, identity
}

func mergeUsersRequest(t *testing.T, fiberApp *fiber.App, appKey string, source entity.User, targetName string) *http.Response {
	t.Helper()

	token, err := common.LoginGetUserToken(source, appKey, 3600)
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/auth/merge", bytes.NewBufferString(fmt.Sprintf(`{"user_name":%q}`, targetName)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := fiberApp.Test(req)
	require.NoError(t, err)
	return resp
}

func assertRecordUserID(t *testing.T, db *dao.Dao, record any, want uint) {
	t.Helper()
	require.NoError(t, db.DB().First(record).Error)

	var got uint
	switch value := record.(type) {
	case *entity.Comment:
		got = value.UserID
	case *entity.Notify:
		got = value.UserID
	case *entity.Vote:
		got = value.UserID
	case *entity.AuthIdentity:
		got = value.UserID
	default:
		t.Fatalf("unsupported record type %T", record)
	}
	assert.Equal(t, want, got)
}

package handler

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/artalkjs/artalk/v2/internal/core"
	"github.com/artalkjs/artalk/v2/internal/dao"
	"github.com/artalkjs/artalk/v2/internal/entity"
	"github.com/artalkjs/artalk/v2/internal/log"
	"github.com/artalkjs/artalk/v2/internal/sync"
	"github.com/artalkjs/artalk/v2/internal/utils"
	"github.com/artalkjs/artalk/v2/server/common"
	"github.com/gofiber/fiber/v2"
	"github.com/samber/lo"
	"gorm.io/gorm"
)

type RequestAuthDataMergeApply struct {
	UserName string `json:"user_name" validate:"required"`
}

type ResponseAuthDataMergeApply struct {
	UpdatedComment int64  `json:"update_comments_count"`
	UpdatedNotify  int64  `json:"update_notifies_count"`
	UpdatedVote    int64  `json:"update_votes_count"`
	DeletedUser    int64  `json:"deleted_user_count"`
	UserToken      string `json:"user_token"` // Empty if login user is target user no need to re-login
}

// @Id           ApplyDataMerge
// @Summary      Apply data merge
// @Description  This function is to solve the problem of multiple users with the same email address, should be called after user login and then check, and perform data merge.
// @Tags         Auth
// @Security     ApiKeyAuth
// @Param        data  body  RequestAuthDataMergeApply  true  "The data"
// @Success      200  {object}  ResponseAuthDataMergeApply
// @Failure      400  {object}  Map{msg=string}
// @Failure      500  {object}  Map{msg=string}
// @Accept       json
// @Produce      json
// @Router       /auth/merge  [post]
func AuthMergeApply(app *core.App, router fiber.Router) {
	mutexMap := sync.NewKeyMutex[uint]()

	router.Post("/auth/merge", common.LoginGuard(app, func(c *fiber.Ctx, user entity.User) error {
		// Mutex for each user to avoid concurrent merge operation
		mutex := mutexMap.GetLock(user.ID)
		mutex.Lock()
		defer mutex.Unlock()

		if user.Email == "" {
			return common.RespError(c, 500, "User email is empty")
		}

		var p RequestAuthDataMergeApply
		if isOK, resp := common.ParamsDecode(c, &p); !isOK {
			return resp
		}

		resp := ResponseAuthDataMergeApply{}
		var (
			sameEmailUsers []entity.User
			targetUser     entity.User
			otherUsers     []entity.User
			movedComments  []entity.Comment
		)

		// Functions for log
		getMergeLogSummary := func() string {
			getUserInfo := func(u entity.User) string {
				return fmt.Sprintf("[%d, %s, %s]", u.ID, strconv.Quote(u.Name), strconv.Quote(u.Email))
			}
			getUsersInfo := func(otherUsers []entity.User) string {
				return strings.Join(lo.Map(otherUsers, func(u entity.User, _ int) string { return getUserInfo(u) }), ", ")
			}
			return " | " + getUsersInfo(otherUsers) + " -> " + getUserInfo(targetUser)
		}

		// Merge all user data, including user rows, in one transaction.
		if err := app.Dao().DB().Transaction(func(tx *gorm.DB) error {
			if err := tx.Where("LOWER(email) = LOWER(?)", user.Email).Find(&sameEmailUsers).Error; err != nil {
				return err
			}
			if len(sameEmailUsers) == 0 {
				return errors.New("no user with same email")
			}

			targetName := strings.TrimSpace(p.UserName)
			targetEmail := strings.TrimSpace(user.Email)
			targetLink := strings.TrimSpace(user.Link)
			if targetName == "" || targetEmail == "" {
				return errors.New("name and email are required")
			}
			if targetLink != "" && !utils.ValidateURL(targetLink) {
				targetLink = ""
			}

			findTarget := tx.Where(
				"LOWER(name) = LOWER(?) AND LOWER(email) = LOWER(?)",
				targetName,
				targetEmail,
			).First(&targetUser)
			if errors.Is(findTarget.Error, gorm.ErrRecordNotFound) {
				targetUser = entity.User{Name: targetName, Email: targetEmail, Link: targetLink}
				if err := tx.Create(&targetUser).Error; err != nil {
					return err
				}
			} else if findTarget.Error != nil {
				return findTarget.Error
			}

			// Recover administrator permission on the selected target user.
			isAdmin := false
			for _, u := range sameEmailUsers {
				if u.IsAdmin {
					isAdmin = true
					break
				}
			}
			targetUser.IsAdmin = isAdmin
			if result := tx.Model(&entity.User{}).Where("id = ?", targetUser.ID).Update("is_admin", isAdmin); result.Error != nil {
				return result.Error
			}

			otherUsers = lo.Filter(sameEmailUsers, func(u entity.User, _ int) bool {
				return u.ID != targetUser.ID
			})
			otherUserIDs := lo.Map(otherUsers, func(u entity.User, _ int) uint { return u.ID })
			if len(otherUserIDs) == 0 {
				return nil
			}

			if err := tx.Where("user_id IN ?", otherUserIDs).Find(&movedComments).Error; err != nil {
				return err
			}

			// Merge all user data to target user
			// @note Search code files under `./internal/entity` keyword 'UserID` to find all related tables
			for _, u := range otherUsers {
				// comments
				if result := tx.Model(&entity.Comment{}).
					Where("user_id = ?", u.ID).Update("user_id", targetUser.ID); result.Error != nil {
					return result.Error // if error the whole transaction will be rollback
				} else {
					resp.UpdatedComment += result.RowsAffected
				}

				// notifies
				if result := tx.Model(&entity.Notify{}).
					Where("user_id = ?", u.ID).Update("user_id", targetUser.ID); result.Error != nil {
					return result.Error
				} else {
					resp.UpdatedNotify += result.RowsAffected
				}

				// votes
				if result := tx.Model(&entity.Vote{}).
					Where("user_id = ?", u.ID).Update("user_id", targetUser.ID); result.Error != nil {
					return result.Error
				} else {
					resp.UpdatedVote += result.RowsAffected
				}

				// auth_identities
				if result := tx.Model(&entity.AuthIdentity{}).
					Where("user_id = ?", u.ID).Update("user_id", targetUser.ID); result.Error != nil {
					return result.Error
				}
			}

			deleted := tx.Unscoped().Where("id IN ?", otherUserIDs).Delete(&entity.User{})
			if deleted.Error != nil {
				return deleted.Error
			}
			resp.DeletedUser = deleted.RowsAffected

			return nil
		}); err != nil {
			log.Error("Failed to merge user data: ", err.Error(), getMergeLogSummary())
			return common.RespError(c, 500, "Failed to merge data")
		}
		// Invalidate caches only after the transaction has committed.
		app.Dao().CacheAction(func(cache *dao.DaoCache) {
			cache.UserCacheDel(&targetUser)
			_ = cache.UserCacheSave(&targetUser)
			for _, u := range otherUsers {
				cache.UserCacheDel(&u)
			}
			for _, comment := range movedComments {
				cache.CommentCacheDel(&comment)
			}
		})

		// Re-login
		jwtToken, err := common.LoginGetUserToken(targetUser, app.Conf().AppKey, app.Conf().LoginTimeout)
		if err != nil {
			return common.RespError(c, 500, "Failed to re-login")
		}
		resp.UserToken = jwtToken

		// Log
		log.Info("User data merged successfully", getMergeLogSummary(), " | ",
			"Updated Comments: ", resp.UpdatedComment, " | ",
			"Updated Notifies: ", resp.UpdatedNotify, " | ",
			"Updated Votes: ", resp.UpdatedVote, " | ",
			"Deleted Users: ", resp.DeletedUser)

		return common.RespData(c, resp)
	}))
}

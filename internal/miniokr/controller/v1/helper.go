package v1

import (
	"github.com/gin-gonic/gin"

	"github.com/imxw/miniokr/internal/miniokr/services/user"
	"github.com/imxw/miniokr/internal/pkg/core"
	"github.com/imxw/miniokr/internal/pkg/errno"
	"github.com/imxw/miniokr/internal/pkg/known"
	"github.com/imxw/miniokr/internal/pkg/log"
)

// Contains checks if an element is present in a slice.
// T is a type parameter constrained to types that are comparable.
func Contains[T comparable](slice []T, element T) bool {
	for _, v := range slice {
		if v == element {
			return true
		}
	}
	return false
}

func CheckPermission(c *gin.Context, SrcUserId string, roles []string, targetUserId string, userService user.Service) bool {

	if SrcUserId == targetUserId {
		return true
	}

	for _, role := range roles {
		if role == known.AdminRoleName {
			return true
		}
	}

	for _, role := range roles {
		if role == known.LeaderRoleName {
			managedUserIDs, err := userService.GetManagedUserIDs(c, SrcUserId)
			if err != nil {
				log.C(c).Errorw("failed to get managed user IDs", "err", err)
				core.WriteResponse(c, errno.InternalServerError, nil)
				return false
			}

			for _, managerUserID := range managedUserIDs {
				if targetUserId == managerUserID {
					return true
				}
			}
			break
		}
	}
	core.WriteResponse(c, errno.ErrForbidden, nil)
	return false
}

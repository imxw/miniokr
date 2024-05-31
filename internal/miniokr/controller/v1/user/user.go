package user

import (
	"github.com/gin-gonic/gin"

	ctrlV1 "github.com/imxw/miniokr/internal/miniokr/controller/v1"
	"github.com/imxw/miniokr/internal/miniokr/services/user"
	"github.com/imxw/miniokr/internal/pkg/core"
	"github.com/imxw/miniokr/internal/pkg/errno"
	"github.com/imxw/miniokr/internal/pkg/known"
)

type Controller struct {
	us user.Service
}

func New(us user.Service) *Controller {
	return &Controller{us: us}
}

// func (ctrl *Controller) GetUser(c *gin.Context) {
// 	var req v1.UserRequest
// 	if err := c.ShouldBindQuery(&req); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request parameters"})
// 		return
// 	}

// 	var user *v1.UserResponse
// 	var err error

// 	if req.ID != "" {
// 		user, err = ctrl.us.GetUserByID(c, req.ID)
// 	} else if req.Name != "" {
// 		user, err = ctrl.us.GetUserByName(c, req.Name)
// 	} else {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "id or username query parameter required"})
// 		return
// 	}

// 	if err != nil {
// 		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
// 		return
// 	}
// 	c.JSON(http.StatusOK, user)
// }

func (ctrl *Controller) GetCurrentUser(c *gin.Context) {

	userID, ok := c.MustGet(known.XUserIDKey).(string)
	if !ok {
		core.WriteResponse(c, errno.InternalServerError, nil)
		return
	}
	user, err := ctrl.us.GetUserByID(c, userID)
	if err != nil {
		core.WriteResponse(c, errno.ErrUserNotFound, nil)
		return
	}
	core.WriteResponse(c, nil, user)
}

func (ctrl *Controller) GetDepartmentsTree(c *gin.Context) {
	userID, ok := c.MustGet(known.XUserIDKey).(string)
	if !ok {
		core.WriteResponse(c, errno.InternalServerError, nil)
		return
	}

	// 判断用户角色
	roles, err := ctrl.us.GetUserRolesByID(c, userID)
	if err != nil {
		core.WriteResponse(c, errno.InternalServerError, nil)
		return
	}
	// 根据角色返回相应的树
	if len(roles) == 0 {
		core.WriteResponse(c, errno.ErrForbidden, nil)
		return
	}

	if ctrlV1.Contains(roles, known.AdminRoleName) {

		tree, err := ctrl.us.GetCompanyDepartmentTree(c)
		if err != nil {
			core.WriteResponse(c, errno.InternalServerError, nil)
			return
		}
		core.WriteResponse(c, nil, tree)

	} else if ctrlV1.Contains(roles, known.LeaderRoleName) {
		tree, err := ctrl.us.GetUserDepartmentTree(c, userID)
		if err != nil {
			core.WriteResponse(c, errno.InternalServerError, nil)
			return
		}

		core.WriteResponse(c, nil, tree)

	} else {
		core.WriteResponse(c, errno.ErrForbidden, nil)
		return
	}

}

func (ctrl *Controller) GetUserDepartmentsTree(c *gin.Context) {

	userID := c.Param("id")

	tree, err := ctrl.us.GetUserDepartmentTree(c, userID)
	if err != nil {
		core.WriteResponse(c, err, nil)
		return
	}

	core.WriteResponse(c, nil, tree)
}

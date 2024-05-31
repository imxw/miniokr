package sync

import (
	"context"
	"time"

	"github.com/zhaoyunxing92/dingtalk/v2"
	"github.com/zhaoyunxing92/dingtalk/v2/request"
	"github.com/zhaoyunxing92/dingtalk/v2/response"

	"github.com/imxw/miniokr/internal/miniokr/services/notify"
	"github.com/imxw/miniokr/internal/miniokr/services/user"
	"github.com/imxw/miniokr/internal/miniokr/store"
	"github.com/imxw/miniokr/internal/pkg/log"
	"github.com/imxw/miniokr/internal/pkg/model"
)

var _ Service = (*SyncService)(nil)

var rootDeptId = user.RootDeptID

type SyncService struct {
	dingClient     *dingtalk.DingTalk
	store          store.SyncStorer
	notifier       notify.Notifier
	excludeDeptIDs map[int]bool
}

// NewSyncService 创建一个新的 SyncService 实例
func NewSyncService(dingClient *dingtalk.DingTalk, store store.SyncStorer, notifier notify.Notifier, excludeDeptIDs map[int]bool) *SyncService {
	return &SyncService{
		dingClient:     dingClient,
		store:          store,
		notifier:       notifier,
		excludeDeptIDs: excludeDeptIDs,
	}
}

func (s *SyncService) SyncDepartmentsAndUsers(ctx context.Context) error {
	departments, err := fetchAndPersistAllDepartments(s.dingClient, rootDeptId, s.excludeDeptIDs)
	if err != nil {
		log.Errorw("Department sync failed", "err", err)
		s.notifier.Send("DingTalk department sync task failed: " + err.Error())
		return err
	}

	users, userDepts, err := fetchAndPersistAllUsers(s.dingClient, rootDeptId, s.excludeDeptIDs)
	if err != nil {
		log.Errorw("User sync failed", "err", err)
		s.notifier.Send("DingTalk user sync task failed: " + err.Error())
		return err
	}

	err = persistDepartments(s.store, departments)
	if err != nil {
		log.Errorw("Persist departments failed", "err", err)
		s.notifier.Send("Persist departments task failed: " + err.Error())
		return err
	}

	err = persistUsers(s.store, users)
	if err != nil {
		log.Errorw("Persist users failed", "err", err)
		s.notifier.Send("Persist users task failed: " + err.Error())
		return err
	}

	err = persistUserDepartments(s.store, userDepts)
	if err != nil {
		log.Errorw("Persist user departments failed", "err", err)
		s.notifier.Send("Persist user departments task failed: " + err.Error())
		return err
	}

	s.notifier.Send("DingTalk sync task succeeded")
	return nil
}

func fetchAndPersistAllDepartments(dingClient *dingtalk.DingTalk, rootDeptId int, excludeDeptIDs map[int]bool) ([]model.Department, error) {
	departments, err := getAllDeptIds(dingClient, rootDeptId, nil, excludeDeptIDs)
	if err != nil {
		return nil, err
	}

	return departments, nil
}

// 递归获取所有部门ID和名称
func getAllDeptIds(dingClient *dingtalk.DingTalk, deptId int, parentId *int, excludeDeptIDs map[int]bool) ([]model.Department, error) {
	var departments []model.Department

	// 检查当前部门是否在排除列表中
	if excludeDeptIDs[deptId] {
		return departments, nil
	}

	req := &request.DeptList{
		DeptId: deptId,
	}
	var res response.DeptList
	err := withRetry(func() error {
		var err error
		res, err = dingClient.GetDeptList(req)
		return err
	})
	if err != nil {
		return nil, err
	}

	for idx, dept := range res.List {
		if excludeDeptIDs[dept.Id] {
			continue
		}
		department := model.Department{
			DepartmentID: dept.Id,
			Name:         dept.Name,
			ParentID:     &dept.ParentId,
			Sort:         idx,
		}
		departments = append(departments, department)

		// 递归获取子部门
		subDepartments, err := getAllDeptIds(dingClient, dept.Id, &dept.ParentId, excludeDeptIDs)
		if err != nil {
			return nil, err
		}
		departments = append(departments, subDepartments...)
	}

	return departments, nil
}

func fetchAndPersistAllUsers(dingClient *dingtalk.DingTalk, rootDeptId int, excludeDeptIDs map[int]bool) ([]model.User, []model.UserDepartment, error) {
	users, userDepts, err := getAllDeptUsers(dingClient, rootDeptId, excludeDeptIDs)
	if err != nil {
		return nil, nil, err
	}

	return users, userDepts, nil
}

// Get all users and user-department mappings recursively
func getAllDeptUsers(dingClient *dingtalk.DingTalk, deptId int, excludeDeptIDs map[int]bool) ([]model.User, []model.UserDepartment, error) {
	var users []model.User
	var userDepts []model.UserDepartment

	// 用于递归遍历所有部门的函数
	var getUsersFromDept func(deptId int) error

	// 用于存储已访问的部门，防止重复访问
	visitedDepts := make(map[int]bool)

	getUsersFromDept = func(deptId int) error {
		if visitedDepts[deptId] {
			return nil
		}
		visitedDepts[deptId] = true

		// 检查是否需要排除该部门
		if excludeDeptIDs[deptId] {
			return nil
		}

		var cursor int
		for {
			req := &request.DeptDetailUserInfo{
				DeptId: deptId,
				Cursor: cursor,
				Size:   100,
			}

			var res response.DeptDetailUserInfo
			err := withRetry(func() error {
				var err error
				res, err = dingClient.GetDeptDetailUserInfo(req)
				return err
			})
			if err != nil {
				return err
			}

			for idx, userInfo := range res.Page.List {
				hiredDate := time.Unix(int64(userInfo.HiredDate/1000), 0)
				var hiredDatePtr *time.Time
				if userInfo.HiredDate != 0 {
					hiredDatePtr = &hiredDate
				}
				user := model.User{
					UserID:    userInfo.UserId,
					Name:      userInfo.Name,
					Title:     userInfo.Title,
					Status:    "active",
					Mobile:    userInfo.Mobile,
					Avatar:    userInfo.Avatar,
					JobNumber: userInfo.JobNumber,
					HiredDate: hiredDatePtr,
					Sort:      idx,
				}
				users = append(users, user)

				userDept := model.UserDepartment{
					UserID:       userInfo.UserId,
					DepartmentID: deptId,
					IsLeader:     model.Bool(0),
				}
				if userInfo.Leader {
					userDept.IsLeader = model.Bool(1)
				}
				userDepts = append(userDepts, userDept)
			}

			if !res.Page.HasMore {
				break
			}
			cursor = res.Page.NextCursor
		}

		subDeptList, err := dingClient.GetSubDeptList(deptId)
		if err != nil {
			return err
		}

		for _, subDeptId := range subDeptList.Result.Ids {
			err := getUsersFromDept(subDeptId)
			if err != nil {
				return err
			}
		}

		return nil
	}

	err := getUsersFromDept(deptId)
	if err != nil {
		return nil, nil, err
	}

	return users, userDepts, nil
}

func persistDepartments(store store.SyncStorer, departments []model.Department) error {
	return store.PersistDepartments(nil, departments)
}

func persistUsers(store store.SyncStorer, users []model.User) error {
	return store.PersistUsers(nil, users)
}

func persistUserDepartments(store store.SyncStorer, userDepts []model.UserDepartment) error {
	return store.PersistUserDepartments(nil, userDepts)
}

// 使用重试机制的函数
func withRetry(fn func() error) error {
	const maxRetries = 5
	const baseDelay = time.Second
	var err error
	for i := 0; i < maxRetries; i++ {
		err = fn()
		if err == nil {
			return nil
		}
		log.Errorw("请求失败，重试次数", "error", err, "retry", i+1)
		time.Sleep(baseDelay * time.Duration(i+1))
	}
	return err
}

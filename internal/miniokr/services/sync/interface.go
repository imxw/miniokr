package sync

import "context"

type Service interface {
	SyncDepartmentsAndUsers(context.Context) error
}

package sync

import (
	"context"

	"github.com/robfig/cron/v3"

	"github.com/imxw/miniokr/internal/miniokr/services/sync"
)

type Controller struct {
	syncService sync.Service
}

func NewSyncController(syncService sync.Service) *Controller {
	return &Controller{syncService: syncService}
}

func (c *Controller) StartCronJob() {
	cronScheduler := cron.New()
	cronScheduler.AddFunc("15 1 * * *", func() {
		ctx := context.Background()
		c.syncService.SyncDepartmentsAndUsers(ctx)
	})
	cronScheduler.Start()
}

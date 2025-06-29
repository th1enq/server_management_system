package jobs

import (
	"context"

	"github.com/th1enq/server_management_system/internal/services"
)

type IntervalCheckStatus interface {
	Run(context.Context) error
}

type intervalCheckStatus struct {
	serverService services.ServerService
}

func NewIntervalCheckStatus(serverService services.ServerService) IntervalCheckStatus {
	return &intervalCheckStatus{
		serverService: serverService,
	}
}
func (s *intervalCheckStatus) Run(ctx context.Context) error {
	return s.serverService.CheckServerStatus(ctx)
}

package presenters

import "github.com/google/wire"

var WireSet = wire.NewSet(
	NewAuthPresenter,
	NewServerPresenter,
	NewUserPresenter,
	NewReportPresenter,
	NewJobsPresenter,
)

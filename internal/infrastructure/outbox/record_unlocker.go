package outbox

import (
	"time"

	"github.com/th1enq/server_management_system/internal/infrastructure/database"
)

type recordUnlocker struct {
	store                   database.DatabaseClient
	MaxLockTimeDurationMins time.Duration
}

func newRecordUnlocker(store database.DatabaseClient, maxLockTimeDurationMins time.Duration) recordUnlocker {
	return recordUnlocker{
		store:                   store,
		MaxLockTimeDurationMins: maxLockTimeDurationMins,
	}
}

func (d recordUnlocker) UnlockExpiredMessages() error {
	unlockTime := time.Now().Add(-d.MaxLockTimeDurationMins * time.Minute).UTC()
	err := d.store.ClearLocksWithDurationBeforeDate(unlockTime)
	if err != nil {
		return err
	}
	return nil
}

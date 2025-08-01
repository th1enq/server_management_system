package outbox

import (
	"time"

	"github.com/th1enq/server_management_system/internal/infrastructure/database"
)

type recordCleaner struct {
	store             database.DatabaseClient
	MaxRecordLifetime time.Duration
}

func newRecordCleaner(store database.DatabaseClient, maxRecordLifetime time.Duration) recordCleaner {
	return recordCleaner{
		store:             store,
		MaxRecordLifetime: maxRecordLifetime,
	}
}

func (d recordCleaner) RemoveExpiredMessages() error {
	expiryTime := time.Now().Add(-d.MaxRecordLifetime)
	err := d.store.RemoveRecordsBeforeDatetime(expiryTime)
	if err != nil {
		return err
	}
	return nil
}

package outbox

import (
	"fmt"
	"time"

	"github.com/th1enq/server_management_system/internal/configs"
	"github.com/th1enq/server_management_system/internal/infrastructure/database"
	"github.com/th1enq/server_management_system/internal/infrastructure/models"
	"github.com/th1enq/server_management_system/internal/infrastructure/mq"
)

type defaultRecordProcessor struct {
	messageBroker mq.MessageBroker
	store         database.DatabaseClient
	machineID     string
	cfg           configs.RetrialPolicy
}

func newProcessor(store database.DatabaseClient, messageBroker mq.MessageBroker, machineID string, cfg configs.RetrialPolicy) *defaultRecordProcessor {
	return &defaultRecordProcessor{
		store:         store,
		messageBroker: messageBroker,
		machineID:     machineID,
		cfg:           cfg,
	}
}

func (d defaultRecordProcessor) lockUnprocessedEntities() error {
	lockTime := time.Now().UTC()
	lockErr := d.store.UpdateRecordLockByState(d.machineID, lockTime, models.PendingDelivery)
	if lockErr != nil {
		return lockErr
	}
	return nil
}

func (d defaultRecordProcessor) publishMessages(records []models.Record) error {
	for _, rec := range records {
		// Send message to message broker
		now := time.Now().UTC()
		rec.LastAttemptAt = &now
		rec.NumberOfAttempts++
		err := d.messageBroker.Send(rec.Message)
		// If an error occurs, remove the lock information, update retrial times and continue
		if err != nil {
			rec.LockedAt = nil
			rec.LockID = nil
			errorMsg := err.Error()
			rec.Error = &errorMsg
			if d.cfg.MaxSendAttemptsEnabled && rec.NumberOfAttempts == d.cfg.MaxSendAttempts {
				rec.State = models.MaxAttemptsReached
			}
			dbErr := d.store.UpdateRecordByID(rec)
			if dbErr != nil {
				return fmt.Errorf("could not update the record in the db: %w", dbErr)
			}

			return fmt.Errorf("an error occurred when trying to send the message to the broker: %w", err)
		}

		// Remove lock information and update state
		rec.State = models.Delivered
		rec.LockedAt = nil
		rec.LockID = nil
		rec.ProcessedAt = &now
		err = d.store.UpdateRecordByID(rec)

		if err != nil {
			return fmt.Errorf("culd not update the record in the db: %w", err)
		}
	}
	return nil
}

func (d defaultRecordProcessor) ProcessRecords() error {
	err := d.lockUnprocessedEntities()
	defer d.store.ClearLocksByLockID(d.machineID)
	if err != nil {
		return err
	}

	records, err := d.store.GetRecordsByLockID(d.machineID)
	if err != nil {
		return err
	}

	if len(records) == 0 {
		return nil
	}
	return d.publishMessages(records)
}

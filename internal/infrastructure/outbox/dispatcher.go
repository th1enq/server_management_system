package outbox

import (
	"log"
	"time"

	"github.com/th1enq/server_management_system/internal/configs"
	"github.com/th1enq/server_management_system/internal/infrastructure/database"
	"github.com/th1enq/server_management_system/internal/infrastructure/mq"
)

type processor interface {
	ProcessRecords() error
}

type unlocker interface {
	UnlockExpiredMessages() error
}

type cleaner interface {
	RemoveExpiredMessages() error
}

// Dispatcher initializes and runs the outbox dispatcher
type Dispatcher struct {
	recordProcessor processor
	recordUnlocker  unlocker
	recordCleaner   cleaner
	settings        configs.Dispatcher
}

// NewDispatcher constructor
func NewDispatcher(store database.DatabaseClient, broker mq.MessageBroker, cfg configs.Dispatcher) Dispatcher {
	return Dispatcher{
		recordProcessor: newProcessor(
			store,
			broker,
			cfg.MachineID,
			cfg.RetrialPolicy,
		),
		recordUnlocker: newRecordUnlocker(
			store,
			cfg.MaxLockTimeDuration,
		),
		recordCleaner: newRecordCleaner(
			store,
			cfg.MessagesRetentionDuration,
		),
		settings: cfg,
	}
}

// Run periodically checks for new outbox messages from the Store, sends the messages through the MessageBroker
// and updates the message status accordingly
func (d Dispatcher) Run(errChan chan<- error, doneChan <-chan struct{}) {
	doneProc := make(chan struct{}, 1)
	doneUnlock := make(chan struct{}, 1)
	doneClear := make(chan struct{}, 1)

	go func() {
		<-doneChan
		doneProc <- struct{}{}
		doneUnlock <- struct{}{}
		doneClear <- struct{}{}
	}()

	go d.runRecordProcessor(errChan, doneProc)
	go d.runRecordUnlocker(errChan, doneUnlock)
	go d.runRecordCleaner(errChan, doneClear)
}

// runRecordProcessor processes the unsent records of the store
func (d Dispatcher) runRecordProcessor(errChan chan<- error, doneChan <-chan struct{}) {
	ticker := time.NewTicker(d.settings.ProcessInterval)
	for {
		log.Print("Record processor Running")
		err := d.recordProcessor.ProcessRecords()
		if err != nil {
			errChan <- err
		}
		log.Print("Record Processing Finished")

		select {
		case <-ticker.C:
			continue
		case <-doneChan:
			ticker.Stop()
			log.Print("Stopping Record processor")
			return
		}
	}
}

func (d Dispatcher) runRecordUnlocker(errChan chan<- error, doneChan <-chan struct{}) {
	ticker := time.NewTicker(d.settings.LockCheckerInterval)
	for {
		log.Print("Record unlocker Running")
		err := d.recordUnlocker.UnlockExpiredMessages()
		if err != nil {
			errChan <- err
		}
		log.Print("Record unlocker Finished")
		select {
		case <-ticker.C:
			continue
		case <-doneChan:
			ticker.Stop()
			log.Print("Stopping Record unlocker")
			return

		}
	}
}

func (d Dispatcher) runRecordCleaner(errChan chan<- error, doneChan <-chan struct{}) {
	ticker := time.NewTicker(d.settings.CleanupWorkerInterval)
	for {
		log.Print("Record retention cleaner Running")
		err := d.recordCleaner.RemoveExpiredMessages()
		if err != nil {
			errChan <- err
		}
		log.Print("Record retention cleaner Finished")
		select {
		case <-ticker.C:
			continue
		case <-doneChan:
			ticker.Stop()
			log.Print("Stopping Record retention cleaner")
			return

		}
	}
}

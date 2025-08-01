package configs

import "time"

type RetrialPolicy struct {
	MaxSendAttemptsEnabled bool `yaml:"max_send_attempts_enabled"`
	MaxSendAttempts        int  `yaml:"max_send_attempts"`
}

type Dispatcher struct {
	ProcessInterval           time.Duration `yaml:"process_interval"`
	LockCheckerInterval       time.Duration `yaml:"lock_checker_interval"`
	MaxLockTimeDuration       time.Duration `yaml:"max_lock_time_duration"`
	CleanupWorkerInterval     time.Duration `yaml:"cleanup_worker_interval"`
	RetrialPolicy             RetrialPolicy `yaml:"retrial_policy"`
	MessagesRetentionDuration time.Duration `yaml:"messages_retention_duration"`
	MachineID                 string        `yaml:"machine_id"`
}

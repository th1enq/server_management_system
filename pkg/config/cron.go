package config

type SendDailyReport struct {
	Schedule string `yaml:"schedule"`
}

type IntervalCheckStatus struct {
	Schedule string `yaml:"schedule"`
}

type Cron struct {
	SendDailyReport     SendDailyReport     `yaml:"send_daily_report"`
	IntervalCheckStatus IntervalCheckStatus `yaml:"interval_check_status"`
}

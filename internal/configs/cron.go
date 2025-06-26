package configs

type SendDailyReport struct {
	Schedule string `yaml:"schedule"`
}

type Cron struct {
	SendDailyReport SendDailyReport `yaml:"send_daily_report"`
}

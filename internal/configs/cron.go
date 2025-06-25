package configs

type SendDailyReport struct {
	Schedule string `yaml:"schedule"`
	Email    string `yaml:"email"`
}

type Cron struct {
	SendDailyReport SendDailyReport `yaml:"send_daily_report"`
}

package configs

type DailyReport struct {
	Name     string `yaml:"name"`
	Schedule string `yaml:"schedule"`
}

type UpdateStatus struct {
	Name     string `yaml:"name"`
	Schedule string `yaml:"schedule"`
}

type Cron struct {
	DailyReport  DailyReport  `yaml:"daily_report"`
	UpdateStatus UpdateStatus `yaml:"update_status"`
}

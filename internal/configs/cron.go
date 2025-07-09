package configs

type DailyReport struct {
	Name     string `yaml:"name"`
	Schedule string `yaml:"schedule"`
}

type HealthCheckServer struct {
	Name     string `yaml:"name"`
	Schedule string `yaml:"schedule"`
}

type Cron struct {
	DailyReport       DailyReport       `yaml:"daily_report"`
	HealthCheckServer HealthCheckServer `yaml:"health_check_server"`
}

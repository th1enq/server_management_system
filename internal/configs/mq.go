package configs

type MQ struct {
	Addresses      []string        `yaml:"addresses"`
	ClientID       string          `yaml:"client_id"`
	ConsumerGroups []ConsumerGroup `yaml:"consumer_groups"`
}

type ConsumerGroup struct {
	GroupID string   `yaml:"group_id"`
	Topics  []string `yaml:"topics"`
}

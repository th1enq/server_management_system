package config

import "fmt"

type Redis struct {
	Port     int    `yaml:"port"`
	Host     string `yaml:"host"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
	PoolSize int    `yaml:"pool_size"`
}

func (r *Redis) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

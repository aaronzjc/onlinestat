package internal

import (
	"errors"

	"github.com/spf13/viper"
)

type HttpConfig struct {
	Port int `yaml:"port"`
}

type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	Database int    `yaml:"database"`
}
type AppConfig struct {
	Timeout int         `yaml:"timeout"`
	Redis   RedisConfig `yaml:"redis"`
}

type Config struct {
	Name   string               `yaml:"name"`
	Env    string               `yaml:"env"`
	Http   HttpConfig           `yaml:"http"`
	Driver string               `yaml:"driver"`
	Apps   map[string]AppConfig `yaml:"apps"`
}

var (
	config *Config
)

func init() {
	config = new(Config)
}

func LoadConfig(path string) error {
	vip := viper.New()
	vip.SetConfigFile(path)
	vip.SetConfigType("yml")
	if err := vip.ReadInConfig(); err != nil {
		return errors.New("read config err")
	}
	vip.Unmarshal(&config)
	return nil
}

func GetConfig() *Config {
	return config
}

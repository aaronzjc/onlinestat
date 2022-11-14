package internal

import (
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis"
)

type Stater interface {
	Set(string, string) bool
	Get(string) int
}

type RedisStater struct {
	clientMap map[string]*redis.Client
}

var (
	OnlineStater RedisStater
)

func (s *RedisStater) Set(app string, ip string) bool {
	client, ok := s.clientMap[app]
	if !ok {
		return false
	}
	if _, err := client.Set(ip, 1, time.Second*60).Result(); err != nil {
		return false
	}
	return true
}

func (s *RedisStater) Get(app string) int {
	var total int64
	var err error
	client, ok := s.clientMap[app]
	if !ok {
		return 0
	}
	if total, err = client.DbSize().Result(); err != nil {
		return 0
	}
	return int(total)
}

func SetupStater(conf *Config) error {
	OnlineStater.clientMap = make(map[string]*redis.Client)
	for app, item := range conf.Apps {
		if item.Database == 0 {
			return errors.New("cannot use 0 as database")
		}
		client := redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%d", item.Host, item.Port),
			Password: item.Password,
			DB:       item.Database,
		})
		if _, err := client.Ping().Result(); err != nil {
			return errors.New("init default redis err")
		}
		OnlineStater.clientMap[app] = client
	}
	return nil
}

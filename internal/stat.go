package internal

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis"
)

type Stater interface {
	Set(string, string) bool // set online
	Get(string) int          // get online counts
	Run() error              // start stater
	Dump(string) []string    // dump all ips
}

type MemCounter struct {
	timeout time.Duration
	counter map[string]int64
	sync.RWMutex
}

type MemStater struct {
	Apps map[string]*MemCounter
}

func (s *MemStater) Set(app string, ip string) bool {
	stater, ok := s.Apps[app]
	if !ok {
		return false
	}
	stater.Lock()
	stater.counter[ip] = time.Now().UnixMicro()
	stater.Unlock()
	return true
}

func (s *MemStater) Get(app string) int {
	stater, ok := s.Apps[app]
	if !ok {
		return 0
	}
	return len(stater.counter)
}

func (s *MemStater) Run() error {
	clean := func(ctx context.Context, app string) {
		ticker := time.NewTicker(time.Second * 10)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				stater, ok := s.Apps[app]
				if !ok {
					continue
				}
				for ip, activeTime := range stater.counter {
					if activeTime+stater.timeout.Microseconds() < time.Now().UnixMicro() {
						stater.Lock()
						delete(stater.counter, ip)
						stater.Unlock()
					}
				}
			}
		}
	}
	for app, item := range config.Apps {
		s.Apps[app] = &MemCounter{
			timeout: time.Duration(item.Timeout * int(time.Second)),
			counter: make(map[string]int64),
		}
		ctx := context.Background()
		go clean(ctx, app)
	}
	return nil
}

func (s *MemStater) Dump(app string) []string {
	res := []string{}
	stater, ok := s.Apps[app]
	if !ok {
		return res
	}
	for kk := range stater.counter {
		res = append(res, kk)
	}
	return res
}

type RedisStater struct {
	clientMap map[string]*redis.Client
}

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

func (s *RedisStater) Run() error {
	for app, item := range config.Apps {
		if item.Redis.Database == 0 {
			return errors.New("cannot use 0 as database")
		}
		client := redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%d", item.Redis.Host, item.Redis.Port),
			Password: item.Redis.Password,
			DB:       item.Redis.Database,
		})
		if _, err := client.Ping().Result(); err != nil {
			return errors.New("init default redis err")
		}
		s.clientMap[app] = client
	}
	return nil
}

func (s *RedisStater) Dump(app string) []string {
	return []string{}
}

func SetupStater() error {
	if config.Driver == "redis" {
		OnlineStater = &RedisStater{
			clientMap: make(map[string]*redis.Client),
		}
	} else {
		OnlineStater = &MemStater{
			Apps: make(map[string]*MemCounter),
		}
	}
	OnlineStater.Run()
	return nil
}

var (
	OnlineStater Stater
)

// Package config 负责从环境变量加载服务配置。
package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Addr        string
	MaxPageSize int
	// AuthToken 为鉴权中间件的有效访问令牌；为空则鉴权中间件放行。
	AuthToken string
	// RateLimit 为每客户端每分钟允许的最大请求数；0 表示不限流。
	RateLimit int
}

func Load() *Config {
	cfg := &Config{
		Addr:        ":" + getenv("PORT", "8080"),
		MaxPageSize: getenvInt("MAX_PAGE_SIZE", 100),
		AuthToken:   os.Getenv("AUTH_TOKEN"),
		RateLimit:   getenvInt("RATE_LIMIT", 0),
	}
	if v := os.Getenv("ADDR"); v != "" {
		cfg.Addr = v
	}
	return cfg
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getenvInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return def
	}
	return n
}

func (c *Config) String() string {
	return fmt.Sprintf("addr=%s max_page_size=%d rate_limit=%d auth_enabled=%v",
		c.Addr, c.MaxPageSize, c.RateLimit, c.AuthToken != "")
}

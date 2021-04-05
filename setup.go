package xcloud_dopaas_coredns

import (
	"strconv"
	"time"

	"github.com/caddyserver/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
)

const (
	PluginName = "xcloud_dopaas_coredns"
)

func init() {
	caddy.RegisterPlugin(PluginName, caddy.Plugin{
		ServerType: "dns",
		Action:     setup,
	})
}

func setup(c *caddy.Controller) error {
	redis, err := initRedisClusterClient(c)
	if err != nil {
		return plugin.Error("redis", err)
	}
	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		redis.Next = next
		return redis
	})
	return nil
}

func initRedisClusterClient(c *caddy.Controller) (*Redis, error) {
	redis := Redis{
		address:            "localhost:6379,localhost:6380,localhost:6381,localhost:7379,localhost:7380,localhost:7381",
		password:           "",
		connectTimeout:     5000,
		readTimeout:        10000,
		writeTimeout:       5000,
		maxRetries:         10,
		poolSize:           10,
		ttl:                360,
		keyPrefix:          "_coredns:",
		localCacheExpireMs: 5000,
	}

	for c.Next() {
		if c.NextBlock() {
			for {
				switch c.Val() {
				case "address":
					if !c.NextArg() {
						return &Redis{}, c.ArgErr()
					}
					if c.Val() != "" {
						redis.address = c.Val()
					}
				case "password":
					if !c.NextArg() {
						return &Redis{}, c.ArgErr()
					}
					if c.Val() != "" {
						redis.password = c.Val()
					}
				case "connect_timeout":
					if !c.NextArg() {
						return &Redis{}, c.ArgErr()
					}
					_connectTimeout, err := strconv.Atoi(c.Val())
					if err == nil {
						redis.connectTimeout = time.Duration(_connectTimeout)
					}
				case "read_timeout":
					if !c.NextArg() {
						return &Redis{}, c.ArgErr()
					}
					_readTimeout, err1 := strconv.Atoi(c.Val())
					if err1 == nil {
						redis.readTimeout = time.Duration(_readTimeout)
					}
				case "write_timeout":
					if !c.NextArg() {
						return &Redis{}, c.ArgErr()
					}
					_writeTimeout, err2 := strconv.Atoi(c.Val())
					if err2 == nil {
						redis.writeTimeout = time.Duration(_writeTimeout)
					}
				case "max_retries":
					if !c.NextArg() {
						return &Redis{}, c.ArgErr()
					}
					_maxRetries, err3 := strconv.Atoi(c.Val())
					if err3 == nil {
						redis.maxRetries = _maxRetries
					}
				case "pool_size":
					if !c.NextArg() {
						return &Redis{}, c.ArgErr()
					}
					_poolSize, err4 := strconv.Atoi(c.Val())
					if err4 == nil {
						redis.poolSize = _poolSize
					}
				case "ttl":
					if !c.NextArg() {
						return &Redis{}, c.ArgErr()
					}
					_ttl, err5 := strconv.Atoi(c.Val())
					if err5 == nil {
						redis.ttl = uint32(_ttl)
					}
				case "prefix":
					if !c.NextArg() {
						return &Redis{}, c.ArgErr()
					}
					if c.Val() != "" {
						redis.keyPrefix = c.Val()
					}
				case "local_cache_expire_ms":
					if !c.NextArg() {
						return &Redis{}, c.ArgErr()
					}
					localCacheExpireMs, err6 := strconv.Atoi(c.Val())
					if err6 != nil {
						redis.localCacheExpireMs = int64(localCacheExpireMs)
					}
				default:
					if c.Val() != "}" {
						return &Redis{}, c.Errf("Unknown config property '%s'", c.Val())
					}
				}
				if !c.Next() {
					break
				}
			}
		}
		redis.Connect()
		return &redis, nil
	}
	return &Redis{}, nil
}

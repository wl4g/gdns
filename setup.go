package coredns_gdns

import (
	"strconv"
	"time"

	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
)

const (
	PluginName = "coredns_gdns"
)

func init() {
	caddy.RegisterPlugin(PluginName, caddy.Plugin{
		ServerType: "dns",
		Action:     setup,
	})
}

func setup(c *caddy.Controller) error {
	redisService, err := initRedisClusterClient(c)
	if err != nil {
		return plugin.Error("coredns_gdns", err)
	}
	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		redisService.Next = next
		return redisService
	})
	return nil
}

func initRedisClusterClient(c *caddy.Controller) (*RedisService, error) {
	redisService := RedisService{
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
						return &RedisService{}, c.ArgErr()
					}
					if c.Val() != "" {
						redisService.address = c.Val()
					}
				case "password":
					if !c.NextArg() {
						return &RedisService{}, c.ArgErr()
					}
					if c.Val() != "" {
						redisService.password = c.Val()
					}
				case "connect_timeout":
					if !c.NextArg() {
						return &RedisService{}, c.ArgErr()
					}
					_connectTimeout, err := strconv.Atoi(c.Val())
					if err == nil {
						redisService.connectTimeout = time.Duration(_connectTimeout)
					}
				case "read_timeout":
					if !c.NextArg() {
						return &RedisService{}, c.ArgErr()
					}
					_readTimeout, err1 := strconv.Atoi(c.Val())
					if err1 == nil {
						redisService.readTimeout = time.Duration(_readTimeout)
					}
				case "write_timeout":
					if !c.NextArg() {
						return &RedisService{}, c.ArgErr()
					}
					_writeTimeout, err2 := strconv.Atoi(c.Val())
					if err2 == nil {
						redisService.writeTimeout = time.Duration(_writeTimeout)
					}
				case "max_retries":
					if !c.NextArg() {
						return &RedisService{}, c.ArgErr()
					}
					_maxRetries, err3 := strconv.Atoi(c.Val())
					if err3 == nil {
						redisService.maxRetries = _maxRetries
					}
				case "pool_size":
					if !c.NextArg() {
						return &RedisService{}, c.ArgErr()
					}
					_poolSize, err4 := strconv.Atoi(c.Val())
					if err4 == nil {
						redisService.poolSize = _poolSize
					}
				case "ttl":
					if !c.NextArg() {
						return &RedisService{}, c.ArgErr()
					}
					_ttl, err5 := strconv.Atoi(c.Val())
					if err5 == nil {
						redisService.ttl = uint32(_ttl)
					}
				case "prefix":
					if !c.NextArg() {
						return &RedisService{}, c.ArgErr()
					}
					if c.Val() != "" {
						redisService.keyPrefix = c.Val()
					}
				case "local_cache_expire_ms":
					if !c.NextArg() {
						return &RedisService{}, c.ArgErr()
					}
					localCacheExpireMs, err6 := strconv.Atoi(c.Val())
					if err6 != nil {
						redisService.localCacheExpireMs = int64(localCacheExpireMs)
					}
				default:
					if c.Val() != "}" {
						return &RedisService{}, c.Errf("Unknown config property '%s'", c.Val())
					}
				}
				if !c.Next() {
					break
				}
			}
		}
		redisService.Connect()
		return &redisService, nil
	}
	return &RedisService{}, nil
}

package redis

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/coredns/coredns/plugin"
	clog "github.com/coredns/coredns/plugin/pkg/log"
	"github.com/miekg/dns"

	redisCon "github.com/go-redis/redis/v7"
)

var log = clog.NewWithPlugin("coredns-redisc")

type Redis struct {
	Next           plugin.Handler
	ClusterClient  *redisCon.ClusterClient
	address        string
	password       string
	connectTimeout time.Duration
	readTimeout    time.Duration
	writeTimeout   time.Duration
	maxRetries     int
	poolSize       int
	ttl            uint32
	keyPrefix      string
	localCacheExpireMs int64
}

func InterfaceToArray(i interface{}) []string {
	var result []string
	if i == nil {
		return result
	}
	iArray := i.([]interface{})
	for _, elem := range iArray {
		result = append(result, elem.(string))
	}
	return result
}

func (redis *Redis) A(name string, z *Zone, record *Record) (answers, extras []dns.RR) {
	for _, a := range record.A {
		if a.Ip == nil {
			continue
		}
		r := new(dns.A)
		r.Hdr = dns.RR_Header{Name: dns.Fqdn(name), Rrtype: dns.TypeA,
			Class: dns.ClassINET, Ttl: redis.minTtl(a.Ttl)}
		r.A = a.Ip
		answers = append(answers, r)
	}
	return
}

func (redis Redis) AAAA(name string, z *Zone, record *Record) (answers, extras []dns.RR) {
	for _, aaaa := range record.AAAA {
		if aaaa.Ip == nil {
			continue
		}
		r := new(dns.AAAA)
		r.Hdr = dns.RR_Header{Name: dns.Fqdn(name), Rrtype: dns.TypeAAAA,
			Class: dns.ClassINET, Ttl: redis.minTtl(aaaa.Ttl)}
		r.AAAA = aaaa.Ip
		answers = append(answers, r)
	}
	return
}

func (redis *Redis) CNAME(name string, z *Zone, record *Record) (answers, extras []dns.RR) {
	for _, cname := range record.CNAME {
		if len(cname.Host) == 0 {
			continue
		}
		r := new(dns.CNAME)
		r.Hdr = dns.RR_Header{Name: dns.Fqdn(name), Rrtype: dns.TypeCNAME,
			Class: dns.ClassINET, Ttl: redis.minTtl(cname.Ttl)}
		r.Target = dns.Fqdn(cname.Host)
		answers = append(answers, r)
	}
	return
}

func (redis *Redis) TXT(name string, z *Zone, record *Record) (answers, extras []dns.RR) {
	for _, txt := range record.TXT {
		if len(txt.Text) == 0 {
			continue
		}
		r := new(dns.TXT)
		r.Hdr = dns.RR_Header{Name: dns.Fqdn(name), Rrtype: dns.TypeTXT,
			Class: dns.ClassINET, Ttl: redis.minTtl(txt.Ttl)}
		r.Txt = split255(txt.Text)
		answers = append(answers, r)
	}
	return
}

func (redis *Redis) NS(name string, z *Zone, record *Record) (answers, extras []dns.RR) {
	for _, ns := range record.NS {
		if len(ns.Host) == 0 {
			continue
		}
		r := new(dns.NS)
		r.Hdr = dns.RR_Header{Name: dns.Fqdn(name), Rrtype: dns.TypeNS,
			Class: dns.ClassINET, Ttl: redis.minTtl(ns.Ttl)}
		r.Ns = ns.Host
		answers = append(answers, r)
		extras = append(extras, redis.hosts(ns.Host, z)...)
	}
	return
}

func (redis *Redis) MX(name string, z *Zone, record *Record) (answers, extras []dns.RR) {
	for _, mx := range record.MX {
		if len(mx.Host) == 0 {
			continue
		}
		r := new(dns.MX)
		r.Hdr = dns.RR_Header{Name: dns.Fqdn(name), Rrtype: dns.TypeMX,
			Class: dns.ClassINET, Ttl: redis.minTtl(mx.Ttl)}
		r.Mx = mx.Host
		r.Preference = mx.Preference
		answers = append(answers, r)
		extras = append(extras, redis.hosts(mx.Host, z)...)
	}
	return
}

func (redis *Redis) SRV(name string, z *Zone, record *Record) (answers, extras []dns.RR) {
	for _, srv := range record.SRV {
		if len(srv.Target) == 0 {
			continue
		}
		r := new(dns.SRV)
		r.Hdr = dns.RR_Header{Name: dns.Fqdn(name), Rrtype: dns.TypeSRV,
			Class: dns.ClassINET, Ttl: redis.minTtl(srv.Ttl)}
		r.Target = srv.Target
		r.Weight = srv.Weight
		r.Port = srv.Port
		r.Priority = srv.Priority
		answers = append(answers, r)
		extras = append(extras, redis.hosts(srv.Target, z)...)
	}
	return
}

func (redis *Redis) SOA(name string, z *Zone, record *Record) (answers, extras []dns.RR) {
	r := new(dns.SOA)
	if record.SOA.Ns == "" {
		r.Hdr = dns.RR_Header{Name: dns.Fqdn(name), Rrtype: dns.TypeSOA,
			Class: dns.ClassINET, Ttl: redis.ttl}
		r.Ns = "ns1." + name
		r.Mbox = "hostmaster." + name
		r.Refresh = 86400
		r.Retry = 7200
		r.Expire = 3600
		r.Minttl = redis.ttl
	} else {
		r.Hdr = dns.RR_Header{Name: dns.Fqdn(z.Name), Rrtype: dns.TypeSOA,
			Class: dns.ClassINET, Ttl: redis.minTtl(record.SOA.Ttl)}
		r.Ns = record.SOA.Ns
		r.Mbox = record.SOA.MBox
		r.Refresh = record.SOA.Refresh
		r.Retry = record.SOA.Retry
		r.Expire = record.SOA.Expire
		r.Minttl = record.SOA.MinTtl
	}
	r.Serial = redis.serial()
	answers = append(answers, r)
	return
}

func (redis *Redis) CAA(name string, z *Zone, record *Record) (answers, extras []dns.RR) {
	if record == nil {
		return
	}
	for _, caa := range record.CAA {
		if caa.Value == "" || caa.Tag == "" {
			continue
		}
		r := new(dns.CAA)
		r.Hdr = dns.RR_Header{Name: dns.Fqdn(name), Rrtype: dns.TypeCAA, Class: dns.ClassINET}
		r.Flag = caa.Flag
		r.Tag = caa.Tag
		r.Value = caa.Value
		answers = append(answers, r)
	}
	return
}

func (redis *Redis) AXFR(z *Zone) (records []dns.RR) {
	//soa, _ := redis.SOA(z.Name, z, record)
	soa := make([]dns.RR, 0)
	answers := make([]dns.RR, 0, 10)
	extras := make([]dns.RR, 0, 10)

	// Allocate slices for rr Records
	records = append(records, soa...)
	for key := range z.Locations {
		if key == "@" {
			location := redis.findLocation(z.Name, z)
			record := redis.get(location, z)
			soa, _ = redis.SOA(z.Name, z, record)
		} else {
			fqdnKey := dns.Fqdn(key) + z.Name
			var as []dns.RR
			var xs []dns.RR

			location := redis.findLocation(fqdnKey, z)
			record := redis.get(location, z)

			// Pull all zone records
			as, xs = redis.A(fqdnKey, z, record)
			answers = append(answers, as...)
			extras = append(extras, xs...)

			as, xs = redis.AAAA(fqdnKey, z, record)
			answers = append(answers, as...)
			extras = append(extras, xs...)

			as, xs = redis.CNAME(fqdnKey, z, record)
			answers = append(answers, as...)
			extras = append(extras, xs...)

			as, xs = redis.MX(fqdnKey, z, record)
			answers = append(answers, as...)
			extras = append(extras, xs...)

			as, xs = redis.SRV(fqdnKey, z, record)
			answers = append(answers, as...)
			extras = append(extras, xs...)

			as, xs = redis.TXT(fqdnKey, z, record)
			answers = append(answers, as...)
			extras = append(extras, xs...)
		}
	}

	records = soa
	records = append(records, answers...)
	records = append(records, extras...)
	records = append(records, soa...)

	log.Debugf("Query AXFR of request: %s", records)
	return
}

func (redis *Redis) hosts(name string, z *Zone) []dns.RR {
	var (
		record  *Record
		answers []dns.RR
	)
	location := redis.findLocation(name, z)
	if location == "" {
		return nil
	}
	record = redis.get(location, z)
	a, _ := redis.A(name, z, record)
	answers = append(answers, a...)
	aaaa, _ := redis.AAAA(name, z, record)
	answers = append(answers, aaaa...)
	cname, _ := redis.CNAME(name, z, record)
	answers = append(answers, cname...)
	return answers
}

func (redis *Redis) serial() uint32 {
	return uint32(time.Now().Unix())
}

func (redis *Redis) minTtl(ttl uint32) uint32 {
	if ttl == 0 {
		return redis.ttl
	}
	if redis.ttl < ttl {
		return redis.ttl
	}
	return ttl
}

func (redis *Redis) findLocation(query string, z *Zone) string {
	var (
		ok                                 bool
		closestEncloser, sourceOfSynthesis string
	)

	// request for zone records
	if query == z.Name {
		return query
	}

	query = strings.TrimSuffix(query, "."+z.Name)

	if _, ok = z.Locations[query]; ok {
		return query
	}

	closestEncloser, sourceOfSynthesis, ok = splitQuery(query)
	for ok {
		ceExists := keyMatches(closestEncloser, z) || keyExists(closestEncloser, z)
		ssExists := keyExists(sourceOfSynthesis, z)
		if ceExists {
			if ssExists {
				return sourceOfSynthesis
			} else {
				return ""
			}
		} else {
			closestEncloser, sourceOfSynthesis, ok = splitQuery(closestEncloser)
		}
	}
	return ""
}

func (redis *Redis) get(key string, z *Zone) *Record {
	var label string
	if key == z.Name {
		label = "@"
	} else {
		label = key
	}
	r := z.Locations[label]
	return r
}

func keyExists(key string, z *Zone) bool {
	_, ok := z.Locations[key]
	return ok
}

func keyMatches(key string, z *Zone) bool {
	for value := range z.Locations {
		if strings.HasSuffix(value, key) {
			return true
		}
	}
	return false
}

func splitQuery(query string) (string, string, bool) {
	if query == "" {
		return "", "", false
	}
	var (
		splits            []string
		closestEncloser   string
		sourceOfSynthesis string
	)
	splits = strings.SplitAfterN(query, ".", 2)
	if len(splits) == 2 {
		closestEncloser = splits[1]
		sourceOfSynthesis = "*." + closestEncloser
	} else {
		closestEncloser = ""
		sourceOfSynthesis = "*"
	}
	return closestEncloser, sourceOfSynthesis, true
}

func (redis *Redis) Connect() {
	log.Infof("Connecting to redis cluster ... - for address: %s, password: %s, connectTimeout: %s, readTimeout: %s, writeTimeout: %s, maxRetries: %d, poolSize: %d, ttl: %d, keyPrefix: %s",
		redis.address,
		redis.password,
		redis.connectTimeout,
		redis.readTimeout,
		redis.writeTimeout,
		redis.maxRetries,
		redis.poolSize,
		redis.ttl,
		redis.keyPrefix,
	)

	_address := strings.Split(redis.address, ",")
	redis.ClusterClient = redisCon.NewClusterClient(&redisCon.ClusterOptions{
		Addrs:        _address,
		Password:     redis.password,
		DialTimeout:  redis.connectTimeout * time.Second,
		ReadTimeout:  redis.readTimeout * time.Second,
		WriteTimeout: redis.writeTimeout * time.Second,

		MaxRetries:   redis.maxRetries,
		PoolSize:     redis.poolSize,
	})

}

func (redis *Redis) save(zone string, subdomain string, value string) error {
	var err error

	conn := redis.ClusterClient
	if conn == nil {
		log.Error("Error connecting to redis")
		return nil
	}
	//defer conn.Close()

	err = conn.HSet(redis.keyPrefix+zone, subdomain, value).Err()
	return err
}

func (redis *Redis) load(zone string) *Zone {

	conn := redis.ClusterClient
	if conn == nil {
		log.Error("error connecting to redis")
		return nil
	}

	//Step1: Get from local cache (in localCacheExpireMs)
	bet := time.Now().UnixNano() / 1e6 - lastCacheTime[zone]
	log.Debugf("time=%d  localCacheExpireMs=%d",bet,redis.localCacheExpireMs)
	if bet < redis.localCacheExpireMs {
		z := localCache[zone]
		if z != nil {
			log.Infof("get from local cache: %s",z.Name)
			return z
		}
	}

	//Step2: Get from redis
	hGetAll,err := conn.HGetAll(redis.keyPrefix + zone).Result();

	//Step3: if get nil from redis,try get from local cache
	if err != nil || len(hGetAll) == 0{
		z := localCache[zone]
		if z != nil {
			log.Infof("get redis nil, get from local cache: %s",z.Name)
			return z
		}else{
			log.Info("get redis nil, get from local nil")
			return nil
		}
	}

	z := new(Zone)
	z.Name = zone
	z.Locations = make(map[string]*Record)
	for key, val := range hGetAll{
		log.Debugf("HGetAll from Redis: %s val:%s", key, val)
		r := new(Record)
		err := json.Unmarshal([]byte(val), r)
		if err != nil {
			log.Error("parse config error ", val, err)
			continue
		}
		z.Locations[key] = r
	}
	//save cache
	localCache[zone] = z
	lastCacheTime[zone] = time.Now().UnixNano() / 1e6
	log.Infof("get from redis : %s",z.Name)

	return z
}

func (redis *Redis) GetBlacklist() []string {
	conn := redis.ClusterClient
	if conn == nil {
		log.Error("error connecting to redis")
		return nil
	}
	_key := redis.getCacheKey(blacklistKeySuffix)
	smembers, err := conn.SMembers(_key).Result()
	if err != nil {
		log.Error("error get dns blacklist", err)
		return nil
	}
	log.Debugf("GetBlacklist: %s vals:%s", _key, smembers)
	return smembers
}

func (redis *Redis) GetWhitelist() []string {
	conn := redis.ClusterClient
	if conn == nil {
		log.Error("error connecting to redis")
		return nil
	}
	_key := redis.getCacheKey(whitelistKeySuffix)
	smembers, err := conn.SMembers(_key).Result()
	if err != nil {
		log.Error("Error get dns whitelist", err)
		return nil
	}
	log.Debugf("GetWhitelist: %s vals:%s", _key, smembers)
	return smembers
}

func (redis *Redis) getCacheKey(suffix string) string {
	return redis.keyPrefix + suffix
}

func split255(s string) []string {
	if len(s) < 255 {
		return []string{s}
	}
	sx := []string{}
	p, i := 0, 255
	for {
		if i <= len(s) {
			sx = append(sx, s[p:i])
		} else {
			sx = append(sx, s[p:])
			break
		}
		p, i = p+255, i+255
	}
	return sx
}

// Some special top-level domain names are defined here, because they have two
// levels of top-level names, which need special treatment when dealing with DNS query.
var SpecialDomains = [...]string{"com.cn.", "net.cn.", ".ac.cn.", ".org.cn.", ".gov.cn.", ".mil.cn.", ".edu.cn."}

var localCache map[string]*Zone = map[string]*Zone{}
var lastCacheTime map[string]int64 = map[string]int64{}

const (
	transferLength     = 1000
	blacklistKeySuffix = ":dns:blacklist"
	whitelistKeySuffix = ":dns:whitelist"
)

package redis

import (
	"encoding/json"
	"github.com/coredns/coredns/plugin"
	clog "github.com/coredns/coredns/plugin/pkg/log"
	"github.com/miekg/dns"
	"strings"
	"time"

	redisCon "github.com/go-redis/redis/v7"
)

var log = clog.NewWithPlugin("coredns-redisc")

type Redis struct {
	Next           plugin.Handler
	ClusterClient  *redisCon.ClusterClient
	redisAddress   string
	redisPassword  string
	connectTimeout int
	readTimeout    int
	keyPrefix      string
	keySuffix      string
	Ttl            uint32
	//Zones          []string
	LastZoneUpdate time.Time
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
			Class: dns.ClassINET, Ttl: redis.Ttl}
		r.Ns = "ns1." + name
		r.Mbox = "hostmaster." + name
		r.Refresh = 86400
		r.Retry = 7200
		r.Expire = 3600
		r.Minttl = redis.Ttl
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
	if redis.Ttl == 0 && ttl == 0 {
		return defaultTtl
	}
	if redis.Ttl == 0 {
		return ttl
	}
	if ttl == 0 {
		return redis.Ttl
	}
	if redis.Ttl < ttl {
		return redis.Ttl
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
	var (
		err error
		//reply interface{}
		val string
	)
	conn := redis.ClusterClient
	if conn == nil {
		log.Error("error connecting to redis")
		return nil
	}
	//defer conn.Close()

	var label string
	if key == z.Name {
		label = "@"
	} else {
		label = key
	}

	val = conn.HGet(redis.keyPrefix+z.Name+redis.keySuffix, label).Val()
	log.Debugf("HGet: %s label: %s val: %s", redis.keyPrefix+z.Name+redis.keySuffix, label, val)

	r := new(Record)
	err = json.Unmarshal([]byte(val), r)
	if err != nil {
		log.Error("parse config error ", val, err)
		return nil
	}
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
	log.Info("connecting to redis.........")

	redisAddress := strings.Split(redis.redisAddress, ",")
	redis.ClusterClient = redisCon.NewClusterClient(&redisCon.ClusterOptions{
		Addrs:        redisAddress,
		Password:     redis.redisPassword, // 设置密码
		DialTimeout:  5 * time.Second,     // 设置连接超时
		ReadTimeout:  5 * time.Second,     // 设置读取超时
		WriteTimeout: 5 * time.Second,     // 设置写入超时
		MaxRetries:   8,
		PoolSize:     4,
	})

}

func (redis *Redis) save(zone string, subdomain string, value string) error {
	var err error

	conn := redis.ClusterClient
	if conn == nil {
		log.Error("error connecting to redis")
		return nil
	}
	//defer conn.Close()

	err = conn.HSet(redis.keyPrefix+zone+redis.keySuffix, subdomain, value).Err()
	return err
}

func (redis *Redis) load(zone string) *Zone {

	conn := redis.ClusterClient
	if conn == nil {
		log.Error("error connecting to redis")
		return nil
	}
	//defer conn.Close()

	vals := conn.HKeys(redis.keyPrefix + zone + redis.keySuffix).Val()
	log.Infof("load HKEYS: %s vals:%s", redis.keyPrefix+zone+redis.keySuffix, vals)

	z := new(Zone)
	z.Name = zone
	z.Locations = make(map[string]struct{})
	for _, val := range vals {
		z.Locations[val] = struct{}{}
	}

	return z
}

func (redis *Redis) GetBlacklist() []string {
	conn := redis.ClusterClient
	if conn == nil {
		log.Error("error connecting to redis")
		return nil
	}
	smembers, err := conn.SMembers(blacklistKey).Result()
	if err != nil {
		log.Error("error get SMembers", err)
		return nil
	}
	log.Debugf("GetBlacklist: %s vals:%s", blacklistKey, smembers)
	return smembers
}

func (redis *Redis) GetWhitelist() []string {
	conn := redis.ClusterClient
	if conn == nil {
		log.Error("error connecting to redis")
		return nil
	}
	smembers, err := conn.SMembers(whitelistKey).Result()
	if err != nil {
		log.Error("error get SMembers", err)
		return nil
	}
	log.Debugf("GetWhitelist: %s vals:%s", whitelistKey, smembers)
	return smembers
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

//这里仅包含中国类别域名,其实还有行政区域名,中文域名,这里暂时不作处理
var SpecialDomains = [...]string{"com.cn.", "net.cn.", ".ac.cn.", ".org.cn.", ".gov.cn.", ".mil.cn.", ".edu.cn."}

const (
	defaultTtl     = 360
	transferLength = 1000
	blacklistKey   = "_dns_blacklist"
	whitelistKey   = "_dns_whitelist"
)

package coredns_gdns

import (
	"fmt"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
	"golang.org/x/net/context"
)

// ServeDNS implements the plugin.Handler interface.
func (redisService *RedisService) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := request.Request{W: w, Req: r}

	qname := state.Name()
	qtype := state.Type()

	zone := Qname2Zone(qname)

	//Blacklist use this way
	access := redisService.filter(qname)
	if !access {
		return dns.RcodeRefused, nil
	}

	if zone == "" {
		return plugin.NextOrFailure(qname, redisService.Next, ctx, w, r)
	}

	z := redisService.load(zone)
	if z == nil {
		return plugin.NextOrFailure(qname, redisService.Next, ctx, w, r)
	}

	if qtype == "AXFR" {
		records := redisService.AXFR(z)

		ch := make(chan *dns.Envelope)
		tr := new(dns.Transfer)
		tr.TsigSecret = nil

		go func(ch chan *dns.Envelope) {
			j, l := 0, 0

			for i, r := range records {
				l += dns.Len(r)
				if l > transferLength {
					ch <- &dns.Envelope{RR: records[j:i]}
					l = 0
					j = i
				}
			}
			if j < len(records) {
				ch <- &dns.Envelope{RR: records[j:]}
			}
			close(ch)
		}(ch)

		err := tr.Out(w, r, ch)
		if err != nil {
			fmt.Println(err)
		}
		w.Hijack()
		return dns.RcodeSuccess, nil
	}

	location := redisService.findLocation(qname, z)
	if len(location) == 0 { // empty, no results
		//return redisService.errorResponse(state, zone, dns.RcodeNameError, nil)
		return plugin.NextOrFailure(qname, redisService.Next, ctx, w, r)
	}

	answers := make([]dns.RR, 0, 10)
	extras := make([]dns.RR, 0, 10)

	record := redisService.get(location, z)
	if record == nil {
		return plugin.NextOrFailure(qname, redisService.Next, ctx, w, r)
	}

	switch qtype {
	case "A":
		answers, extras = redisService.A(qname, z, record)
	case "AAAA":
		answers, extras = redisService.AAAA(qname, z, record)
	case "CNAME":
		answers, extras = redisService.CNAME(qname, z, record)
	case "TXT":
		answers, extras = redisService.TXT(qname, z, record)
	case "NS":
		answers, extras = redisService.NS(qname, z, record)
	case "MX":
		answers, extras = redisService.MX(qname, z, record)
	case "SRV":
		answers, extras = redisService.SRV(qname, z, record)
	case "SOA":
		answers, extras = redisService.SOA(qname, z, record)
	case "CAA":
		answers, extras = redisService.CAA(qname, z, record)

	default:
		return redisService.errorResponse(state, zone, dns.RcodeNotImplemented, nil)
	}

	m := new(dns.Msg)
	m.SetReply(r)
	m.Authoritative, m.RecursionAvailable, m.Compress = true, false, true

	m.Answer = append(m.Answer, answers...)
	m.Extra = append(m.Extra, extras...)

	state.SizeAndDo(m)
	m = state.Scrub(m)
	_ = w.WriteMsg(m)
	return dns.RcodeSuccess, nil
}

// Name implements the Handler interface.
func (redisService *RedisService) Name() string { return PluginName }

func (redisService *RedisService) errorResponse(state request.Request, zone string, rcode int, err error) (int, error) {
	m := new(dns.Msg)
	m.SetRcode(state.Req, rcode)
	m.Authoritative, m.RecursionAvailable, m.Compress = true, false, true

	state.SizeAndDo(m)
	_ = state.W.WriteMsg(m)
	// Return success as the rcode to signal we have written to the client.
	return dns.RcodeSuccess, err
}

// blacklist and whitelist
func (redisService *RedisService) filter(qname string) bool {
	if len(qname) <= 0 {
		return false
	}
	qname = qname[0 : len(qname)-1]
	whitelistExp := redisService.GetWhitelist()
	blacklistExp := redisService.GetBlacklist()
	for _, expression := range whitelistExp {
		match := ExpressionMatch(qname, expression)
		if match {
			return true
		}
	}
	for _, expression := range blacklistExp {
		match := ExpressionMatch(qname, expression)
		if match {
			return false
		}
	}
	return true
}

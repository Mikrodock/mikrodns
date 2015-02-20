package dns

import (
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/mediocregopher/gobdns/config"
	"github.com/mediocregopher/gobdns/ips"
	"github.com/miekg/dns"
)

func suffixMatch(m *dns.Msg) string {
	addr := ""
	for _, q := range m.Question {
		for _, f := range config.ForwardSuffixes {
			if strings.HasSuffix(q.Name, f.Suffix) {
				if addr != "" && addr != f.ForwardAddr {
					return ""
				}
				addr = f.ForwardAddr
			}
		}
	}
	return addr
}

func doProxy(m *dns.Msg) *dns.Msg {
	var addr string
	if suffixAddr := suffixMatch(m); suffixAddr != "" {
		addr = suffixAddr
	} else if config.ForwardAddr != "" {
		addr = config.ForwardAddr
	} else {
		return nil
	}
	aM, err := dns.Exchange(m, addr)
	if err != nil {
		log.Println(err)
		return nil
	}
	return aM
}

func handleRequest(w dns.ResponseWriter, r *dns.Msg) {

	domain := r.Question[0].Name

	if ip, ok := ips.Get(domain); ok {

		// If the stored "ip" isn't actually an ip but a domain instead, we
		// proxy the request for that domain
		if net.ParseIP(ip) == nil {
			m := new(dns.Msg)
			m.SetQuestion(dns.Fqdn(ip), dns.TypeA)
			proxiedM := doProxy(m)
			if proxiedM == nil {
				dns.HandleFailed(w, r)
				return
			}

			cname, err := dns.NewRR(fmt.Sprintf("%s IN CNAME %s", domain, ip))
			if err != nil {
				log.Println(err)
				dns.HandleFailed(w, r)
				return
			}

			proxiedM.SetReply(r)
			proxiedM.Answer = append(proxiedM.Answer, cname)
			w.WriteMsg(proxiedM)
			return
		}

		a, err := dns.NewRR(fmt.Sprintf("%s IN A %s", domain, ip))
		if err != nil {
			log.Println(err)
			dns.HandleFailed(w, r)
			return
		}

		m := new(dns.Msg)
		m.SetReply(r)
		m.Answer = []dns.RR{a}
		w.WriteMsg(m)
		return
	}

	proxiedR := doProxy(r)
	if proxiedR == nil {
		dns.HandleFailed(w, r)
		return
	}
	w.WriteMsg(proxiedR)
}

func init() {
	handler := dns.HandlerFunc(handleRequest)
	if config.UDPAddr != "" {
		go func() {
			log.Printf("Listening on UDP %s", config.UDPAddr)
			err := dns.ListenAndServe(config.UDPAddr, "udp", handler)
			if err != nil {
				log.Fatal(err)
			}
		}()
	}
	if config.TCPAddr != "" {
		go func() {
			log.Printf("Listening on TCP %s", config.TCPAddr)
			err := dns.ListenAndServe(config.TCPAddr, "tcp", handler)
			if err != nil {
				log.Fatal(err)
			}
		}()
	}
}

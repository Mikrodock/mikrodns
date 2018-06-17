package dns

import (
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"gobdns/config"
	"gobdns/ips"

	"github.com/miekg/dns"
)

var timings map[string]*Timing

type Timing struct {
	start   time.Time
	end     time.Time
	Elapsed time.Duration
}

func (t *Timing) Start() {
	t.start = time.Now()
}

func (t *Timing) End() {
	if t.end.IsZero() {
		t.end = time.Now()
		t.Elapsed = t.end.Sub(t.start)
	}
}

func PrintTimings() {
	if len(timings) == 0 {
		log.Printf("No timings\n")
	}
	for name, time := range timings {
		if time.end.IsZero() {
			log.Printf("%s not closed\n", name)
		} else {
			log.Printf("%s took %s\n", name, time.Elapsed)
		}
	}
	timings = make(map[string]*Timing)
}

func StartTiming(name string) {
	timings[name] = &Timing{}
	timings[name].Start()
}

func EndTiming(name string) {
	if t, ok := timings[name]; ok {
		t.End()
	}
}

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
	log.Printf("Querying %s for IP of %s", addr, m.Question[0].Name)
	aM, err := dns.Exchange(m, addr)
	if err != nil {
		log.Println(err)
		return nil
	}
	return aM
}

func handleRequest(w dns.ResponseWriter, r *dns.Msg) {

	domain := r.Question[0].Name

	log.Println("Handling new request : " + domain + " for remote "+w.RemoteAddr().String())

	StartTiming("request")
	defer log.Println("===== END OF REQUEST =====")
	defer PrintTimings()
	defer EndTiming("request")

	StartTiming("ips.Get")
	if ip, ok := ips.Get(domain); ok {
		EndTiming("ips.Get")

		// If the stored "ip" isn't actually an ip but a domain instead, we
		// proxy the request for that domain
		StartTiming("net.ParseIP")
		if net.ParseIP(ip) == nil {

			log.Printf("Could not parse IP %s, it is a domain...\n", ip)

			EndTiming("net.ParseIP")
			m := new(dns.Msg)
			m.SetQuestion(dns.Fqdn(ip), dns.TypeA)
			log.Println("Proxying...")
			StartTiming("doProxy")
			proxiedM := doProxy(m)
			EndTiming("doProxy")
			if proxiedM == nil {
				dns.HandleFailed(w, r)
				return
			}
			StartTiming("proxy.NewRR")
			cname, err := dns.NewRR(fmt.Sprintf("%s IN CNAME %s", domain, ip))
			EndTiming("proxy.NewRR")
			if err != nil {
				log.Println(err)
				dns.HandleFailed(w, r)
				return
			}

			StartTiming("proxy.SendReply")
			proxiedM.SetReply(r)
			proxiedM.Answer = append(proxiedM.Answer, cname)
			w.WriteMsg(proxiedM)
			EndTiming("proxy.SendReply")
			return
		}
		EndTiming("net.ParseIP")

		StartTiming("NewRR")
		log.Println("Sending response for " + domain + " : " + ip)
		a, err := dns.NewRR(fmt.Sprintf("%s IN A %s", domain, ip))
		EndTiming("NewRR")
		if err != nil {
			log.Println(err)
			dns.HandleFailed(w, r)
			return
		}
		StartTiming("SendReply")
		m := new(dns.Msg)
		m.SetReply(r)
		m.Answer = []dns.RR{a}
		w.WriteMsg(m)
		EndTiming("SendReply")
		return
	}
	EndTiming("ips.Get")

	log.Println("No response for " + domain + ". Forwarding...")

	StartTiming("doProxy")
	proxiedR := doProxy(r)
	EndTiming("doProxy")
	if proxiedR == nil {
		dns.HandleFailed(w, r)
		return
	}
	StartTiming("proxy.SendReply")
	w.WriteMsg(proxiedR)
	EndTiming("proxy.SendReply")
}

func init() {

	timings = make(map[string]*Timing)

	handler := dns.HandlerFunc(handleRequest)
	if config.UDPAddr != "" {
		go func() {
			log.Printf("Listening on UDP (with timings) %s", config.UDPAddr)
			err := dns.ListenAndServe(config.UDPAddr, "udp", handler)
			if err != nil {
				log.Fatal(err)
			}
		}()
	}
	if config.TCPAddr != "" {
		go func() {
			log.Printf("Listening on TCP (with timings) %s", config.TCPAddr)
			err := dns.ListenAndServe(config.TCPAddr, "tcp", handler)
			if err != nil {
				log.Fatal(err)
			}
		}()
	}
}

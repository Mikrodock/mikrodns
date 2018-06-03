package ips

import (
	"fmt"
	"strings"
	"sync"
)

type DomainEntry struct {
	Resolvers map[string]int
}

type Mapping map[string]*DomainEntry

var m = Mapping{}
var rndDomains = make(map[string]*VoseRandom, 0)
var debug = make(map[string]int, 0)
var mLock sync.RWMutex

// Retrieves the best matching ip for the given domain. Or not. Will return the
// most specific match.
func Get(domain string) (string, bool) {
	for {
		ip, ok := GetExact(domain)

		if ok {

			if count, exist := debug[ip]; exist {
				debug[ip] = count + 1
			} else {
				debug[ip] = 1
			}

			fmt.Printf("%v\n", debug)

			return ip, true
		} else {
			return "", false
		}
	}
}

// Retrieves an ip exactly matching the given domain (although this will append
// the period to the end if necessary)
func GetExact(domain string) (string, bool) {
	mLock.RLock()
	defer mLock.RUnlock()
	_, ok := m[appendPeriod(domain)]
	if ok {
		ip := rndDomains[domain].Next()
		return ip, true
	} else {
		return "", false
	}

}

// Get's a copy of the map which maps all known domains to all known ips
func GetAll() Mapping {
	m2 := Mapping{}
	mLock.RLock()
	defer mLock.RUnlock()
	for domain, ip := range m {
		m2[domain] = ip
	}
	return m2
}

// Sets the given domain to point to the given ip. If the given domain doesn't
// end in a period one will be appended to it
func Set(domain, ip string, weight int) {
	if domain == "" {
		return
	}
	domain = appendPeriod(domain)

	mLock.Lock()
	if entry, ok := m[domain]; ok {
		entry.Resolvers[ip] = weight
	} else {

		entry := &DomainEntry{
			Resolvers: map[string]int{ip: weight},
		}
		m[domain] = entry
	}
	rndDomains[domain] = NewVoseRandom(FlattenProbs(m[domain].Resolvers))
	mLock.Unlock()
}

// Sets the current snapshot to a copy of the given one
func SetAll(m2 Mapping) {
	mLock.Lock()
	defer mLock.Unlock()
	m = Mapping{}
	for domain, ip := range m2 {
		m[domain] = ip
		rndDomains[domain] = NewVoseRandom(FlattenProbs(m[domain].Resolvers))
	}
}

// If the given domain is set to an ip, unsets it
func Unset(domain string, ip string) {
	domain = appendPeriod(domain)
	mLock.Lock()
	if entry, ok := m[domain]; ok {
		delete(entry.Resolvers, ip)
		if len(entry.Resolvers) == 0 {
			delete(m, domain)
			delete(rndDomains, domain)
		} else {
			rndDomains[domain] = NewVoseRandom(FlattenProbs(m[domain].Resolvers))
		}
	}
	mLock.Unlock()
}

func appendPeriod(domain string) string {
	if !strings.HasSuffix(domain, ".") {
		domain += "."
	}
	return domain
}

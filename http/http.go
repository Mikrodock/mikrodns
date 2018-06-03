package http

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"strconv"
	"strings"

	"gobdns/config"
	"gobdns/ips"
	"gobdns/snapshot"
)

func init() {
	if config.APIAddr == "" {
		return
	}

	go func() {
		log.Printf("API Listening on %s", config.APIAddr)
		http.HandleFunc("/api/domains/all", getAll)
		http.HandleFunc("/api/domains/", putDelete)
		http.HandleFunc("/api/snapshot", getSnapshot)
		http.ListenAndServe(config.APIAddr, nil)
	}()
}

func getAll(w http.ResponseWriter, r *http.Request) {
	for domain, ip := range ips.GetAll() {
		fmt.Fprintf(w, "%s %s\n", domain, ip)
	}
}

func putDelete(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/api/domains/" {
		w.WriteHeader(404)
		return
	}

	domain := path.Base(r.URL.Path)

	switch r.Method {
	case "PUT", "POST":
		body, err := ioutil.ReadAll(r.Body)
		stringBody := string(body)
		parts := strings.Split(stringBody, " ")
		if len(parts) != 2 {
			log.Println("Body", stringBody, "doesn't have the 2 required parts")
			w.WriteHeader(400)
			return
		}
		ip := parts[0]
		weight, err := strconv.Atoi(strings.TrimSpace(parts[1]))
		if err != nil {
			log.Println(err)
			w.WriteHeader(500)
			return
		}
		ip = strings.TrimSpace(ip)
		ips.Set(domain, ip, weight)

	case "DELETE":
		ipB, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Println(err)
			w.WriteHeader(500)
			return
		}
		ip := strings.TrimSpace(string(ipB))
		ips.Unset(domain, ip)
	}
}

func getIP(r *http.Request) string {
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return ip
	}
	portIdent := strings.LastIndex(r.RemoteAddr, ":")
	return r.RemoteAddr[:portIdent]
}

func getSnapshot(w http.ResponseWriter, r *http.Request) {
	b, err := snapshot.CreateEncoded()
	if err != nil {
		log.Printf("Creating snapshot: %s", err)
		w.WriteHeader(500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
	return
}

package main

import (
	_ "gobdns/dns"
	_ "gobdns/http"
	_ "gobdns/persist"
	_ "gobdns/repl"
)

func main() {
	select {}
}

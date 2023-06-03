package main

import (
	"mini-k8s/pkg/dns"
)

func main() {
	dns.AddHost("example.org", "10.0.0.1")
}

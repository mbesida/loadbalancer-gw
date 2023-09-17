package main

import (
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"
)

const port = 8000

var workers = []string{
	"http://localhost:9551",
	"http://localhost:9552",
	"http://localhost:9553",
}

func parseUrl(u string) url.URL {
	url, err := url.Parse(u)
	if err != nil {
		panic("Incorrect worker url. Chnage the configuration")
	}
	return *url
}

func main() {
	urls := make([]url.URL, len(workers))

	for i, u := range workers {
		urls[i] = parseUrl(u)
	}

	balancer := NewBalancer(urls)

	http.Handle("/get-fortune", balancer)

	l, err := net.Listen("tcp", ":"+strconv.Itoa(port))

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Load balancer has started on port %d and balancing among following %v workers\n", port, workers)

	log.Fatal(http.Serve(l, nil))
}

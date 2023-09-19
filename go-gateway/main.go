package main

import (
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const port = 8000

var workers = []string{
	"http://localhost:9551",
	"http://localhost:9552",
	"http://localhost:9553",
	"http://localhost:9554",
}

func setupLogging() {

}

func parseUrl(u string) url.URL {
	url, err := url.Parse(u)
	if err != nil {
		log.Fatalln("Incorrect worker url. Chnage the configuration")
	}
	return *url
}

func main() {
	setupLogging()

	var urls []url.URL = make([]url.URL, len(workers))

	for i, u := range workers {
		urls[i] = parseUrl(u)
	}

	balancer := NewBalancer(urls)

	http.Handle("/get-fortune", balancer)

	l, err := net.Listen("tcp", ":"+strconv.Itoa(port))

	if err != nil {
		log.Fatalln(err)
	}

	log.Printf("Load balancer has started on port %d and balancing among following workers:\n%s", port, strings.Join(workers, "\n"))

	log.Fatalln(http.Serve(l, nil))
}

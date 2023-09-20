package main

import (
	"log"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

const (
	Port     = 8000
	Endpoint = "/get-fortune"
)

var workers = []string{
	"http://localhost:9551",
	"http://localhost:9552",
	"http://localhost:9553",
}

func setupLogging() {
	l := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(l)
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

	balancer := NewBalancer2(urls)

	http.Handle(Endpoint, balancer)

	l, err := net.Listen("tcp", ":"+strconv.Itoa(Port))

	if err != nil {
		log.Fatalln(err)
	}

	log.Printf("Load balancer has started on port %d and balancing among following workers: %s",
		Port, strings.Join(workers, ", "))

	log.Fatalln(http.Serve(l, nil))
}

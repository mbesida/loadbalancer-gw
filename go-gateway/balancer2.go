package main

import (
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type Balancer2 struct {
	workers map[url.URL]*sync.Mutex
}

func NewBalancer2(workers []url.URL) *Balancer2 {
	b := &Balancer2{make(map[url.URL]*sync.Mutex, len(workers))}

	for _, v := range workers {
		b.workers[v] = &sync.Mutex{}
	}

	return b
}

func (b *Balancer2) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET methods are allowed", http.StatusMethodNotAllowed)
		return
	}

	b.loadBalance2(w, r, 0)
}

func (b *Balancer2) loadBalance2(w http.ResponseWriter, r *http.Request, attempt int) {
	attemptNumber := attempt + 1

	var proxyError error

	for url, mu := range b.workers {
		if mu.TryLock() {
			proxyError = handleWorkerRequest(w, r, url)
			mu.Unlock()
			if proxyError != nil {
				slog.Error(proxyError.Error())
				break
			}

			slog.Debug("Success", "attempt", attemptNumber, "worker", r.Host)
			return
		}
	}

	if proxyError != nil {
		slog.Debug("Http error happened", "client", r.RemoteAddr, "attempt", attemptNumber, "err", proxyError.Error())
	} else {
		slog.Debug("All workers are busy", "client", r.RemoteAddr, "attempt", attemptNumber)
	}

	time.Sleep(sleepDuration)
	b.loadBalance2(w, r, attemptNumber)
}

func handleWorkerRequest(w http.ResponseWriter, r *http.Request, target url.URL) error {
	r.Host = target.Host
	r.URL.Host = target.Host
	r.URL.Scheme = target.Scheme
	r.RequestURI = ""

	resp, err := http.DefaultClient.Do(r)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	w.WriteHeader(resp.StatusCode)

	_, err = io.Copy(w, resp.Body)

	return err
}

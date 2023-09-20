package main

import (
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"
)

const sleepDuration = 1500 * time.Millisecond

type worker struct {
	mu    sync.Mutex
	proxy httputil.ReverseProxy
}

func (w *worker) proxyRequest(wr http.ResponseWriter, r *http.Request) bool {
	isAcquired := w.mu.TryLock()
	if isAcquired {
		defer w.mu.Unlock()
		w.proxy.ServeHTTP(wr, r)
	}
	return isAcquired
}

type Balancer struct {
	workers map[url.URL]*worker
}

func NewBalancer(workers []url.URL) *Balancer {
	b := &Balancer{make(map[url.URL]*worker, len(workers))}

	for _, v := range workers {
		b.workers[v] = &worker{
			mu:    sync.Mutex{},
			proxy: makeProxy(v),
		}
	}

	return b
}

func (b *Balancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET methods are allowed", http.StatusMethodNotAllowed)
		return
	}

	b.loadBalance(w, r, 0)
}

func (b *Balancer) loadBalance(w http.ResponseWriter, r *http.Request, attempt int) {
	attemptNumber := attempt + 1
	for wh, worker := range b.workers {
		if worker.proxyRequest(w, r) {
			slog.Debug("Success", "attempt", attemptNumber, "worker", wh.Host)
			return
		}
	}
	slog.Debug("All workers are busy", "client", r.RemoteAddr, "attempt", attemptNumber)
	time.Sleep(sleepDuration)
	b.loadBalance(w, r, attemptNumber)
}

func makeProxy(u url.URL) httputil.ReverseProxy {
	return *httputil.NewSingleHostReverseProxy(&u)
}

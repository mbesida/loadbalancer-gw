package main

import (
	"log"
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
	for wh, worker := range b.workers {
		if worker.proxyRequest(w, r) {
			log.Printf("Success from %d attempt for worker %s\n", attempt+1, wh.Host)
			return
		}
	}
	log.Printf("All workers are busy. Client %s is waiting. Attempt %d\n", r.RemoteAddr, attempt+1)
	time.Sleep(sleepDuration)
	b.loadBalance(w, r, attempt+1)
}

func makeProxy(u url.URL) httputil.ReverseProxy {
	return *httputil.NewSingleHostReverseProxy(&u)
}

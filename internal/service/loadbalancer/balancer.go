package loadbalancer

import (
	"errors"
	"math/rand"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"
	"time"
)

var (
	ErrNoBackends = errors.New("no backends in param")
	ErrNoHealth   = errors.New("no health servers")
)

type Backend struct {
	URL         *url.URL
	Connections uint64
	Healthy     bool
	mu          sync.RWMutex
}

func (b *Backend) AddConnection() {
	atomic.AddUint64(&b.Connections, 1)

}

func (b *Backend) RemoveConnection() {
	atomic.AddUint64(&b.Connections, ^uint64(0))
}

type Strategy int

const (
	LeastConnections Strategy = iota
	RoundRobin
	Random
)

type Balancer struct {
	backends []*Backend
	mu       sync.RWMutex
	index    uint32
	strategy Strategy
	interval time.Duration
}

func NewLoadBalancer(backendAddrs []string, strat Strategy, interval time.Duration) (*Balancer, error) {
	if len(backendAddrs) == 0 {
		return &Balancer{}, ErrNoBackends
	}
	var bks []*Backend
	for _, addr := range backendAddrs {
		u, err := url.Parse(addr)
		if err != nil {
			return &Balancer{}, err
		}
		bks = append(bks, &Backend{URL: u, Healthy: false})
	}
	b := &Balancer{backends: bks, strategy: strat, interval: interval}
	go b.startHealthChecks()
	return b, nil
}

func (b *Balancer) Len() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.backends)
}

func (b *Balancer) startHealthChecks() {
	ticker := time.NewTicker(b.interval)
	defer ticker.Stop()
	for {
		<-ticker.C
		b.mu.RLock()
		wg := sync.WaitGroup{}
		for _, bk := range b.backends {
			wg.Add(1)
			go func(be *Backend) {
				defer wg.Done()
				httpClient := http.Client{Timeout: 2 * time.Second}
				resp, err := httpClient.Get(be.URL.String())
				healthy := err == nil && resp.StatusCode < 500
				if resp != nil {
					resp.Body.Close()
				}
				be.mu.Lock()
				be.Healthy = healthy
				be.mu.Unlock()
			}(bk)
		}
		b.mu.RUnlock()
		wg.Wait()
	}
}

func (b *Balancer) NextBackend() (*Backend, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	healthy := make([]*Backend, 0, len(b.backends))
	for _, bk := range b.backends {
		bk.mu.Lock()
		if bk.Healthy {
			healthy = append(healthy, bk)
		}
		bk.mu.Unlock()
	}
	if len(healthy) == 0 {
		return &Backend{}, ErrNoHealth
	}

	switch b.strategy {
	case RoundRobin:
		n := atomic.AddUint32(&b.index, 1)
		return healthy[(int(n))%len(healthy)], nil
	case Random:
		return healthy[rand.Intn(len(healthy))], nil
	case LeastConnections:
		fallthrough
	default:
		var sel = healthy[0]
		min := atomic.LoadUint64(&sel.Connections)
		for i := 1; i < len(healthy); i++ {
			bk := healthy[i]
			c := atomic.LoadUint64(&bk.Connections)
			if c < min {
				min = c
				sel = bk
			}
		}
		return sel, nil
	}
}

// A function is passed to the parameters, which will be passed to the new next server, or error.
func (b *Balancer) NextBackendWithWork(f func(host *url.URL, err error)) {
	backend, err := b.NextBackend()
	backend.AddConnection()
	defer backend.RemoveConnection()

	f(backend.URL, err)
}

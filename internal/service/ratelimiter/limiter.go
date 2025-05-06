package ratelimiter

import (
	"sync"
	"time"

	"github.com/TOMMy-Net/go-balancer/internal/domain/models"
	"github.com/jinzhu/copier"
)

// Required for initial initialization of limits in the config
const (
	defaultCapacity   = 100
	defaultRefillRate = 10
	defaultInterval   = time.Second
)

type BucketConfig struct {
	Capacity   int
	RefillRate int // Tokens to add each interval
	Tokens     int
}

type Bucket struct {
	tokens     int
	capacity   int
	refillRate int
	mu         sync.Mutex
}

type RateLimiter struct {
	buckets  map[string]*Bucket
	mu       sync.RWMutex
	ticker   *time.Ticker
	quit     chan struct{}
	quitOnce sync.Once
	config   *LimiterConfig
}

type LimiterConfig struct {
	DefaultInterval   time.Duration // Interval for bucket replenishment
	DefaultCapacity   int           // Default bucket capacity
	DefaultRefillRate int           // Default refill rate per interval
}

// CheckForData validates the LimiterConfig and applies defaults where needed.
func (l *LimiterConfig) CheckForData() {
	if l.DefaultInterval <= 0 {
		l.DefaultInterval = defaultInterval
	}
	if l.DefaultCapacity <= 0 {
		l.DefaultCapacity = defaultCapacity
	}
	if l.DefaultRefillRate <= 0 {
		l.DefaultRefillRate = defaultRefillRate
	}
}

func NewRateLimiter(c LimiterConfig) *RateLimiter {
	c.CheckForData()
	var limiter = &RateLimiter{
		buckets: make(map[string]*Bucket, 100),
		ticker:  time.NewTicker(c.DefaultInterval),
		quit:    make(chan struct{}),
		config:  &c,
	}
	go limiter.refillWorker()

	return limiter
}

func (rl *RateLimiter) Stop() {
	rl.quitOnce.Do(func() {
		close(rl.quit)
	})
}

// To add and update settings from the client
func (rl *RateLimiter) addClient(client string, cfg *BucketConfig) {
	var newBucket = new(Bucket)
	if cfg != nil {
		newBucket.capacity = cfg.Capacity
		newBucket.refillRate = cfg.RefillRate
		newBucket.tokens = cfg.Capacity
	} else {
		newBucket.capacity = rl.config.DefaultCapacity
		newBucket.refillRate = rl.config.DefaultRefillRate
		newBucket.tokens = rl.config.DefaultCapacity
	}

	rl.mu.Lock()
	rl.buckets[client] = newBucket
	rl.mu.Unlock()
}

func (rl *RateLimiter) AddClient(client string, m models.BucketConfig) {
	var target BucketConfig
	copier.Copy(&target, &m)
	rl.addClient(client, &target)
}

// Check if the client exists in the limit.
func (rl *RateLimiter) Check(client string) bool {
	rl.mu.RLock()
	defer rl.mu.RUnlock()
	_, ok := rl.buckets[client]
	return ok
}

// Checks if there is a client, if not, adds it with the usual settings and returns them
func (rl *RateLimiter) CheckAndAddDefault(client string) (bool, *models.BucketConfig) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	var m models.BucketConfig
	bucket, ok := rl.buckets[client]
	if !ok {
		bucket.capacity = rl.config.DefaultCapacity
		bucket.refillRate = rl.config.DefaultRefillRate
		bucket.tokens = rl.config.DefaultCapacity

		rl.buckets[client] = bucket
		copier.Copy(&m, &bucket)
	}
	return ok, &m
}

func (rl *RateLimiter) GetDefaultLimits() {

}

// Allow attempts to take one token for the given client.
// Returns true if allowed, false otherwise.
func (rl *RateLimiter) Allow(client string) bool {
	var bucket = &Bucket{}
	rl.mu.Lock()
	if bucket, exists := rl.buckets[client]; !exists {
		bucket.capacity = rl.config.DefaultCapacity
		bucket.refillRate = rl.config.DefaultRefillRate
		bucket.tokens = rl.config.DefaultCapacity

		rl.buckets[client] = bucket
	}
	rl.mu.Unlock()

	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	if bucket.tokens > 0 {
		bucket.tokens--
		return true
	}
	return false
}

func (rl *RateLimiter) refillWorker() {
	for {
		select {
		case <-rl.ticker.C:
			rl.refillAll()
		case <-rl.quit:
			rl.ticker.Stop()
			return
		}
	}
}

func (rl *RateLimiter) refillAll() {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	for _, bucket := range rl.buckets {
		go func(buc *Bucket) {
			buc.mu.Lock()
			buc.tokens += buc.refillRate
			if buc.tokens > buc.capacity {
				buc.tokens = buc.capacity
			}
			buc.mu.Unlock()
		}(bucket)
	}
}

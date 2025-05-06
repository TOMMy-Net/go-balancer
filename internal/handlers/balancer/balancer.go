package balancer

import (
	"context"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/TOMMy-Net/go-balancer/internal/domain/models"
	"github.com/TOMMy-Net/go-balancer/internal/render"
	"github.com/TOMMy-Net/go-balancer/internal/service/loadbalancer"
	"github.com/sirupsen/logrus"
)

type RateLimiter interface {
	Allow(client string) bool
	Stop()
	Check(client string) bool
	CheckAndAddDefault(client string) (bool, *models.BucketConfig)
	AddClient(client string, m models.BucketConfig)
}

type Database interface {
	AddClient(ctx context.Context, m *models.Client) error
}

// All fields must be filled in
type LoadBalancerHandler struct {
	LB          *loadbalancer.Balancer
	Logger      *logrus.Logger
	RateLimiter RateLimiter
	DB          Database
}

func NewLoadBalancerHandler(logger *logrus.Logger, backends []string, rl RateLimiter, healthInterval time.Duration) (*LoadBalancerHandler, error) {
	lb, err := loadbalancer.NewLoadBalancer(backends, loadbalancer.LeastConnections, healthInterval)
	if err != nil {
		return &LoadBalancerHandler{}, err
	}

	h := LoadBalancerHandler{
		LB:          lb,
		Logger:      logger,
		RateLimiter: rl,
	}
	return &h, nil
}

// For load-balancer server
func (lb *LoadBalancerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ok, config := lb.RateLimiter.CheckAndAddDefault(r.RemoteAddr)
	if !ok {
		err := lb.DB.AddClient(r.Context(), &models.Client{
			IP:              r.RemoteAddr,
			Capacity:        config.Capacity,
			RatePerInterval: config.RefillRate,
		})
		if err != nil {
			lb.Logger.WithFields(logrus.Fields{
				"error": err.Error(),
				"time":  time.Now().Format(time.RFC3339),
				"unix":  time.Now().Unix(),
				"addr":  r.RemoteAddr,
			}).Warn()
			render.SendError(w, http.StatusInternalServerError, &render.ErrorResponse{
				Error: ErrDatabase.Error(),
				Time:  time.Now().Format(time.RFC3339),
			})
			return
		}
	}
	if !lb.RateLimiter.Allow(r.RemoteAddr) {
		render.SendError(w, http.StatusTooManyRequests, &render.ErrorResponse{
			Error: ErrLimitEnd.Error(),
			Time:  time.Now().Format(time.RFC3339),
		})
		return
	}

	lb.LB.NextBackendWithWork(func(host *url.URL, err error) {
		if err != nil {
			timeNow := time.Now()
			lb.Logger.WithFields(logrus.Fields{
				"error": err.Error(),
				"time":  timeNow.Format(time.RFC3339),
				"unix":  timeNow.Unix(),
				"addr":  r.RemoteAddr,
			}).Warn()
			render.SendError(w, http.StatusServiceUnavailable, &render.ErrorResponse{
				Error: err.Error(),
				Time:  timeNow.Format(time.RFC3339),
			})
			return
		}

		proxy := httputil.NewSingleHostReverseProxy(host)
		proxy.ErrorHandler = func(w http.ResponseWriter, req *http.Request, err error) {
			lb.Logger.WithFields(logrus.Fields{
				"error":   err,
				"time":    time.Now().Format(time.RFC3339),
				"unix":    time.Now().Unix(),
				"backend": req.Host,
				"host":    r.Host,
				"code":    req.Response.StatusCode,
			}).Warn()
			render.SendError(w, http.StatusBadGateway, &render.ErrorResponse{
				Error: err.Error(),
				Time:  time.Now().Format(time.RFC3339),
			})
		}
		proxy.ServeHTTP(w, r)
	})
}

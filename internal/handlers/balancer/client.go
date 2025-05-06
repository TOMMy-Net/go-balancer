package balancer

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/TOMMy-Net/go-balancer/internal/domain/models"
	"github.com/TOMMy-Net/go-balancer/internal/render"
	"github.com/jinzhu/copier"
	"github.com/sirupsen/logrus"
)

type Client struct {
	IP              string `json:"ip"`
	Capacity        int    `json:"capacity"`
	RatePerInterval int    `json:"rate_per_interval"`
	Tokens          int    `json:"tokens"`
}

func (api *LoadBalancerHandler) Client() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			api.ClientAdd().ServeHTTP(w, r)
		case http.MethodDelete:
			api.RemoveClient().ServeHTTP(w, r)
		}
	}
}

func (api *LoadBalancerHandler) ClientAdd() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var c Client
		err := json.NewDecoder(r.Body).Decode(&c)
		if err != nil {
			t := time.Now()
			api.Logger.WithFields(logrus.Fields{
				"error":  err.Error(),
				"time":   t.Format(time.RFC3339),
				"addr":   r.RemoteAddr,
				"method": r.Method,
				"path":   r.URL.Path,
			})

			render.SendError(w, http.StatusBadRequest, &render.ErrorResponse{
				Error: err.Error(),
				Time:  t.Format(time.RFC3339),
			})
			return
		}
		var m models.Client
		copier.Copy(&m, &c)

		err = api.DB.AddClient(r.Context(), &m)
		if err != nil {
			api.Logger.WithFields(logrus.Fields{
				"error":  err.Error(),
				"time":   time.Now().Format(time.RFC3339),
				"addr":   r.RemoteAddr,
				"method": r.Method,
				"path":   r.URL.Path,
			})

			render.SendError(w, http.StatusInternalServerError, &render.ErrorResponse{
				Error: ErrDatabase.Error(),
				Time:  time.Now().Format(time.RFC3339),
			})
			return
		}
		var m1 models.BucketConfig
		m1.Capacity = m.Capacity
		m1.RefillRate = m.RatePerInterval
		m1.Tokens = m.Tokens
		api.RateLimiter.AddClient(r.RemoteAddr, m1)

		render.SendResponseJSON(w, http.StatusOK, &render.Response{
			Message: "client added",
			Time: time.Now().Format(time.RFC3339),
		})
	}
}

func (api *LoadBalancerHandler) RemoveClient() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

func (api *LoadBalancerHandler) UpdateLimits() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}

}

package balancer

import (
	"net/http"

)


// For API load-balancer server
func NewApiMux(api *LoadBalancerHandler) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/client", api.Client())
	return mux
}
package proxy

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type ServiceProxy struct {
	proxies map[string]*httputil.ReverseProxy
}

func New(services map[string]string) (*ServiceProxy, error) {
	proxies := make(map[string]*httputil.ReverseProxy, len(services))

	for name, addr := range services {
		target, err := url.Parse(addr)
		if err != nil {
			return nil, fmt.Errorf("parse service url %s (%s): %w", name, addr, err)
		}
		proxies[name] = httputil.NewSingleHostReverseProxy(target)
	}

	return &ServiceProxy{proxies: proxies}, nil
}

func (sp *ServiceProxy) Handler(service string) http.Handler {
	p, ok := sp.proxies[service]
	if !ok {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "service not found", http.StatusBadGateway)
		})
	}
	return p
}

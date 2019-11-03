package handlers

import (
	"net/http"
)

type MetricsHandler struct{}

func NewMetricsHandler() http.Handler {
	return &MetricsHandler{}
}

func (h *MetricsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
		return
	}
	errStatus := http.StatusMethodNotAllowed
	w.WriteHeader(errStatus)
	w.Write([]byte(http.StatusText(errStatus)))
	return
}

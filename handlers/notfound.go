package handlers

import (
	"net/http"
)

type NotFoundHandler struct{}

func NewNotFoundHandler() http.Handler {
	return &NotFoundHandler{}
}

func (h *NotFoundHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	errStatus := http.StatusNotFound
	w.WriteHeader(errStatus)
	w.Write([]byte(http.StatusText(errStatus)))
}

package handlers

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/negroni"
)

type Routes struct {
	LivenessHandler  http.Handler
	ReadinessHandler http.Handler
	MetricsHandler   http.Handler
	NotFoundHandler  http.Handler
}

func NewRoutes() *Routes {
	return &Routes{}
}

func (routes *Routes) Handler() http.Handler {
	r := mux.NewRouter().StrictSlash(true)
	r.Handle("/readiness", routes.ReadinessHandler)
	r.Handle("/health", routes.LivenessHandler)
	r.Handle("/metrics", routes.MetricsHandler)
	r.NotFoundHandler = routes.NotFoundHandler

	return ApplyMiddleware(r)
}

func ApplyMiddleware(h http.Handler) http.Handler {
	n := negroni.New()
	n.Use(negroni.HandlerFunc(loggerMiddleware))
	n.UseHandler(h)
	return n
}

func loggerMiddleware(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	start := time.Now()

	next(w, r)

	elapsed := time.Since(start)
	status := w.(negroni.ResponseWriter).Status()
	responseSize := w.(negroni.ResponseWriter).Size()

	contextLogger := log.WithFields(log.Fields{
		"requestMethod": r.Method,
		"requestUrl":    r.RequestURI,
		"requestSize":   r.ContentLength,
		"status":        status,
		"responseSize":  responseSize,
		"userAgent":     r.UserAgent(),
		"remoteIp":      r.RemoteAddr,
		"referer":       r.Referer(),
		"latency":       elapsed,
		"protocol":      r.Proto,
	})

	// e.g. INFO[0000] 200 GET /readiness in 44.421306ms
	switch {
	case status >= 400 && status < 500:
		contextLogger.Warnf("%d %s %s in %s", status, r.Method, r.URL, elapsed)
	case status >= 500:
		contextLogger.Errorf("%d %s %s in %s", status, r.Method, r.URL, elapsed)
	default:
		contextLogger.Infof("%d %s %s in %s", status, r.Method, r.URL, elapsed)
	}
}

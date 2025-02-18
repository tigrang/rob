package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"
)

type handler struct {
	mu             sync.Mutex
	app            *app
	proxy          *httputil.ReverseProxy
	connectTimeout time.Duration
	notifyRoute    string
	proxyBind      string
	proxyUrl       string
}

// newProxy creates a new proxy.
func newProxy(proxyBind string, notifyRoute string, connectTimeout time.Duration, app *app) (*handler, error) {
	remote, err := url.Parse("http://" + app.url)
	if err != nil {
		return nil, err
	}

	return &handler{
		proxy:          httputil.NewSingleHostReverseProxy(remote),
		proxyUrl:       app.url,
		proxyBind:      proxyBind,
		notifyRoute:    notifyRoute,
		connectTimeout: connectTimeout,
		app:            app,
	}, nil
}

// ServeHTTP is entry point for proxy. It handles notify route by marking app as dirty and proxies all other requests.
func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	slog.Info("Serving request", "url", r.URL)

	if r.URL.Path == h.notifyRoute {
		h.mu.Lock()
		h.app.markAsDirty()
		h.mu.Unlock()
		return
	}

	h.mu.Lock()
	err := h.app.rebuildIfDirty(h.connectTimeout)
	h.mu.Unlock()

	if err != nil {
		slog.Error(err.Error())
		h.respondWithError(w, err)
		return
	}

	slog.Info("Proxying request", "url", r.URL)
	h.proxy.ServeHTTP(w, r)
}

// notify proxy that there are changes in the code requiring a new build.
func (h *handler) notify() error {
	if err := waitForConnection(h.proxyBind, h.connectTimeout); err != nil {
		return fmt.Errorf("failed to connect to proxy: %w", err)
	}

	if _, err := http.Get("http://" + h.proxyBind + h.notifyRoute); err != nil {
		return fmt.Errorf("failed to notify proxy: %w", err)
	}

	return nil
}

// respondWithError renders and responds with error html page.
func (h *handler) respondWithError(w http.ResponseWriter, err error) {
	if err := tmpl.Execute(w, map[string]any{"lines": h.app.lines, "error": err.Error()}); err != nil {
		slog.Error("Failed to execute template", "err", err)
	}
}

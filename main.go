package main

import (
	"flag"
	"log/slog"
	"net/http"
	"time"
)

func main() {
	var notifyFlag bool
	var runFlag string
	var proxyUrlFlag string
	var proxyBindFlag string
	var notifyRouteFlag string
	var buildCmdFlag string
	var connectTimeoutFlag int
	var appPathFlag string

	flag.BoolVar(&notifyFlag, "notify", false, "notify proxy to trigger build")
	flag.StringVar(&runFlag, "run", "./run", "path to script that will run app")
	flag.StringVar(&proxyBindFlag, "proxybind", "localhost:9000", "the addr for error proxy to listen on")
	flag.StringVar(&notifyRouteFlag, "notifyroute", "/internal/build/notify", "path to trigger builds (must be the same when --notify is used)")
	flag.StringVar(&proxyUrlFlag, "proxy", "localhost:3000", "url app is listening on to forward requests")
	flag.StringVar(&buildCmdFlag, "build", "./build", "path to script that will build app")
	flag.IntVar(&connectTimeoutFlag, "timeout", 30, "the number of seconds to wait for proxy to be available")
	flag.StringVar(&appPathFlag, "path", "", "Path to app")
	flag.Parse()

	app := newApp(appPathFlag, proxyUrlFlag, runFlag, buildCmdFlag)

	h, err := newProxy(proxyBindFlag, notifyRouteFlag, time.Duration(connectTimeoutFlag)*time.Second, app)
	if err != nil {
		panic(err)
	}

	if notifyFlag {
		if err := h.notify(); err != nil {
			panic(err)
		}
		return
	}

	slog.Info("Listening on " + proxyBindFlag)
	slog.Info("Proxy to " + proxyUrlFlag)

	if err := http.ListenAndServe(proxyBindFlag, h); err != nil {
		panic(err)
	}
}

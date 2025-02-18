package main

import (
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/exec"
	"time"
)

type app struct {
	path         string
	url          string
	bin          string
	cmdPath      string
	buildOutput  string
	lastErr      error
	lines        []outputLine
	buildTime    time.Time
	startTime    time.Time
	lastModified time.Time
}

// newApp creates a new app.
func newApp(path string, url string, bin string, cmdPath string) *app {
	return &app{
		path:         path,
		url:          url,
		bin:          bin,
		cmdPath:      cmdPath,
		lastModified: time.Now(),
	}
}

// start execute app bin command and waits connectTimeout amount for it to be ready to accept connetions.
func (a *app) start(connectTimeout time.Duration) error {
	cmd := exec.Command(a.bin)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}

	if err := waitForConnection(a.url, connectTimeout); err != nil {
		return err
	}

	return nil
}

// rebuildIfDirty checks if build is out-of-date and runs build command then start the app.
func (a *app) rebuildIfDirty(connectTimeout time.Duration) error {
	if a.buildTime.Before(a.lastModified) {
		if err := a.build(); err != nil {
			fmt.Println(a.buildOutput)
			return err
		}
	}

	if a.lastErr != nil {
		return a.lastErr
	}

	if a.startTime.Before(a.buildTime) {
		if err := a.start(connectTimeout); err != nil {
			return fmt.Errorf("failed to start app: %w", err)
		}

		a.startTime = time.Now()
	}

	return nil
}

// markAsDirty sets the lastModified time to current time.
func (a *app) markAsDirty() {
	now := time.Now()
	if now.After(a.lastModified) {
		a.lastModified = now
	}
}

// build the app and parse output if there is an error.
func (a *app) build() error {
	a.buildTime = time.Now()

	a.lastErr = nil
	a.lines = nil
	cmd := exec.Command(a.cmdPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		a.lastErr = fmt.Errorf("build error: %s", err)
		a.buildOutput = string(output)
		a.lines = parse(a.buildOutput, a.path)
		return a.lastErr
	}

	return nil
}

// waitForConnection waits for connectTimeout amount of time for addr to be reachable.
func waitForConnection(addr string, connectTimeout time.Duration) error {
	slog.Info("Waiting for connection", "addr", addr, "connectTimeout", connectTimeout)

	timer := time.NewTimer(connectTimeout)
	defer timer.Stop()

	ticker := time.NewTicker(time.Millisecond * 500)
	defer ticker.Stop()

	for t := time.Tick(time.Second); ; {
		if _, err := net.Dial("tcp", addr); err == nil {
			return nil
		}

		select {
		case <-t:
			continue
		case <-timer.C:
			return errors.New("timed out waiting for connection")
		}
	}
}

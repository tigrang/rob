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
	url          string
	bin          string
	cmdPath      string
	buildOutput  []byte
	lastErr      error
	buildTime    time.Time
	startTime    time.Time
	lastModified time.Time
}

func newApp(url string, bin string, cmdPath string) *app {
	return &app{
		url:          url,
		bin:          bin,
		cmdPath:      cmdPath,
		lastModified: time.Now(),
	}
}

func (a *app) start(connectTimeout time.Duration) error {
	cmd := exec.Command(a.bin)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return err
	}

	if err := waitForConnection(a.url, connectTimeout); err != nil {
		return err
	}

	return nil
}

func (a *app) rebuildIfDirty(connectTimeout time.Duration) error {
	if a.buildTime.Before(a.lastModified) {
		if err := a.build(); err != nil {
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

func (a *app) markAsDirty() {
	now := time.Now()
	if now.After(a.lastModified) {
		a.lastModified = now
	}
}

func (a *app) build() error {
	a.buildTime = time.Now()

	a.lastErr = nil
	cmd := exec.Command(a.cmdPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		a.lastErr = fmt.Errorf("build error: %s", err)
		a.buildOutput = output
		return a.lastErr
	}

	return nil
}

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

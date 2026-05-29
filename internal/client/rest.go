package client

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"ftso-prices/internal/model"
)

var httpClient = &http.Client{Timeout: 15 * time.Second}

// GetJSON performs a GET request and decodes the JSON body into v.
func GetJSON(url string, v any) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("%s returned status %d", url, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, v)
}

// breaker is a simplified EventCountCircuitBreaker(5 events / 10s): it opens for
// 10s once more than 5 errors occur within a 10s window.
type breaker struct {
	mu        sync.Mutex
	events    []time.Time
	openUntil time.Time
}

const (
	breakerThreshold = 5
	breakerWindow    = 10 * time.Second
)

func (b *breaker) closed() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return time.Now().After(b.openUntil)
}

func (b *breaker) recordError() {
	b.mu.Lock()
	defer b.mu.Unlock()
	now := time.Now()
	cutoff := now.Add(-breakerWindow)
	kept := b.events[:0]
	for _, t := range b.events {
		if t.After(cutoff) {
			kept = append(kept, t)
		}
	}
	b.events = append(kept, now)
	if len(b.events) > breakerThreshold {
		b.openUntil = now.Add(breakerWindow)
		b.events = b.events[:0]
	}
}

// RestRunner polls a REST exchange on a fixed interval, skipping ticks while the
// circuit breaker is open.
type RestRunner struct {
	ex      RestExchange
	ctx     *Context
	push    func(model.Ticker)
	breaker breaker
	// gate, when set and returning true, skips a poll tick (used by the hybrid
	// runner to pause REST while the websocket feed is fresh).
	gate func() bool
}

func NewRestRunner(ex RestExchange, ctx *Context, push func(model.Ticker)) *RestRunner {
	return &RestRunner{ex: ex, ctx: ctx, push: push}
}

func (r *RestRunner) Name() string { return r.ex.Name() }

func (r *RestRunner) Run(stop <-chan struct{}) {
	ticker := time.NewTicker(r.ex.Interval())
	defer ticker.Stop()
	r.tick() // immediate first poll to seed prices at startup
	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			r.tick()
		}
	}
}

func (r *RestRunner) tick() {
	if r.gate != nil && r.gate() {
		return // websocket feed is fresh; skip the REST poll
	}
	if !r.breaker.closed() {
		return
	}
	if err := r.poll(); err != nil {
		log.Printf("Error polling %s: %v", r.ex.Name(), err)
		r.breaker.recordError()
	}
}

func (r *RestRunner) poll() (err error) {
	defer func() {
		if rec := recover(); rec != nil {
			err = fmt.Errorf("panic: %v", rec)
		}
	}()
	return r.ex.Poll(r.ctx, r.push)
}

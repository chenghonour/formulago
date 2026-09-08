/*
 * Copyright 2023 FormulaGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 */

// Package ratelimit provides an in-memory fixed-window rate limiter with
// failure lockout, used to slow down credential brute force. It keeps no
// per-key goroutines: stale entries are swept lazily on access.
package ratelimit

import (
	"sync"
	"time"
)

// Config tunes the limiter.
type Config struct {
	// MaxRequests allowed per client within one Window.
	MaxRequests int
	// Window over which MaxRequests and MaxFailures are counted.
	Window time.Duration
	// MaxFailures per Window before the client is locked out.
	MaxFailures int
	// Lockout duration once the request or failure budget is exceeded.
	Lockout time.Duration
}

type entry struct {
	windowStart time.Time
	count       int
	failStart   time.Time
	failCount   int
	lockUntil   time.Time
}

// Limiter is safe for concurrent use.
type Limiter struct {
	mu      sync.Mutex
	config  Config
	now     func() time.Time
	clients map[string]*entry
}

// New builds a Limiter with the given Config.
func New(cfg Config) *Limiter {
	if cfg.MaxRequests <= 0 {
		cfg.MaxRequests = 30
	}
	if cfg.Window <= 0 {
		cfg.Window = time.Minute
	}
	if cfg.MaxFailures <= 0 {
		cfg.MaxFailures = 10
	}
	if cfg.Lockout <= 0 {
		cfg.Lockout = 30 * time.Minute
	}
	return &Limiter{
		config:  cfg,
		now:     time.Now,
		clients: make(map[string]*entry),
	}
}

// Allow reports whether a request from key may proceed now and records that
// one request. Call Report(key, ok) afterwards with the request outcome.
func (l *Limiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	e := l.entryFor(key)
	if e.lockUntil.After(l.now()) {
		return false
	}
	e.count++
	if e.count > l.config.MaxRequests {
		e.lockUntil = l.now().Add(l.config.Lockout)
		return false
	}
	return true
}

// Report feeds the request outcome back: a success clears the failure streak,
// a failure (e.g. HTTP 401 on login) advances it towards lockout.
func (l *Limiter) Report(key string, succeeded bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	e := l.entryFor(key)
	now := l.now()
	if e.lockUntil.After(now) {
		return
	}
	if succeeded {
		e.failCount = 0
		return
	}
	if now.Sub(e.failStart) > l.config.Window {
		e.failStart = now
		e.failCount = 0
	}
	e.failCount++
	if e.failCount >= l.config.MaxFailures {
		e.lockUntil = now.Add(l.config.Lockout)
	}
}

// entryFor returns the entry for key, resetting expired windows, and drops
// entries that have been idle longer than two windows to bound memory.
func (l *Limiter) entryFor(key string) *entry {
	now := l.now()
	e, ok := l.clients[key]
	if !ok {
		e = &entry{windowStart: now, failStart: now}
		l.clients[key] = e
	} else {
		if now.Sub(e.windowStart) > l.config.Window {
			e.windowStart = now
			e.count = 0
		}
		if now.Sub(e.failStart) > l.config.Window {
			e.failStart = now
			e.failCount = 0
		}
	}
	for k, v := range l.clients {
		if v != e && now.Sub(v.windowStart) > 2*l.config.Window &&
			now.Sub(v.failStart) > 2*l.config.Window &&
			!v.lockUntil.After(now) {
			delete(l.clients, k)
		}
	}
	return e
}

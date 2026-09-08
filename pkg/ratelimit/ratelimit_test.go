/*
 * Copyright 2023 FormulaGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 */

package ratelimit

import (
	"testing"
	"time"
)

// clock holds a mutable "now" so tests can advance time for the limiter.
type clock struct{ t time.Time }

func newTestLimiter(c *clock) *Limiter {
	l := New(Config{
		MaxRequests: 3,
		Window:      time.Minute,
		MaxFailures: 2,
		Lockout:     time.Hour,
	})
	l.now = func() time.Time { return c.t }
	return l
}

func TestAllowBurstThenLockout(t *testing.T) {
	c := &clock{t: time.Now()}
	l := newTestLimiter(c)

	for i := range 3 {
		if !l.Allow("1.2.3.4") {
			t.Fatalf("request %d should be allowed", i+1)
		}
	}
	// the 4th request within the window exceeds the budget and locks out
	if l.Allow("1.2.3.4") {
		t.Fatal("request beyond MaxRequests should be denied")
	}
	// lockout lasts until Lockout passes, even after the window rolls
	c.t = c.t.Add(2 * time.Minute)
	if l.Allow("1.2.3.4") {
		t.Fatal("request during lockout should be denied")
	}
	c.t = c.t.Add(2 * time.Hour)
	if !l.Allow("1.2.3.4") {
		t.Fatal("request after lockout should be allowed again")
	}
}

func TestFailureLockoutAndRecovery(t *testing.T) {
	c := &clock{t: time.Now()}
	l := newTestLimiter(c)

	// two failures reach MaxFailures -> lockout
	l.Report("1.2.3.4", false)
	l.Report("1.2.3.4", false)
	if l.Allow("1.2.3.4") {
		t.Fatal("request after MaxFailures should be denied")
	}

	// a success before the threshold resets the failure streak
	c2 := &clock{t: time.Now()}
	l2 := newTestLimiter(c2)
	l2.Report("1.2.3.4", false)
	l2.Report("1.2.3.4", true)
	l2.Report("1.2.3.4", false)
	if !l2.Allow("1.2.3.4") {
		t.Fatal("success should reset the failure streak")
	}
}

func TestWindowsRollIndependentPerKey(t *testing.T) {
	c := &clock{t: time.Now()}
	l := newTestLimiter(c)
	if !l.Allow("a") || !l.Allow("b") {
		t.Fatal("different keys must not affect each other")
	}
	// key a exhausts its window budget
	for range 3 {
		l.Allow("a")
	}
	if l.Allow("a") {
		t.Fatal("key a should be over its budget")
	}
	if !l.Allow("b") {
		t.Fatal("key b must keep its own budget")
	}
}

func TestFailureWindowRollsWithoutLockout(t *testing.T) {
	c := &clock{t: time.Now()}
	l := newTestLimiter(c)
	// one failure is below MaxFailures, so no lockout
	l.Report("k", false)
	if !l.Allow("k") {
		t.Fatal("single failure must not lock the client out")
	}
	// after the window rolls the failure count resets, another single failure
	// still does not lock out
	c.t = c.t.Add(2 * time.Minute)
	l.Report("k", false)
	if !l.Allow("k") {
		t.Fatal("failure streak should reset after the window")
	}
}

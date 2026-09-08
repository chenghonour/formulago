/*
 * Copyright 2023 FormulaGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 */

package configs

import "testing"

func TestAccessSecretProblem(t *testing.T) {
	tests := []struct {
		name   string
		secret string
		want   bool // true = a problem was reported
	}{
		{name: "empty", secret: "", want: true},
		{name: "known placeholder", secret: "change-me", want: true},
		{name: "too short", secret: "0123456789abcdef", want: true},
		{name: "exactly 32 bytes", secret: "0123456789abcdef0123456789abcdef", want: false},
		{name: "long enough random-looking", secret: "this-is-a-32+byte-random-signing-key-0123456789", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{}
			cfg.IsProd = true
			cfg.Auth.AccessSecret = tt.secret
			got := accessSecretProblem(cfg) != ""
			if got != tt.want {
				t.Errorf("accessSecretProblem(prod, %q) problem=%v, want %v", tt.secret, got, tt.want)
			}
			// dev mode must never be a hard failure: only warning levels differ
			cfg.IsProd = false
			if got := accessSecretProblem(cfg) != ""; got != tt.want {
				t.Errorf("accessSecretProblem(dev, %q) problem=%v, want %v", tt.secret, got, tt.want)
			}
		})
	}
}

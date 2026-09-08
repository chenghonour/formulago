/*
 * Copyright 2023 FormulaGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 */

package encrypt

import (
	"testing"
)

func TestBcryptCheck(t *testing.T) {
	type args struct {
		password string
		hash     string
	}
	arg := args{
		password: "123456",
		hash:     "$2a$10$RKtc2mfcoc.Op5hJALKYGO9Qw86z20NpWFxhnZLddWjGQRNfKqK7G",
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "BcryptCheck", args: arg, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BcryptCheck(tt.args.password, tt.args.hash); got != tt.want {
				t.Errorf("BcryptCheck() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBcryptEncrypt(t *testing.T) {
	type args struct {
		password string
	}
	arg := args{
		password: "123456",
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "BcryptEncrypt", args: arg, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encryptStr, err := BcryptEncrypt(tt.args.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("BcryptEncrypt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			t.Log(encryptStr)
		})
	}
}

// TestSha256Hex uses NIST test vectors (FIPS 180-2 / RFC 6234).
func TestSha256Hex(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "empty",
			input: "",
			want:  "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
		{
			name:  "abc",
			input: "abc",
			want:  "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Sha256Hex(tt.input); got != tt.want {
				t.Errorf("Sha256Hex(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

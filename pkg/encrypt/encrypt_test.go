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
			_, err := BcryptEncrypt(tt.args.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("BcryptEncrypt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

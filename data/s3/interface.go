/*
 * Copyright 2023 FormulaGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 */

// Package s3 provides a client for Simple Storage Service.

package s3

import (
	"context"
	"io"
)

// FileInfo file store info
type FileInfo struct {
	Name   string     `json:"fileName"`
	Size   int64      `json:"fileSize"`
	Path   string     `json:"filePath"`
	Type   string     `json:"fileType"`
	Reader *io.Reader `json:"fileReader"`
}

// Adapter file store adapter
type Adapter interface {
	UploadFile(ctx context.Context, file *FileInfo) (fileInfo *FileInfo, err error)
	SignedUrl(ctx context.Context, filePath string) (signedURL string, err error)
	DeleteFile(ctx context.Context, filePath []string) (deletedObjects []string, err error)
	FileReader(ctx context.Context, filePath string) (fileReader io.ReadCloser, err error)
}

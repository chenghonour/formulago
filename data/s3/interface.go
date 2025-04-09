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
	// FileName
	Name string `json:"fileName"`
	// FileSize file size in bytes
	Size int64 `json:"fileSize"`
	// File path, no need bucket name(have been set in config), no need / prefix
	Path string `json:"filePath"`
	// File type, eg: img/file
	Type string `json:"fileType"`
	// FileReader
	Reader io.Reader `json:"fileReader"`
}

type ObjectProperties struct {
	// The name of the object.
	Key *string `xml:"Key"`
	// The type of the object. Valid values: Normal, Multipart and Appendable
	Type *string `xml:"Type"`
	// The size of the returned object. Unit: bytes.
	Size int64 `xml:"Size"`
	// The entity tag (ETag). An ETag is created when an object is created to identify the content of the object.
	ETag *string `xml:"ETag"`
	// The time when the returned objects were last modified.
	LastModified string `xml:"LastModified"`
	// The storage class of the object. For example：Standard, IA, Archive, ColdArchive and DeepColdArchive.
	// Get a not Standard storage class object need to restore first
	StorageClass *string `xml:"StorageClass"`
	// The restoration status of the object.
	RestoreInfo *string `xml:"RestoreInfo"`
	// The time when the storage class of the object is converted to Cold Archive or Deep Cold Archive based on lifecycle rules.
	TransitionTime string `xml:"TransitionTime"`
}

// Adapter file store adapter
type Adapter interface {
	UploadFile(ctx context.Context, file *FileInfo) (fileInfo *FileInfo, err error)
	SignedUrl(ctx context.Context, filePath string) (signedURL string, err error)
	DeleteFile(ctx context.Context, filePath []string) (deletedObjects []string, err error)
	FileReader(ctx context.Context, filePath string) (fileReader io.ReadCloser, err error)
	ListObjects(ctx context.Context, prefix string) (results []ObjectProperties, err error)
	// RestoreObject 解冻对象, restoreDays 文件解冻后持续天数,例如3天
	// 建议先调用ListObjects获取文件StorageClass，如非Standard才需要解冻
	// 与查询解冻状态, 若为解冻状态则无需调用此方法
	RestoreObject(ctx context.Context, filePath string, restoreDays int) (err error)
}

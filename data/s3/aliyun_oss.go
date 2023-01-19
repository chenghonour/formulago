/*
 * Copyright 2022 FormulaGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 */

package s3

import (
	"context"
	"fmt"
	"formulago/configs"
	"formulago/pkg/img"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/cockroachdb/errors"
	"io"
	"path/filepath"
	"strings"
)

func NewUploadAliyunOSSAdapter(c configs.Config) *UploadAliyunOSSAdapter {
	return &UploadAliyunOSSAdapter{
		Endpoint:   c.S3.AliyunOSS.Endpoint,
		BucketName: c.S3.AliyunOSS.BucketName,
		BasePath:   c.S3.AliyunOSS.BasePath,
		SecretID:   c.S3.AliyunOSS.SecretID,
		SecretKey:  c.S3.AliyunOSS.SecretKey,
		Domain:     c.S3.AliyunOSS.Domain,
	}
}

type UploadAliyunOSSAdapter struct {
	Endpoint   string
	SecretID   string
	SecretKey  string
	BucketName string
	BasePath   string
	Domain     string
}

func (u *UploadAliyunOSSAdapter) UploadFile(ctx context.Context, file *FileInfo) (fileInfo *FileInfo, err error) {
	if file.Reader == nil {
		err = errors.New("not upload any file")
		return
	}

	file.Path, err = u.upAction(ctx, file)
	return file, err
}

// SignedUrl get signed url of file, default expire time is 30 minutes
func (u *UploadAliyunOSSAdapter) SignedUrl(ctx context.Context, filePath string) (signedURL string, err error) {
	// create oss client
	client, err := oss.New(u.Endpoint, u.SecretID, u.SecretKey)
	if err != nil {
		return
	}
	// get store bucket
	bucket, err := client.Bucket(u.BucketName)
	if err != nil {
		return
	}
	// get signed url
	signedURL, err = bucket.SignURL(filePath, oss.HTTPGet, 1800)
	signedURL = strings.Join(strings.Split(signedURL, "/")[3:], "/")
	signedURL = "https://" + u.Domain + "/" + signedURL

	return
}

// FileReader get file io reader
func (u *UploadAliyunOSSAdapter) FileReader(ctx context.Context, filePath string) (fileReader io.ReadCloser, err error) {
	// create oss client
	client, err := oss.New(u.Endpoint, u.SecretID, u.SecretKey)
	if err != nil {
		return
	}
	// get store bucket
	bucket, err := client.Bucket(u.BucketName)
	if err != nil {
		return
	}
	// get signed url
	fileReader, err = bucket.GetObject(filePath)

	return
}

// DeleteFile delete file in s3
func (u *UploadAliyunOSSAdapter) DeleteFile(ctx context.Context, filePath []string) (deletedObjects []string, err error) {
	// create oss client
	client, err := oss.New(u.Endpoint, u.SecretID, u.SecretKey)
	if err != nil {
		return
	}
	// get store bucket
	bucket, err := client.Bucket(u.BucketName)
	if err != nil {
		return
	}
	// get signed url
	r, err := bucket.DeleteObjects(filePath)
	deletedObjects = r.DeletedObjects
	return
}

// upAction, upload file to aliyun oss
func (u *UploadAliyunOSSAdapter) upAction(ctx context.Context, file *FileInfo) (path string, err error) {
	// <yourObjectName>upload file to OSS need specify the full path of file, like abc/efg/123.jpg
	objectName := u.getPathInBucket(ctx, file)
	// create oss client
	client, err := oss.New(u.Endpoint, u.SecretID, u.SecretKey)
	if err != nil {
		return
	}
	// get store bucket
	bucket, err := client.Bucket(u.BucketName)
	if err != nil {
		return
	}
	// upload file
	// image should be compressed
	if file.Type == "img" && file.Size > 1000*1000 {
		format := filepath.Ext(file.Name)
		fBuffer, err := img.Compress(*file.Reader, format)
		if err != nil {
			return "", err
		}
		file.Reader = &fBuffer
	}
	err = bucket.PutObject(objectName, *file.Reader)
	if err != nil {
		return
	}

	path = objectName
	return
}

// get the full path of file in bucket
func (u *UploadAliyunOSSAdapter) getPathInBucket(ctx context.Context, file *FileInfo) string {
	url := fmt.Sprintf("%s/%s/%s", u.BasePath, file.Path, file.Name)
	return url
}

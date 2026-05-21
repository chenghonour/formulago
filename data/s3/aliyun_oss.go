/*
 * Copyright 2023 FormulaGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 */

package s3

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"formulago/configs"
	"formulago/pkg/times"
	"formulago/pkg/img"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
)

func NewAliyunOSSAdapter(c configs.Config) *AliyunOSSAdapter {
	return &AliyunOSSAdapter{
		Region:     c.S3.AliyunOSS.Region,
		Endpoint:   c.S3.AliyunOSS.Endpoint,
		BucketName: c.S3.AliyunOSS.BucketName,
		BasePath:   c.S3.AliyunOSS.BasePath,
		SecretID:   c.S3.AliyunOSS.SecretID,
		SecretKey:  c.S3.AliyunOSS.SecretKey,
		Domain:     c.S3.AliyunOSS.Domain,
	}
}

type AliyunOSSAdapter struct {
	Region     string
	Endpoint   string
	SecretID   string
	SecretKey  string
	BucketName string
	BasePath   string
	Domain     string

	clientOnce sync.Once
	client     *oss.Client
}

func (a *AliyunOSSAdapter) ossClient() *oss.Client {
	a.clientOnce.Do(func() {
		provider := credentials.NewStaticCredentialsProvider(a.SecretID, a.SecretKey)
		cfg := oss.LoadDefaultConfig().
			WithRegion(a.Region).
			WithEndpoint(a.Endpoint).
			WithCredentialsProvider(provider).
			WithConnectTimeout(20 * time.Second).
			WithReadWriteTimeout(60 * time.Second).
			WithLogLevel(oss.LogError)
		a.client = oss.NewClient(cfg)
	})
	return a.client
}

func (a *AliyunOSSAdapter) UploadFile(ctx context.Context, file *FileInfo) (fileInfo *FileInfo, err error) {
	if file.Reader == nil {
		err = errors.New("not upload any file")
		return
	}

	file.URL, err = a.upAction(ctx, file)
	return file, err
}

// SignedUrl get signed url of file, default expire time is 30 minutes
// if customized domain is not empty, signed url will be joined with domain
func (a *AliyunOSSAdapter) SignedUrl(ctx context.Context, filePath string) (signedURL string, err error) {
	client := a.ossClient()

	// 生成GetObject的预签名URL
	result, err := client.Presign(ctx, &oss.GetObjectRequest{
		Bucket: oss.Ptr(a.BucketName),
		Key:    oss.Ptr(filePath),
	},
		oss.PresignExpires(15*time.Minute),
	)
	if err != nil {
		return
	}
	signedURL = result.URL
	if a.Domain != "" {
		signedURL = strings.Join(strings.Split(signedURL, "/")[3:], "/")
		signedURL = "https://" + a.Domain + "/" + signedURL
	}

	return
}

func (a *AliyunOSSAdapter) ListObjects(ctx context.Context, prefix string) (results []ObjectProperties, err error) {
	client := a.ossClient()

	// 创建列出对象的请求
	request := &oss.ListObjectsV2Request{
		Bucket: oss.Ptr(a.BucketName),
		Prefix: oss.Ptr(prefix),
	}

	// 创建分页器
	p := client.NewListObjectsV2Paginator(request)

	var listObjectsV2Result []*oss.ListObjectsV2Result
	var i int
	// 遍历分页器中的每一页
	for p.HasNext() {
		i++
		// 获取下一页的数据
		page, err := p.NextPage(ctx)
		if err != nil {
			return results, err
		}
		listObjectsV2Result = append(listObjectsV2Result, page)

		// 该页中的每个对象的信息
		for _, obj := range page.Contents {
			r := ObjectProperties{
				Key:          obj.Key,
				Type:         obj.Type,
				Size:         obj.Size,
				ETag:         obj.ETag,
				StorageClass: obj.StorageClass,
				RestoreInfo:  obj.RestoreInfo,
			}
			if obj.LastModified != nil {
				r.LastModified = obj.LastModified.Format(times.TimeFormat)
			}
			results = append(results, r)
		}
	}

	return results, nil
}

func (a *AliyunOSSAdapter) RestoreObject(ctx context.Context, filePath string, restoreDays int) (err error) {
	client := a.ossClient()

	// 创建解冻对象的请求
	request := &oss.RestoreObjectRequest{
		Bucket: oss.Ptr(a.BucketName),
		Key:    oss.Ptr(filePath),
		RestoreRequest: &oss.RestoreRequest{
			Days: int32(restoreDays),
		},
	}

	// 发送解冻对象的请求
	_, err = client.RestoreObject(ctx, request)
	if err != nil {
		return err
	}
	return nil
}

// FileReader get file io reader
func (a *AliyunOSSAdapter) FileReader(ctx context.Context, filePath string) (fileReader io.ReadCloser, err error) {
	client := a.ossClient()

	// get store file object
	result, err := client.GetObject(ctx, &oss.GetObjectRequest{
		Bucket: oss.Ptr(a.BucketName),
		Key:    oss.Ptr(filePath),
	})
	if err != nil {
		return
	}
	fileReader = result.Body

	return
}

// DeleteFile delete file in s3
func (a *AliyunOSSAdapter) DeleteFile(ctx context.Context, filePath []string) (deletedFilePath []string, err error) {
	client := a.ossClient()

	// delete file
	var deleteObjects []oss.DeleteObject
	for _, v := range filePath {
		deleteObjects = append(deleteObjects, oss.DeleteObject{
			Key: oss.Ptr(v),
		})
	}
	result, err := client.DeleteMultipleObjects(ctx, &oss.DeleteMultipleObjectsRequest{
		Bucket:  oss.Ptr(a.BucketName),
		Objects: deleteObjects,
	})
	if err != nil {
		return
	}

	// get deleted file path
	for _, v := range result.DeletedObjects {
		deletedFilePath = append(deletedFilePath, *v.Key)
	}
	return
}

// upAction, upload file to aliyun oss
func (a *AliyunOSSAdapter) upAction(ctx context.Context, file *FileInfo) (URL string, err error) {
	// <yourObjectName>upload file to OSS need specify the full path of file, like abc/efg/123.jpg
	objectName := a.getPathInBucket(ctx, file)
	client := a.ossClient()

	// upload file
	// image should be compressed
	if file.Type == "img" && file.Size > 1000*1000 {
		format := filepath.Ext(file.Name)
		fBuffer, err := img.Compress(file.Reader, format)
		if err != nil {
			return "", err
		}
		file.Reader = fBuffer
	}
	_, err = client.PutObject(ctx, &oss.PutObjectRequest{
		Bucket: oss.Ptr(a.BucketName),
		Key:    oss.Ptr(objectName),
		Body:   file.Reader,
	})
	if err != nil {
		return
	}

	URL = objectName
	return URL, nil
}

// get the full path of file in bucket
func (a *AliyunOSSAdapter) getPathInBucket(ctx context.Context, file *FileInfo) string {
	url := fmt.Sprintf("%s/%s/%s", a.BasePath, file.Path, file.Name)
	return url
}

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
	"time"

	"formulago/configs"
	"formulago/pkg/img"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
	"github.com/samber/lo"
)

func NewUploadAliyunOSSAdapter(c configs.Config) *UploadAliyunOSSAdapter {
	return &UploadAliyunOSSAdapter{
		Region:        c.S3.AliyunOSS.Region,
		Endpoint:      c.S3.AliyunOSS.Endpoint,
		BucketName:    c.S3.AliyunOSS.BucketName,
		BasePath:      c.S3.AliyunOSS.BasePath,
		SecretID:      c.S3.AliyunOSS.SecretID,
		SecretKey:     c.S3.AliyunOSS.SecretKey,
		Domain:        c.S3.AliyunOSS.Domain,
		AllowFileType: c.S3.AllowFileType,
		MaxFileSize:   c.S3.MaxFileSize,
		sem:           make(chan struct{}, 10),
	}
}

type UploadAliyunOSSAdapter struct {
	Region        string
	Endpoint      string
	SecretID      string
	SecretKey     string
	BucketName    string
	BasePath      string
	Domain        string
	AllowFileType []string      // eg .png .jpg
	MaxFileSize   int64         // bytes
	sem           chan struct{} // 限制并发信号量
}

func (u *UploadAliyunOSSAdapter) UploadFile(ctx context.Context, file *FileInfo) (fileInfo *FileInfo, err error) {
	if file.Reader == nil {
		err = errors.New("not upload any file")
		return
	}
	if file.Size == 0 {
		err = errors.New("file size is zero")
		return
	}
	if file.Size > u.MaxFileSize {
		err = fmt.Errorf("file size %d is too large, max size is %d bytes", file.Size, u.MaxFileSize)
		return
	}
	if !lo.Contains(u.AllowFileType, filepath.Ext(file.Name)) {
		err = fmt.Errorf("file type %s is not allowed", filepath.Ext(file.Name))
		return
	}

	file.URL, err = u.upAction(ctx, file)
	return file, err
}

// SignedUrl get signed url of file, default expire time is 30 minutes
// if customized domain is not empty, signed url will be joined with domain
func (u *UploadAliyunOSSAdapter) SignedUrl(ctx context.Context, filePath string) (signedURL string, err error) {
	// create oss client
	provider := credentials.NewStaticCredentialsProvider(u.SecretID, u.SecretKey)
	cfg := oss.LoadDefaultConfig().
		// 设置区域
		WithRegion(u.Region).
		// 设置Endpoint
		WithEndpoint(u.Endpoint).
		// 设置凭证提供者
		WithCredentialsProvider(provider).
		// 设置HTTP连接超时时间为20秒
		WithConnectTimeout(20 * time.Second).
		// HTTP读取或写入超时时间为60秒
		WithReadWriteTimeout(60 * time.Second).
		// 不校验SSL证书校验
		WithInsecureSkipVerify(true).
		// 设置日志
		WithLogLevel(oss.LogError)

	client := oss.NewClient(cfg)

	// 生成GetObject的预签名URL
	result, err := client.Presign(ctx, &oss.GetObjectRequest{
		Bucket: new(u.BucketName),
		Key:    new(filePath),
	},
		oss.PresignExpires(15*time.Minute),
	)
	if err != nil {
		return
	}
	signedURL = result.URL
	if u.Domain != "" {
		signedURL = strings.Join(strings.Split(signedURL, "/")[3:], "/")
		signedURL = "https://" + u.Domain + "/" + signedURL
	}

	return
}

func (u *UploadAliyunOSSAdapter) ListObjects(ctx context.Context, prefix string) (results []ObjectProperties, err error) {
	// create oss client
	provider := credentials.NewStaticCredentialsProvider(u.SecretID, u.SecretKey)
	cfg := oss.LoadDefaultConfig().
		// 设置区域
		WithRegion(u.Region).
		// 设置Endpoint
		WithEndpoint(u.Endpoint).
		// 设置凭证提供者
		WithCredentialsProvider(provider).
		// 设置HTTP连接超时时间为20秒
		WithConnectTimeout(20 * time.Second).
		// HTTP读取或写入超时时间为60秒
		WithReadWriteTimeout(60 * time.Second).
		// 不校验SSL证书校验
		WithInsecureSkipVerify(true).
		// 设置日志
		WithLogLevel(oss.LogError)

	client := oss.NewClient(cfg)

	// 创建列出对象的请求
	request := &oss.ListObjectsV2Request{
		Bucket: new(u.BucketName),
		Prefix: new(prefix), // 列举指定目录下的所有对象
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
				r.LastModified = obj.LastModified.Format("2006-01-02 15:04:05")
			}
			if obj.LastModified != nil {
				r.LastModified = obj.LastModified.Format("2006-01-02 15:04:05")
			}
			results = append(results, r)
			//fmt.Printf("Object:%v, %v, %v\n", oss.ToString(obj.Key), obj.Size, oss.ToTime(obj.LastModified))
		}
	}

	return results, nil
}

func (u *UploadAliyunOSSAdapter) RestoreObject(ctx context.Context, filePath string, restoreDays int) (err error) {
	// create oss client
	provider := credentials.NewStaticCredentialsProvider(u.SecretID, u.SecretKey)
	cfg := oss.LoadDefaultConfig().
		// 设置区域
		WithRegion(u.Region).
		// 设置Endpoint
		WithEndpoint(u.Endpoint).
		// 设置凭证提供者
		WithCredentialsProvider(provider).
		// 设置HTTP连接超时时间为20秒
		WithConnectTimeout(20 * time.Second).
		// HTTP读取或写入超时时间为60秒
		WithReadWriteTimeout(60 * time.Second).
		// 不校验SSL证书校验
		WithInsecureSkipVerify(true).
		// 设置日志
		WithLogLevel(oss.LogError)

	client := oss.NewClient(cfg)

	// 创建解冻对象的请求
	request := &oss.RestoreObjectRequest{
		Bucket: new(u.BucketName), // 存储空间名称
		Key:    new(filePath),     // 对象名称
		RestoreRequest: &oss.RestoreRequest{
			Days: int32(restoreDays), // 设置解冻状态的持续天数为3天
		},
	}

	// 发送解冻对象的请求
	_, err = client.RestoreObject(ctx, request)
	if err != nil {
		return err
	}
	return nil
}

// FileReader get file bytes
func (u *UploadAliyunOSSAdapter) FileReader(ctx context.Context, filePath string) (fileBytes []byte, err error) {
	u.sem <- struct{}{} // acquire semaphore (max 10 concurrent)
	defer func() { <-u.sem }()

	// create oss client
	provider := credentials.NewStaticCredentialsProvider(u.SecretID, u.SecretKey)
	cfg := oss.LoadDefaultConfig().
		// 设置区域
		WithRegion(u.Region).
		// 设置Endpoint
		WithEndpoint(u.Endpoint).
		// 设置凭证提供者
		WithCredentialsProvider(provider).
		// 设置HTTP连接超时时间为20秒
		WithConnectTimeout(20 * time.Second).
		// HTTP读取或写入超时时间为60秒
		WithReadWriteTimeout(60 * time.Second).
		// 不校验SSL证书校验
		WithInsecureSkipVerify(true).
		// 设置日志
		WithLogLevel(oss.LogError)

	client := oss.NewClient(cfg)

	// get store file object
	result, err := client.GetObject(ctx, &oss.GetObjectRequest{
		Bucket: new(u.BucketName),
		Key:    new(filePath),
	})
	if err != nil {
		return
	}
	defer result.Body.Close()

	fileBytes, err = io.ReadAll(result.Body)
	return
}

// DeleteFile delete file in s3
func (u *UploadAliyunOSSAdapter) DeleteFile(ctx context.Context, filePath []string) (deletedFilePath []string, err error) {
	// create oss client
	provider := credentials.NewStaticCredentialsProvider(u.SecretID, u.SecretKey)
	cfg := oss.LoadDefaultConfig().
		// 设置区域
		WithRegion(u.Region).
		// 设置Endpoint
		WithEndpoint(u.Endpoint).
		// 设置凭证提供者
		WithCredentialsProvider(provider).
		// 设置HTTP连接超时时间为20秒
		WithConnectTimeout(20 * time.Second).
		// HTTP读取或写入超时时间为60秒
		WithReadWriteTimeout(60 * time.Second).
		// 不校验SSL证书校验
		WithInsecureSkipVerify(true).
		// 设置日志
		WithLogLevel(oss.LogError)

	client := oss.NewClient(cfg)

	// delete file
	var deleteObjects []oss.DeleteObject
	for _, v := range filePath {
		deleteObjects = append(deleteObjects, oss.DeleteObject{
			Key: new(v),
		})
	}
	result, err := client.DeleteMultipleObjects(ctx, &oss.DeleteMultipleObjectsRequest{
		Bucket:  new(u.BucketName),
		Objects: deleteObjects,
	})
	if err != nil {
		return
	}

	// get deleted file path
	for _, v := range result.DeletedObjects {
		deletedFilePath = append(deletedFilePath, *v.Key)
	}

	if len(deletedFilePath) == 0 {
		return deletedFilePath, errors.New("no file deleted")
	}
	return
}

// upAction, upload file to aliyun oss
func (u *UploadAliyunOSSAdapter) upAction(ctx context.Context, file *FileInfo) (fileFullUrl string, err error) {
	// <yourObjectName>upload file to OSS need specify the full path of file, like abc/efg/123.jpg
	objectName := u.getPathInBucket(ctx, file)
	// create oss client
	provider := credentials.NewStaticCredentialsProvider(u.SecretID, u.SecretKey)
	cfg := oss.LoadDefaultConfig().
		// 设置区域
		WithRegion(u.Region).
		// 设置Endpoint
		WithEndpoint(u.Endpoint).
		// 设置凭证提供者
		WithCredentialsProvider(provider).
		// 设置HTTP连接超时时间为20秒
		WithConnectTimeout(20 * time.Second).
		// HTTP读取或写入超时时间为60秒
		WithReadWriteTimeout(60 * time.Second).
		// 不校验SSL证书校验
		WithInsecureSkipVerify(true).
		// 设置日志
		WithLogLevel(oss.LogError)

	client := oss.NewClient(cfg)

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
		Bucket: new(u.BucketName),
		Key:    new(objectName),
		Body:   file.Reader,
	})
	if err != nil {
		return
	}

	fileFullUrl = objectName
	return fileFullUrl, nil
}

// get the full path of file in bucket
func (u *UploadAliyunOSSAdapter) getPathInBucket(ctx context.Context, file *FileInfo) string {
	url := fmt.Sprintf("%s/%s/%s", u.BasePath, file.Path, file.Name)
	return url
}

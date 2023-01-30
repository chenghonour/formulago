/*
 * Copyright 2023 FormulaGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 */

package data

import (
	"formulago/configs"
	"formulago/data/s3"
)

// fileStore offers file store interface
var fileStore *FileAdapter

func init() {
	fileStore = NewFileAdapter(configs.Data())
}

// FileStore Get a default file store instance
func FileStore() *FileAdapter {
	return fileStore
}

type FileAdapter struct {
	Adapter s3.Adapter
}

func NewFileAdapter(c configs.Config) *FileAdapter {
	if c.S3.Type == "AliyunOSS" {
		return &FileAdapter{
			Adapter: s3.NewUploadAliyunOSSAdapter(c),
		}
	}
	return nil
}

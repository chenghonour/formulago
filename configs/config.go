/*
 * Copyright 2023 FormulaGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 */

// package configs provide configuration file loading and related methods

package configs

import (
	"embed"
	"os"
	"sync"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/knadh/koanf/v2"
)

//go:embed *.yaml
var configFiles embed.FS

// GlobalConfig .
var (
	globalConfig Config
	isInit       = false
)

var initConfigMutex sync.Mutex

func InitConfig(configBytes []byte) {
	initConfigMutex.Lock()
	defer initConfigMutex.Unlock()
	if isInit {
		return
	}
	// log print embed config file
	DirEntry, err := configFiles.ReadDir(".")
	if err != nil {
		hlog.Fatal("read config embed dir error: ", err)
	}
	for _, v := range DirEntry {
		hlog.Info("embed config file: ", v.Name())
	}
	// load config
	globalConfig, _ = load(configBytes)
	isInit = true
}

func Data() Config {
	if !isInit {
		InitConfig(nil)
	}
	return globalConfig
}

func ReLoad(configBytes []byte) {
	globalConfig, _ = load(configBytes)
	hlog.Info("reload config")
	return
}

// load from bytes, if nil, load from embed file
func load(data []byte) (config Config, err error) {
	if len(data) == 0 {
		data, err = FSConfigFileToBytes(configFiles)
		if err != nil {
			hlog.Fatal("load config from embed FS error: ", err)
			return config, err
		}
	}
	// koanf instance. Use . as the key path delimiter. This can be / or anything.
	var k = koanf.New(".")
	err = k.Load(rawbytes.Provider(data), yaml.Parser())
	if err != nil {
		return config, err
	}
	err = k.Unmarshal("", &config)
	if err != nil {
		return config, err
	}
	// load db password from env
	DBPassword := os.Getenv("DB_PASSWORD")
	if DBPassword != "" {
		config.Database.Password = DBPassword
	}
	return
}

func FSConfigFileToBytes(fs embed.FS) ([]byte, error) {
	// Obtain whether it is prod environment from the environment variable
	IsProd := os.Getenv("IS_PROD")
	configFile := "config_dev.yaml"
	if IsProd == "true" {
		// If it is prod environment, load the prod configuration file
		configFile = "config.yaml"
		hlog.Info("load prod config")
	} else {
		// If it is not prod environment, load the dev configuration file
		hlog.Info("load dev config")
	}

	// load config from embed file
	configBytes, err := fs.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	return configBytes, nil
}

// Config is the configuration of the project.
type Config struct {
	Name     string     `yaml:"Name"`
	IsDemo   bool       `yaml:"IsDemo"`
	IsProd   bool       `yaml:"IsProd"`
	Host     string     `yaml:"Host"`
	Port     int        `yaml:"Port"`
	Timeout  int        `yaml:"Timeout"`
	Captcha  Captcha    `yaml:"Captcha"`
	Auth     Auth       `yaml:"Auth"`
	Redis    Redis      `yaml:"Redis"`
	Database Database   `yaml:"Database"`
	Casbin   CasbinConf `yaml:"Casbin"`
	S3       S3         `yaml:"S3"`
	Wecom    Wecom      `yaml:"Wecom"`
}

// Captcha is the configuration of the captcha.
type Captcha struct {
	KeyLong   int `yaml:"KeyLong"`
	ImgWidth  int `yaml:"ImgWidth"`
	ImgHeight int `yaml:"ImgHeight"`
}

// Auth is the configuration of the auth.
type Auth struct {
	OAuthKey     string `yaml:"OAuthKey"`
	AccessSecret string `yaml:"AccessSecret"`
	AccessExpire int    `yaml:"AccessExpire"`
}

// Redis is the configuration of the redis.
type Redis struct {
	Enable   bool   `yaml:"Enable"`
	Host     string `yaml:"Host"`
	Port     int    `yaml:"Port"`
	Type     string `yaml:"Type"`
	Password string `yaml:"Password"`
	DB       int    `yaml:"DB"`
}

// Database is the configuration of the database.
type Database struct {
	Type        string      `yaml:"Type"`
	Host        string      `yaml:"Host"`
	Port        int         `yaml:"Port"`
	DBName      string      `yaml:"DBName"`
	Username    interface{} `yaml:"Username"`
	Password    interface{} `yaml:"Password"`
	MaxOpenConn int         `yaml:"MaxOpenConn"`
	SSLMode     bool        `yaml:"SSLMode"`
	CacheTime   int         `yaml:"CacheTime"`
}

// CasbinConf is the configuration of the casbin.
type CasbinConf struct {
	ModelText string `yaml:"ModelText"`
}

// S3 Simple Storage Service
type S3 struct {
	Type      string `yaml:"Type"`
	AliyunOSS struct {
		Region     string `yaml:"Region"`
		Endpoint   string `yaml:"Endpoint"`
		BucketName string `yaml:"BucketName"`
		BasePath   string `yaml:"BasePath"`
		SecretID   string `yaml:"SecretID"`
		SecretKey  string `yaml:"SecretKey"`
		Domain     string `yaml:"Domain"`
	} `yaml:"AliyunOSS"`
}

// Wecom corp
type Wecom struct {
	CorpID         string `yaml:"CorpId"`
	SecretID       string `yaml:"SecretId"`
	Token          string `yaml:"Token"`
	EncodingAESKey string `yaml:"EncodingAESKey"`
}

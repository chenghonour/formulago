/*
 * Copyright 2023 FormulaGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 */

package admin

import (
	"formulago/biz/domain"
	"formulago/configs"
	"formulago/data"
	"formulago/pkg/captcha"
	"github.com/mojocn/base64Captcha"
)

var (
	captchaDriver *base64Captcha.DriverDigit
	CaptchaStore  *captcha.CacheStore
)

func init() {
	c := configs.Data()
	captchaDriver = base64Captcha.NewDriverDigit(c.Captcha.ImgHeight, c.Captcha.ImgWidth,
		c.Captcha.KeyLong, 0.7, 80)
	CaptchaStore = captcha.NewCacheStore(data.CacheDB("captcha"))
}

type Captcha struct {
	CaptchaDriver *base64Captcha.DriverDigit
	CaptchaStore  *captcha.CacheStore
}

func NewCaptcha() domain.Captcha {
	return &Captcha{
		CaptchaDriver: captchaDriver,
		CaptchaStore:  CaptchaStore,
	}
}

func (c *Captcha) GetCaptcha() (id, b64s string, err error) {
	captchaGen := base64Captcha.NewCaptcha(c.CaptchaDriver, c.CaptchaStore)
	id, b64s, err = captchaGen.Generate()
	return
}

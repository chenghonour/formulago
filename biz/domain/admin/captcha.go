/*
 * Copyright 2023 FormulaGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 */

package admin

type Captcha interface {
	GetCaptcha() (id, b64s string, err error)
}

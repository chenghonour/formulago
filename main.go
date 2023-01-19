/*
 * Copyright 2022 FormulaGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 */

package main

import (
	"fmt"

	"formulago/configs"
	"github.com/cloudwego/hertz/pkg/app/server"
)

func main() {
	c := configs.Data()
	h := server.Default(
		server.WithHostPorts(fmt.Sprintf("%s:%d", c.Host, c.Port)))

	register(h)
	h.Spin()
}

/*
 * Copyright 2023 FormulaGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 */

package main

import (
	"fmt"

	"formulago/configs"
	"formulago/data"

	"github.com/cloudwego/hertz/pkg/app/server"
)

func main() {
	// init
	data.InitDataConfig()

	// start server
	c := configs.Data()
	h := server.Default(
		server.WithHostPorts(fmt.Sprintf("%s:%d", c.Host, c.Port)))

	register(h)
	h.Spin()
}

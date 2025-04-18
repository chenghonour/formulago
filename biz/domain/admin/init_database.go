/*
 * Copyright 2023 FormulaGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 */

package admin

import "context"

type InitDatabase interface {
	InitDatabase(ctx context.Context) error
}

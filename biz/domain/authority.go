/*
 * Copyright 2022 FormulaGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 */

package domain

import "context"

type Authority interface {
	UpdateApiAuthority(ctx context.Context, roleIDStr string, infos []*ApiAuthorityInfo) error
	ApiAuthority(ctx context.Context, roleIDStr string) (infos []*ApiAuthorityInfo, err error)
	UpdateMenuAuthority(ctx context.Context, roleID uint64, menuIDs []uint64) error
	MenuAuthority(ctx context.Context, roleID uint64) (menuIDs []uint64, err error)
}

type ApiAuthorityInfo struct {
	Path   string
	Method string
}

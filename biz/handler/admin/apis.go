// Code generated by hertz generator.

package admin

import (
	"context"
	admin2 "formulago/biz/domain/admin"
	logic "formulago/biz/logic/admin"
	"formulago/data"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/jinzhu/copier"

	"formulago/api/model/admin"
	base "formulago/api/model/base"
	"github.com/cloudwego/hertz/pkg/app"
)

// CreateApi .
// @router /api/admin/api/create [POST]
func CreateApi(ctx context.Context, c *app.RequestContext) {
	var err error
	var req admin.ApiInfo
	resp := new(base.BaseResp)
	err = c.BindAndValidate(&req)
	if err != nil {
		resp.ErrCode = base.ErrCode_Fail
		resp.ErrMsg = err.Error()
		c.JSON(consts.StatusBadRequest, resp)
		return
	}

	var ApiInfoReq admin2.ApiInfo
	err = copier.Copy(&ApiInfoReq, &req)
	if err != nil {
		resp.ErrCode = base.ErrCode_Fail
		resp.ErrMsg = err.Error()
		c.JSON(consts.StatusInternalServerError, resp)
		return
	}

	err = logic.NewApi(data.Default()).Create(ctx, ApiInfoReq)
	if err != nil {
		resp.ErrCode = base.ErrCode_Fail
		resp.ErrMsg = err.Error()
		c.JSON(consts.StatusInternalServerError, resp)
		return
	}

	resp.ErrCode = base.ErrCode_Success
	resp.ErrMsg = "success"
	c.JSON(consts.StatusOK, resp)
}

// UpdateApi .
// @router /api/admin/api/update [POST]
func UpdateApi(ctx context.Context, c *app.RequestContext) {
	var err error
	var req admin.ApiInfo
	resp := new(base.BaseResp)
	err = c.BindAndValidate(&req)
	if err != nil {
		resp.ErrCode = base.ErrCode_Fail
		resp.ErrMsg = err.Error()
		c.JSON(consts.StatusBadRequest, resp)
		return
	}

	var ApiInfoReq admin2.ApiInfo
	err = copier.Copy(&ApiInfoReq, &req)
	if err != nil {
		resp.ErrCode = base.ErrCode_Fail
		resp.ErrMsg = err.Error()
		c.JSON(consts.StatusInternalServerError, resp)
		return
	}

	err = logic.NewApi(data.Default()).Update(ctx, ApiInfoReq)
	if err != nil {
		resp.ErrCode = base.ErrCode_Fail
		resp.ErrMsg = err.Error()
		c.JSON(consts.StatusInternalServerError, resp)
		return
	}

	resp.ErrCode = base.ErrCode_Success
	resp.ErrMsg = "success"
	c.JSON(consts.StatusOK, resp)
}

// DeleteApi .
// @router /api/admin/api [DELETE]
func DeleteApi(ctx context.Context, c *app.RequestContext) {
	var err error
	var req base.IDReq
	resp := new(base.BaseResp)
	err = c.BindAndValidate(&req)
	if err != nil {
		resp.ErrCode = base.ErrCode_Fail
		resp.ErrMsg = err.Error()
		c.JSON(consts.StatusBadRequest, resp)
		return
	}

	err = logic.NewApi(data.Default()).Delete(ctx, req.ID)
	if err != nil {
		resp.ErrCode = base.ErrCode_Fail
		resp.ErrMsg = err.Error()
		c.JSON(consts.StatusInternalServerError, resp)
		return
	}

	resp.ErrCode = base.ErrCode_Success
	resp.ErrMsg = "success"
	c.JSON(consts.StatusOK, resp)
}

// ApiList .
// @router /api/admin/api/list [GET]
func ApiList(ctx context.Context, c *app.RequestContext) {
	var err error
	var req admin.ApiPageReq
	resp := new(admin.ApiListResp)
	err = c.BindAndValidate(&req)
	if err != nil {
		resp.ErrCode = base.ErrCode_Fail
		resp.ErrMsg = err.Error()
		c.JSON(consts.StatusBadRequest, resp)
		return
	}

	var ApiPageReq admin2.ListApiReq
	err = copier.Copy(&ApiPageReq, &req)
	if err != nil {
		resp.ErrCode = base.ErrCode_Fail
		resp.ErrMsg = err.Error()
		c.JSON(consts.StatusInternalServerError, resp)
		return
	}
	list, total, err := logic.NewApi(data.Default()).List(ctx, ApiPageReq)
	if err != nil {
		resp.ErrCode = base.ErrCode_Fail
		resp.ErrMsg = err.Error()
		c.JSON(consts.StatusInternalServerError, resp)
		return
	}
	for _, v := range list {
		var ApiInfo admin.ApiInfo
		err = copier.Copy(&ApiInfo, &v)
		if err != nil {
			resp.ErrCode = base.ErrCode_Fail
			resp.ErrMsg = err.Error()
			c.JSON(consts.StatusInternalServerError, resp)
			return
		}
		resp.Data = append(resp.Data, &ApiInfo)
	}
	resp.Total = uint64(total)
	resp.ErrCode = base.ErrCode_Success
	resp.ErrMsg = "success"
	c.JSON(consts.StatusOK, resp)
}

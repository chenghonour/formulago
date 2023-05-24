// Code generated by hertz generator.

package admin

import (
	"context"
	"formulago/biz/domain"
	logic "formulago/biz/logic/admin"
	"formulago/data"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/jinzhu/copier"

	"formulago/api/model/admin"
	base "formulago/api/model/base"
	"github.com/cloudwego/hertz/pkg/app"
)

// CreateRole .
// @router /api/admin/role/create [POST]
func CreateRole(ctx context.Context, c *app.RequestContext) {
	var err error
	var req admin.RoleInfo
	resp := new(base.BaseResp)
	err = c.BindAndValidate(&req)
	if err != nil {
		resp.ErrCode = base.ErrCode_Fail
		resp.ErrMsg = err.Error()
		c.JSON(consts.StatusBadRequest, resp)
		return
	}

	err = logic.NewRole(data.Default()).Create(ctx, domain.RoleInfo{
		Name:          req.Name,
		Value:         req.Value,
		DefaultRouter: req.DefaultRouter,
		Status:        req.Status,
		Remark:        req.Remark,
		OrderNo:       req.OrderNo,
	})
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

// UpdateRole .
// @router /api/admin/role/update [POST]
func UpdateRole(ctx context.Context, c *app.RequestContext) {
	var err error
	var req admin.RoleInfo
	resp := new(base.BaseResp)
	err = c.BindAndValidate(&req)
	if err != nil {
		resp.ErrCode = base.ErrCode_Fail
		resp.ErrMsg = err.Error()
		c.JSON(consts.StatusBadRequest, resp)
		return
	}

	err = logic.NewRole(data.Default()).Update(ctx, domain.RoleInfo{
		ID:            req.ID,
		Name:          req.Name,
		Value:         req.Value,
		DefaultRouter: req.DefaultRouter,
		Status:        req.Status,
		Remark:        req.Remark,
		OrderNo:       req.OrderNo,
	})
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

// DeleteRole .
// @router /api/admin/role [DELETE]
func DeleteRole(ctx context.Context, c *app.RequestContext) {
	var err error
	var req admin.IDReq
	resp := new(base.BaseResp)
	err = c.BindAndValidate(&req)
	if err != nil {
		resp.ErrCode = base.ErrCode_Fail
		resp.ErrMsg = err.Error()
		c.JSON(consts.StatusBadRequest, resp)
		return
	}

	err = logic.NewRole(data.Default()).Delete(ctx, req.ID)
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

// RoleByID .
// @router /api/admin/role [GET]
func RoleByID(ctx context.Context, c *app.RequestContext) {
	var err error
	var req admin.IDReq
	resp := new(admin.RoleInfoResp)
	err = c.BindAndValidate(&req)
	if err != nil {
		resp.ErrCode = base.ErrCode_Fail
		resp.ErrMsg = err.Error()
		c.JSON(consts.StatusBadRequest, resp)
		return
	}

	roleInfo, err := logic.NewRole(data.Default()).RoleInfoByID(ctx, req.ID)
	if err != nil {
		resp.ErrCode = base.ErrCode_Fail
		resp.ErrMsg = err.Error()
		c.JSON(consts.StatusInternalServerError, resp)
		return
	}
	err = copier.Copy(&resp, &roleInfo)
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

// RoleList .
// @router /api/admin/role/list [GET]
func RoleList(ctx context.Context, c *app.RequestContext) {
	var err error
	var req admin.PageInfoReq
	resp := new(admin.RoleListResp)
	err = c.BindAndValidate(&req)
	if err != nil {
		resp.ErrCode = base.ErrCode_Fail
		resp.ErrMsg = err.Error()
		c.JSON(consts.StatusBadRequest, resp)
		return
	}

	var listReq domain.RoleListReq
	listReq.Page = req.Page
	listReq.PageSize = req.PageSize
	list, total, err := logic.NewRole(data.Default()).List(ctx, &listReq)
	if err != nil {
		resp.ErrCode = base.ErrCode_Fail
		resp.ErrMsg = err.Error()
		c.JSON(consts.StatusInternalServerError, resp)
		return
	}
	var infos []*admin.RoleInfo
	err = copier.Copy(&infos, &list)
	if err != nil {
		resp.ErrCode = base.ErrCode_Fail
		resp.ErrMsg = err.Error()
		c.JSON(consts.StatusInternalServerError, resp)
		return
	}

	resp.Data = infos
	resp.Total = uint64(total)
	resp.ErrCode = base.ErrCode_Success
	resp.ErrMsg = "success"
	c.JSON(consts.StatusOK, resp)
}

// UpdateRoleStatus .
// @router /api/admin/role/status [POST]
func UpdateRoleStatus(ctx context.Context, c *app.RequestContext) {
	var err error
	var req admin.StatusCodeReq
	resp := new(base.BaseResp)
	err = c.BindAndValidate(&req)
	if err != nil {
		resp.ErrCode = base.ErrCode_Fail
		resp.ErrMsg = err.Error()
		c.JSON(consts.StatusBadRequest, resp)
		return
	}

	err = logic.NewRole(data.Default()).UpdateStatus(ctx, req.ID, uint8(req.Status))
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

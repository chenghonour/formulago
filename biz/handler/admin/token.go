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

// UpdateToken .
// @router /api/admin/token/update [POST]
func UpdateToken(ctx context.Context, c *app.RequestContext) {
	var err error
	var req admin.TokenInfo
	resp := new(base.BaseResp)
	err = c.BindAndValidate(&req)
	if err != nil {
		resp.ErrCode = base.ErrCode_Fail
		resp.ErrMsg = err.Error()
		c.JSON(consts.StatusBadRequest, resp)
		return
	}

	var tokenInfo admin2.TokenInfo
	err = copier.Copy(&tokenInfo, &req)
	if err != nil {
		resp.ErrCode = base.ErrCode_Fail
		resp.ErrMsg = err.Error()
		c.JSON(consts.StatusInternalServerError, resp)
		return
	}

	err = logic.NewToken(data.Default()).Update(ctx, &tokenInfo)
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

// DeleteToken .
// @router /api/admin/token [DELETE]
func DeleteToken(ctx context.Context, c *app.RequestContext) {
	var err error
	var req admin.DeleteReq
	resp := new(base.BaseResp)
	err = c.BindAndValidate(&req)
	if err != nil {
		resp.ErrCode = base.ErrCode_Fail
		resp.ErrMsg = err.Error()
		c.JSON(consts.StatusBadRequest, resp)
		return
	}

	err = logic.NewToken(data.Default()).Delete(ctx, req.UserID)
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

// TokenList .
// @router /api/admin/token/list [GET]
func TokenList(ctx context.Context, c *app.RequestContext) {
	var err error
	var req admin.TokenListReq
	resp := new(admin.TokenListResp)
	err = c.BindAndValidate(&req)
	if err != nil {
		resp.ErrCode = base.ErrCode_Fail
		resp.ErrMsg = err.Error()
		c.JSON(consts.StatusBadRequest, resp)
		return
	}

	var tokenListReq admin2.TokenListReq
	err = copier.Copy(&tokenListReq, &req)
	if err != nil {
		resp.ErrCode = base.ErrCode_Fail
		resp.ErrMsg = err.Error()
		c.JSON(consts.StatusInternalServerError, resp)
		return
	}
	tokens, total, err := logic.NewToken(data.Default()).List(ctx, &tokenListReq)
	if err != nil {
		resp.ErrCode = base.ErrCode_Fail
		resp.ErrMsg = err.Error()
		c.JSON(consts.StatusInternalServerError, resp)
		return
	}
	for _, token := range tokens {
		var tokenInfo admin.TokenInfo
		tokenInfo.ID = token.ID
		tokenInfo.UserID = token.UserID
		tokenInfo.UserName = token.UserName
		tokenInfo.CreatedAt = token.CreatedAt
		tokenInfo.UpdatedAt = token.UpdatedAt
		tokenInfo.ExpiredAt = token.ExpiredAt
		resp.Data = append(resp.Data, &tokenInfo)
	}
	resp.Total = uint64(total)
	resp.ErrCode = base.ErrCode_Success
	resp.ErrMsg = "success"
	c.JSON(consts.StatusOK, resp)
}

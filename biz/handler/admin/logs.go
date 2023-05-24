// Code generated by hertz generator.

package admin

import (
	"context"
	"formulago/biz/domain"
	logic "formulago/biz/logic/admin"
	"formulago/data"

	admin "formulago/api/model/admin"
	base "formulago/api/model/base"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

// GetLogsList .
// @router /api/admin/logs/list [GET]
func GetLogsList(ctx context.Context, c *app.RequestContext) {
	var err error
	var req admin.LogsListReq
	resp := new(admin.LogsListResp)
	err = c.BindAndValidate(&req)
	if err != nil {
		resp.ErrCode = base.ErrCode_Fail
		resp.ErrMsg = err.Error()
		c.JSON(consts.StatusBadRequest, resp)
		return
	}

	var logsListReq domain.LogsListReq
	logsListReq.Page = req.Page
	logsListReq.PageSize = req.PageSize
	logsListReq.Type = req.Type
	logsListReq.Api = req.Api
	logsListReq.Method = req.Method
	logsListReq.Operator = req.Operator
	switch req.Success {
	case "true":
		success := true
		logsListReq.Success = &success
	case "false":
		success := false
		logsListReq.Success = &success
	default:
		logsListReq.Success = nil
	}
	logsList, total, err := logic.NewLogs(data.Default()).List(ctx, &logsListReq)
	if err != nil {
		resp.ErrCode = base.ErrCode_Fail
		resp.ErrMsg = err.Error()
		c.JSON(consts.StatusInternalServerError, resp)
		return
	}

	var list []*admin.LogsInfo
	for _, v := range logsList {
		var logsInfo admin.LogsInfo
		logsInfo.Type = v.Type
		logsInfo.Method = v.Method
		logsInfo.Api = v.Api
		logsInfo.Success = v.Success
		logsInfo.ReqContent = v.ReqContent
		logsInfo.RespContent = v.RespContent
		logsInfo.Ip = v.Ip
		logsInfo.UserAgent = v.UserAgent
		logsInfo.Operator = v.Operator
		logsInfo.Time = v.Time
		logsInfo.CreatedAt = v.CreatedAt
		logsInfo.UpdatedAt = v.UpdatedAt
		list = append(list, &logsInfo)
	}

	resp.Data = list
	resp.Total = uint64(total)
	resp.ErrCode = base.ErrCode_Success
	resp.ErrMsg = "success"
	c.JSON(consts.StatusOK, resp)
}

// DeleteLogs .
// @router /api/admin/logs/deleteAll [DELETE]
func DeleteLogs(ctx context.Context, c *app.RequestContext) {
	var err error
	var req admin.Empty
	err = c.BindAndValidate(&req)
	if err != nil {
		c.String(consts.StatusBadRequest, err.Error())
		return
	}

	resp := new(base.BaseResp)
	resp.ErrCode = base.ErrCode_Success
	resp.ErrMsg = "success"
	c.JSON(consts.StatusOK, resp)
}

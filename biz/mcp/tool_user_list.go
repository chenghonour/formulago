/*
 * Copyright 2023 FormulaGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package mcp

import (
	"context"
	"encoding/json"

	domain "formulago/biz/domain/admin"
	logic "formulago/biz/logic/admin"
	"formulago/pkg/times"

	"github.com/mark3labs/mcp-go/mcp"
)

// userListDef wraps the HTTP endpoint POST /api/admin/user/list so the casbin
// permission check matches exactly. Rows mirror the HTTP response fields
// (time formatted with pkg/times.TimeFormat) without the errCode/errMsg shell.
func userListDef(s *Server) toolDef {
	tool := mcp.NewTool("user_list",
		mcp.WithDescription("分页查询用户列表（对应接口 POST /api/admin/user/list）。返回 page, total 与 data 列表，data 项字段：id, username, nickname, email, mobile, status, roleID, avatar, createdAt, updatedAt"),
		mcp.WithNumber("page", mcp.Description("页码，从 1 开始；缺省 1")),
		mcp.WithNumber("pageSize", mcp.Description("每页条数，缺省 10")),
		mcp.WithString("username", mcp.Description("按用户名模糊过滤，可选")),
		mcp.WithString("nickname", mcp.Description("按昵称模糊过滤，可选")),
		mcp.WithString("email", mcp.Description("按邮箱精确过滤，可选")),
		mcp.WithString("mobile", mcp.Description("按手机号精确过滤，可选")),
		mcp.WithNumber("roleID", mcp.Description("按角色 ID 精确过滤，可选")),
	)
	return toolDef{
		tool:      tool,
		apiPath:   "/api/admin/user/list",
		apiMethod: "POST",
		run:       runUserList(s),
	}
}

// userListRow mirrors the HTTP UserInfoResp fields served by UserList.
type userListRow struct {
	ID        uint64 `json:"id"`
	Avatar    string `json:"avatar"`
	RoleID    uint64 `json:"roleID"`
	Mobile    string `json:"mobile"`
	Email     string `json:"email"`
	Status    uint8  `json:"status"`
	Username  string `json:"username"`
	Nickname  string `json:"nickname"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

func runUserList(s *Server) func(context.Context, *authUser, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, _ *authUser, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var dreq domain.UserListReq
		if err := req.BindArguments(&dreq); err != nil {
			return mcp.NewToolResultError("参数解析失败：" + err.Error() +
				"（可选参数：page、pageSize、username、nickname、email、mobile、roleID）"), nil
		}
		// normalize paging: the HTTP layer underflows its offset when Page==0
		// and misqueries when PageSize==0
		if dreq.Page < 1 {
			dreq.Page = 1
		}
		if dreq.PageSize < 1 {
			dreq.PageSize = 10
		}
		list, total, err := logic.NewUser(s.data).List(ctx, dreq)
		if err != nil {
			return mcp.NewToolResultError("查询用户列表失败"), err
		}
		rows := make([]userListRow, 0, len(list))
		for _, v := range list {
			rows = append(rows, userListRow{
				ID:        v.ID,
				Avatar:    v.Avatar,
				RoleID:    v.RoleID,
				Mobile:    v.Mobile,
				Email:     v.Email,
				Status:    v.Status,
				Username:  v.Username,
				Nickname:  v.Nickname,
				CreatedAt: v.CreatedAt.Format(times.TimeFormat),
				UpdatedAt: v.UpdatedAt.Format(times.TimeFormat),
			})
		}
		body, err := json.Marshal(struct {
			Page  uint64        `json:"page"`
			Total int           `json:"total"`
			Data  []userListRow `json:"data"`
		}{Page: dreq.Page, Total: total, Data: rows})
		if err != nil {
			return mcp.NewToolResultError("序列化结果失败"), err
		}
		return mcp.NewToolResultText(string(body)), nil
	}
}

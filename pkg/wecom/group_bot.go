/*
 * Copyright 2023 FormulaGo Authors
 *
 * Created by hua
 */

package wecom

import (
	"fmt"
	"strings"
)

type MarkdownMsg struct {
	Msgtype  string `json:"msgtype"`
	Markdown struct {
		Content string `json:"content"`
	} `json:"markdown"`
}

const (
	SuccessEmoji = "✅"
	CarEmoji     = "🚚"
	PlaneEmoji   = "✈️"
	ErrorEmoji   = "❌"
	WarningEmoji = "‼️"
	EmailEmoji   = "💌"
	PoliceEmoji  = "👮"
)

func NewMarkdownMsg(title, content string, emoji string) *MarkdownMsg {
	contentList := strings.Split(content, "\n")
	for i, v := range contentList {
		contentList[i] = fmt.Sprintf("> %s", v)
	}
	content = strings.Join(contentList, "\n")

	switch emoji {
	case SuccessEmoji:
		content = fmt.Sprintf("<font color=\"info\">%s %s</font>\n%s", emoji, title, content)
	case EmailEmoji:
		content = fmt.Sprintf("<font color=\"info\">%s %s</font>\n%s", emoji, title, content)
	case CarEmoji:
		content = fmt.Sprintf("<font color=\"info\">%s %s</font>\n%s", emoji, title, content)
	case PlaneEmoji:
		content = fmt.Sprintf("<font color=\"info\">%s %s</font>\n%s", emoji, title, content)
	case ErrorEmoji:
		content = fmt.Sprintf("<font color=\"warning\">%s %s</font>\n%s", emoji, title, content)
	case WarningEmoji:
		content = fmt.Sprintf("<font color=\"warning\">%s %s</font>\n%s", emoji, title, content)
	case PoliceEmoji:
		content = fmt.Sprintf("<font color=\"warning\">%s %s</font>\n%s", emoji, title, content)
	}

	// 企业微信机器人markdown内容，最长不超过4096个字节，必须是utf8编码
	if len([]byte(content)) > 4096 {
		newByte := []byte(content)[:4096]
		content = string(newByte)
	}
	return &MarkdownMsg{
		Msgtype: "markdown",
		Markdown: struct {
			Content string `json:"content"`
		}{Content: content},
	}
}

// PostToGroupBot 推送信息到企业微信机器人
// url 支持多个机器人地址，用逗号,分隔
func PostToGroupBot(urlList []string, msg *MarkdownMsg) error {
	for _, v := range urlList {
		if v == "" {
			continue
		}
		// code==0, 表示推送成功，否则失败
		type SheinResp struct {
			Code int    `json:"errcode"`
			Msg  string `json:"errmsg"`
		}
		var resp SheinResp
		_, err := client.R().SetHeader("Content-Type", "application/json").
			SetBody(msg).
			SetErrorResult(&resp).
			SetSuccessResult(&resp).
			Post(v)
		if err != nil {
			return err
		}
		if resp.Code != 0 {
			err = fmt.Errorf("推送信息到企业微信机器人失败，失败原因：%s", resp.Msg)
			return err
		}
	}
	return nil
}

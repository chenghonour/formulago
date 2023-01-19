package admin

import (
	"bytes"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/json"
	"github.com/cloudwego/hertz/pkg/common/test/assert"
	"github.com/cloudwego/hertz/pkg/common/ut"
	"formulago/api/model/admin"
	"testing"
)

func TestHealthCheck(t *testing.T) {
	h := server.Default()
	h.GET("/api/health", HealthCheck)
	json := `{"version":"v0.0.1"}`
	w := ut.PerformRequest(h.Engine, "GET", "/api/health",
		&ut.Body{Body: bytes.NewBufferString(json), Len: len(json)},
		ut.Header{Key: "Connection", Value: "close"},
		ut.Header{Key: "Content-Type", Value: "application/json"})
	resp := w.Result()
	assert.DeepEqual(t, 200, resp.StatusCode())
	assert.DeepEqual(t, `{"errCode":0,"errMsg":"success"}`, string(resp.Body()))
}

func TestCaptcha(t *testing.T) {
	h := server.Default()
	h.GET("/api/captcha", HealthCheck)
	w := ut.PerformRequest(h.Engine, "GET", "/api/captcha", nil,
		ut.Header{Key: "Connection", Value: "close"},
		ut.Header{Key: "Content-Type", Value: "application/json"})
	resp := w.Result()
	captchaInfoResp := new(admin.CaptchaInfoResp)
	err := json.Unmarshal(resp.Body(), &captchaInfoResp)
	if err != nil {
		t.Error(err)
		return
	}
	assert.DeepEqual(t, 200, resp.StatusCode())
	assert.DeepEqual(t, admin.ErrCode_Success, captchaInfoResp.ErrCode)
	assert.DeepEqual(t, "success", captchaInfoResp.ErrMsg)
}

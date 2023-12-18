package admin

import (
	"bytes"
	"testing"

	"formulago/api/model/admin"
	"formulago/api/model/base"

	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/json"
	"github.com/cloudwego/hertz/pkg/common/test/assert"
	"github.com/cloudwego/hertz/pkg/common/ut"
)

func TestHealthCheck(t *testing.T) {
	h := server.Default()
	h.GET("/api/health", HealthCheck)
	jsonStr := `{"version":"v0.0.1"}`
	w := ut.PerformRequest(h.Engine, "GET", "/api/health",
		&ut.Body{Body: bytes.NewBufferString(jsonStr), Len: len(jsonStr)},
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
	assert.DeepEqual(t, base.ErrCode_Success, captchaInfoResp.ErrCode)
	assert.DeepEqual(t, "success", captchaInfoResp.ErrMsg)
}

func TestDeleteStructTag(t *testing.T) {
	h := server.Default()
	h.POST("/api/deleteStructTag", DeleteStructTag)
	jsonStr := "{\"structStr\":\"// test struct\\ntype a struct {\\n\\t// primary key\\n\\tID uint64 `json:\\\"ID\\\"`\\n\\t// name\\n\\tName string `json:\\\"name\\\"`\\n\\t// created time\\n\\tCreatedAt time.Time `json:\\\"created_at\\\"`\\n}\"}"
	w := ut.PerformRequest(h.Engine, "POST", "/api/deleteStructTag",
		&ut.Body{Body: bytes.NewBufferString(jsonStr), Len: len(jsonStr)},
		ut.Header{Key: "Connection", Value: "close"},
		ut.Header{Key: "Content-Type", Value: "application/json"})
	resp := w.Result()
	assert.DeepEqual(t, 200, resp.StatusCode())
	assert.DeepEqual(t, `{"errCode":0,"errMsg":"success","structStr":"// test struct\ntype a struct {\n  // primary key\n  ID uint64 \n  // name\n  Name string \n  // created time\n  CreatedAt time.Time \n}\n"}`,
		string(resp.Body()))
}

func TestStructToProto(t *testing.T) {
	h := server.Default()
	h.POST("/api/structToProto", StructToProto)
	jsonStr := `{"structStr":"// test struct\ntype a struct {\n\t// primary key\n\tID uint64 \n\t// name\n\tName string\n\t// created time\n\tCreatedAt time.Time \n}"}`
	w := ut.PerformRequest(h.Engine, "POST", "/api/structToProto",
		&ut.Body{Body: bytes.NewBufferString(jsonStr), Len: len(jsonStr)},
		ut.Header{Key: "Connection", Value: "close"},
		ut.Header{Key: "Content-Type", Value: "application/json"})
	resp := w.Result()
	assert.DeepEqual(t, 200, resp.StatusCode())
	assert.DeepEqual(t, `{"errCode":0,"errMsg":"success","protoStr":"// test struct\nmessage a {\n  // primary key\n  uint64 ID = 1;\n  // name\n  string name = 2;\n  // created time\n  string createdAt = 3;\n}\n"}`,
		string(resp.Body()))
}

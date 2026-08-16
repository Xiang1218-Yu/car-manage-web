package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"carmanageweb/internal/models"
	"carmanageweb/internal/store"
)

func newFeeEstimateTestServer(t *testing.T) (*Server, int64) {
	t.Helper()
	st, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("打开内存数据库失败: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })

	ruleID, err := st.CreateFeeRule(models.FeeRule{
		Name: "预估测试规则", FreeMinutes: 0,
		FirstBlockMinutes: 60, FirstBlockPrice: 10,
		UnitMinutes: 30, UnitPrice: 5, Active: true,
	})
	if err != nil {
		t.Fatalf("创建测试规则失败: %v", err)
	}
	srv, err := New(st)
	if err != nil {
		t.Fatalf("创建 HTTP 服务失败: %v", err)
	}
	return srv, ruleID
}

func TestFeeEstimateRejectsMalformedCheckoutTime(t *testing.T) {
	srv, ruleID := newFeeEstimateTestServer(t)
	req := httptest.NewRequest(http.MethodGet,
		"/api/fee/estimate?rule_id="+strconv.FormatInt(ruleID, 10)+
			"&check_in=2026-08-16T10:00:00Z&check_out=not-a-time", nil)
	rec := httptest.NewRecorder()

	srv.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("无效 check_out 不应被当作当前时间继续计算，实际状态码 %d，响应 %s", rec.Code, rec.Body.String())
	}
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("错误响应不是 JSON: %v", err)
	}
	if body["error"] != "check_out 格式应为 RFC3339" {
		t.Fatalf("错误提示不正确，got %q", body["error"])
	}
}

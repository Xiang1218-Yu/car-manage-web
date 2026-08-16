package server

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

// writeJSON 写入 JSON 响应。
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeJSONError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func atoiOr(s string, def int) int {
	if v, err := strconv.Atoi(s); err == nil {
		return v
	}
	return def
}

func atofOr(s string, def float64) float64 {
	if v, err := strconv.ParseFloat(s, 64); err == nil {
		return v
	}
	return def
}

func zeroTime() time.Time { return time.Time{} }

func nowTime() time.Time { return time.Now().UTC() }

// parseQueryTime 解析 RFC3339 时间参数。
func parseQueryTime(r *http.Request, key string) (time.Time, error) {
	v := r.URL.Query().Get(key)
	if v == "" {
		return time.Time{}, &time.ParseError{}
	}
	return time.Parse(time.RFC3339, v)
}

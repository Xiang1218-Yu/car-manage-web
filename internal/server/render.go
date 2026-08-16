package server

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"net/url"
	"strings"
	"time"

	"carmanageweb/internal/models"
)

//go:embed all:static
var staticFS embed.FS

// staticSubFS 是 static 子目录的 fs.FS，作为 FileServer 的根，
// 使 /static/style.css 直接映射到 static/style.css。
var staticSubFS = func() fs.FS {
	sub, err := fs.Sub(staticFS, "static")
	if err != nil {
		panic(err)
	}
	return sub
}()

// render 渲染指定模板，自动套用 layout。模板名形如 "lots.html"。
func (s *Server) render(w http.ResponseWriter, r *http.Request, name string, data any) {
	// data 包装为含 flash 与导航上下文的视图模型
	vm := struct {
		Page    string
		Data    any
		Content template.HTML
		Flash   string
	}{
		Page: pageFromName(name),
		Data: data,
	}
	var buf strings.Builder
	if err := s.tpl.ExecuteTemplate(&buf, name, data); err != nil {
		http.Error(w, "渲染失败: "+err.Error(), http.StatusInternalServerError)
		return
	}
	vm.Content = template.HTML(buf.String())
	// 取 flash 后清除
	if f, ok := popFlash(w, r); ok {
		vm.Flash = f
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.tpl.ExecuteTemplate(w, "layout.html", vm); err != nil {
		http.Error(w, "渲染布局失败: "+err.Error(), http.StatusInternalServerError)
	}
}

// flash 通过 cookie 存储一次性消息，使用 URL 编码以支持中文与特殊字符。
func setFlash(w http.ResponseWriter, msg string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "flash",
		Value:    url.QueryEscape(msg),
		Path:     "/",
		MaxAge:   10,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func popFlash(w http.ResponseWriter, r *http.Request) (string, bool) {
	c, err := r.Cookie("flash")
	if err != nil {
		return "", false
	}
	// 清除 cookie
	http.SetCookie(w, &http.Cookie{Name: "flash", Path: "/", MaxAge: -1})
	if msg, err := url.QueryUnescape(c.Value); err == nil {
		return msg, true
	}
	return c.Value, true
}

// redirect 跳转并设置 flash。
func (s *Server) redirectWithFlash(w http.ResponseWriter, r *http.Request, target, msg string) {
	if msg != "" {
		setFlash(w, msg)
	}
	http.Redirect(w, r, target, http.StatusSeeOther)
}

func pageFromName(name string) string {
	switch {
	case strings.HasPrefix(name, "dashboard"):
		return "dashboard"
	case strings.HasPrefix(name, "lots"), strings.HasPrefix(name, "lot_detail"):
		return "lots"
	case strings.HasPrefix(name, "spots"):
		return "spots"
	case strings.HasPrefix(name, "vehicles"):
		return "vehicles"
	case strings.HasPrefix(name, "rules"):
		return "rules"
	case strings.HasPrefix(name, "records"):
		return "records"
	}
	return ""
}

// --- 模板函数 ---

func statusLabel(status interface{}) string {
	s := toStr(status)
	switch s {
	case string(models.SpotAvailable):
		return "空闲"
	case string(models.SpotOccupied):
		return "占用"
	case string(models.SpotMaintenance):
		return "维护中"
	case string(models.RecordActive):
		return "在场"
	case string(models.RecordCompleted):
		return "已离场"
	}
	return s
}

func statusClass(status interface{}) string {
	s := toStr(status)
	switch s {
	case string(models.SpotAvailable), string(models.RecordCompleted):
		return "ok"
	case string(models.SpotOccupied):
		return "warn"
	case string(models.SpotMaintenance):
		return "muted"
	case string(models.RecordActive):
		return "warn"
	}
	return ""
}

// toStr 把模板参数（可能是 string 或自定义 string 类型）转为普通字符串。
func toStr(v interface{}) string {
	switch x := v.(type) {
	case string:
		return x
	case models.SpotStatus:
		return string(x)
	case models.RecordStatus:
		return string(x)
	}
	return fmt.Sprintf("%v", v)
}

func formatRMB(v float64) string {
	return fmt.Sprintf("¥%.2f", v)
}

func yesno(b bool) string {
	if b {
		return "是"
	}
	return "否"
}

func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}

func fmtTime(t interface{}) string {
	var tt time.Time
	switch v := t.(type) {
	case time.Time:
		tt = v
	case *time.Time:
		if v == nil {
			return "—"
		}
		tt = *v
	case models.Time:
		tt = v.Std()
	case *models.Time:
		if v == nil {
			return "—"
		}
		tt = v.Std()
	default:
		return "—"
	}
	if tt.IsZero() {
		return "—"
	}
	return tt.Local().Format("2006-01-02 15:04")
}

// fmtDuration 计算从 checkIn 到 checkOut（或现在）的中文时长。
func fmtDuration(checkIn, checkOut interface{}) string {
	in := toStdTime(checkIn)
	if in.IsZero() {
		return "—"
	}
	out := toStdTime(checkOut)
	if out.IsZero() {
		out = time.Now()
	}
	if out.Before(in) {
		return "—"
	}
	d := out.Sub(in)
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	mins := int(d.Minutes()) % 60
	parts := []string{}
	if days > 0 {
		parts = append(parts, fmt.Sprintf("%d天", days))
	}
	if hours > 0 {
		parts = append(parts, fmt.Sprintf("%d小时", hours))
	}
	parts = append(parts, fmt.Sprintf("%d分钟", mins))
	return strings.Join(parts, " ")
}

// toStdTime 把模板传入的多种时间类型统一为 time.Time。
func toStdTime(t interface{}) time.Time {
	switch v := t.(type) {
	case time.Time:
		return v
	case *time.Time:
		if v == nil {
			return time.Time{}
		}
		return *v
	case models.Time:
		return v.Std()
	case *models.Time:
		if v == nil {
			return time.Time{}
		}
		return v.Std()
	}
	return time.Time{}
}

func lotScopeLabel(rule interface{}) string {
	r, ok := rule.(*models.FeeRule)
	if !ok {
		return ""
	}
	if r.LotID == nil {
		return "全局"
	}
	return fmt.Sprintf("停车场 #%d", *r.LotID)
}

func hasSuffix(s, suffix string) bool {
	return strings.HasSuffix(s, suffix)
}

func trimPrefix(s, prefix string) string {
	return strings.TrimPrefix(s, prefix)
}

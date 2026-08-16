// Package server 实现停车场管理系统的 HTTP 服务层。
package server

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"

	"carmanageweb/internal/fee"
	"carmanageweb/internal/store"
)

//go:embed templates/*.html
var templateFS embed.FS

// Server 持有存储、费用引擎与已解析模板。
type Server struct {
	store *store.Store
	fee   *fee.Calculator
	tpl   *template.Template
}

// New 创建并初始化 Server。
func New(s *store.Store) (*Server, error) {
	tpl, err := parseTemplates()
	if err != nil {
		return nil, err
	}
	return &Server{
		store: s,
		fee:   fee.New(),
		tpl:   tpl,
	}, nil
}

func parseTemplates() (*template.Template, error) {
	funcMap := template.FuncMap{
		"statusLabel":   statusLabel,
		"statusClass":   statusClass,
		"rmb":           formatRMB,
		"yesno":         yesno,
		"truncate":      truncate,
		"fmtTime":       fmtTime,
		"fmtDuration":   fmtDuration,
		"lotScopeLabel": lotScopeLabel,
	}
	tpl := template.New("").Funcs(funcMap)
	err := fs.WalkDir(templateFS, "templates", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !hasSuffix(path, ".html") {
			return nil
		}
		data, err := templateFS.ReadFile(path)
		if err != nil {
			return err
		}
		name := trimPrefix(path, "templates/")
		_, err = tpl.New(name).Parse(string(data))
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("解析模板: %w", err)
	}
	return tpl, nil
}

// Routes 注册所有路由。使用 Go 1.22+ 的 http.ServeMux 方法与路径通配。
func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	// 静态资源
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticSubFS))))

	// 页面
	mux.HandleFunc("GET /", s.handleDashboard)
	mux.HandleFunc("GET /lots", s.handleLotsPage)
	mux.HandleFunc("GET /lots/{id}", s.handleLotDetailPage)
	mux.HandleFunc("GET /spots", s.handleSpotsPage)
	mux.HandleFunc("GET /vehicles", s.handleVehiclesPage)
	mux.HandleFunc("GET /rules", s.handleRulesPage)
	mux.HandleFunc("GET /records", s.handleRecordsPage)

	// 停车场 CRUD
	mux.HandleFunc("POST /lots", s.handleCreateLot)
	mux.HandleFunc("POST /lots/{id}/update", s.handleUpdateLot)
	mux.HandleFunc("POST /lots/{id}/delete", s.handleDeleteLot)

	// 车位 CRUD
	mux.HandleFunc("POST /spots", s.handleCreateSpot)
	mux.HandleFunc("POST /spots/{id}/update", s.handleUpdateSpot)
	mux.HandleFunc("POST /spots/{id}/delete", s.handleDeleteSpot)

	// 车辆
	mux.HandleFunc("POST /vehicles", s.handleUpsertVehicle)
	mux.HandleFunc("POST /vehicles/{id}/delete", s.handleDeleteVehicle)

	// 费用规则 CRUD
	mux.HandleFunc("POST /rules", s.handleCreateRule)
	mux.HandleFunc("POST /rules/{id}/update", s.handleUpdateRule)
	mux.HandleFunc("POST /rules/{id}/delete", s.handleDeleteRule)

	// 入场/出场
	mux.HandleFunc("POST /records/checkin", s.handleCheckIn)
	mux.HandleFunc("POST /records/{id}/checkout", s.handleCheckOut)
	mux.HandleFunc("POST /records/{id}/delete", s.handleDeleteRecord)

	// 费用预估（AJAX）
	mux.HandleFunc("GET /api/fee/estimate", s.handleFeeEstimate)

	// 仪表盘统计（AJAX）
	mux.HandleFunc("GET /api/stats", s.handleStatsAPI)

	return logRequests(mux)
}

// logRequests 简单的请求日志中间件。
func logRequests(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 生产环境可用结构化日志；此处简化
		h.ServeHTTP(w, r)
	})
}

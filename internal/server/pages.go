package server

import (
	"net/http"
	"strconv"
)

// handleDashboard 仪表盘：展示统计与最近记录。
func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	stats, err := s.store.Stats()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	recent, err := s.store.ListRecords("")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(recent) > 8 {
		recent = recent[:8]
	}
	s.render(w, r, "dashboard.html", map[string]any{
		"Stats":   stats,
		"Recent":  recent,
	})
}

// handleLotsPage 停车场列表页。
func (s *Server) handleLotsPage(w http.ResponseWriter, r *http.Request) {
	lots, err := s.store.ListLots()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// 统计每个停车场的车位占用
	type lotRow struct {
		Lot          interface{}
		SpotCount    int
		Available    int
		Occupied     int
	}
	rows := make([]lotRow, 0, len(lots))
	for _, l := range lots {
		spots, _ := s.store.ListSpotsByLot(l.ID)
		av, oc := 0, 0
		for _, sp := range spots {
			switch sp.Status {
			case "available":
				av++
			case "occupied":
				oc++
			}
		}
		rows = append(rows, lotRow{Lot: l, SpotCount: len(spots), Available: av, Occupied: oc})
	}
	s.render(w, r, "lots.html", map[string]any{"Rows": rows})
}

// handleLotDetailPage 单个停车场详情：含车位网格与入场表单。
func (s *Server) handleLotDetailPage(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "无效的 ID", http.StatusBadRequest)
		return
	}
	lot, err := s.store.GetLot(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	spots, err := s.store.ListSpotsByLot(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	vehicles, _ := s.store.ListVehicles()
	s.render(w, r, "lot_detail.html", map[string]any{
		"Lot":      lot,
		"Spots":    spots,
		"Vehicles": vehicles,
	})
}

// handleSpotsPage 所有车位列表页。
func (s *Server) handleSpotsPage(w http.ResponseWriter, r *http.Request) {
	lots, _ := s.store.ListLots()
	type spotView struct {
		LotName string
		Spots   interface{}
	}
	views := []spotView{}
	for _, l := range lots {
		spots, _ := s.store.ListSpotsByLot(l.ID)
		if len(spots) > 0 {
			views = append(views, spotView{LotName: l.Name, Spots: spots})
		}
	}
	s.render(w, r, "spots.html", map[string]any{"Views": views, "Lots": lots})
}

// handleVehiclesPage 车辆列表页。
func (s *Server) handleVehiclesPage(w http.ResponseWriter, r *http.Request) {
	vehicles, err := s.store.ListVehicles()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.render(w, r, "vehicles.html", map[string]any{"Vehicles": vehicles})
}

// handleRulesPage 费用规则列表页。
func (s *Server) handleRulesPage(w http.ResponseWriter, r *http.Request) {
	rules, err := s.store.ListFeeRules()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	lots, _ := s.store.ListLots()
	// 给规则附加停车场名称
	type ruleView struct {
		Rule    interface{}
		LotName string
	}
	views := make([]ruleView, 0, len(rules))
	lotMap := map[int64]string{}
	for _, l := range lots {
		lotMap[l.ID] = l.Name
	}
	for _, rl := range rules {
		ln := "全局"
		if rl.LotID != nil {
			ln = lotMap[*rl.LotID]
		}
		views = append(views, ruleView{Rule: rl, LotName: ln})
	}
	s.render(w, r, "rules.html", map[string]any{"Views": views, "Lots": lots})
}

// handleRecordsPage 停车记录页，支持 status 过滤。
func (s *Server) handleRecordsPage(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	records, err := s.store.ListRecords(status)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.render(w, r, "records.html", map[string]any{
		"Records":    records,
		"CurStatus":  status,
	})
}

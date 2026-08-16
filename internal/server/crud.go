package server

import (
	"net/http"
	"strconv"
	"strings"

	"carmanageweb/internal/models"
)

// --- 停车场 ---

func (s *Server) handleCreateLot(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimSpace(r.FormValue("name"))
	address := strings.TrimSpace(r.FormValue("address"))
	total, _ := strconv.Atoi(r.FormValue("total_spots"))
	if name == "" {
		s.redirectWithFlash(w, r, "/lots", "停车场名称不能为空")
		return
	}
	if total < 0 {
		total = 0
	}
	id, err := s.store.CreateLot(name, address, total)
	if err != nil {
		s.redirectWithFlash(w, r, "/lots", "创建失败: "+err.Error())
		return
	}
	s.redirectWithFlash(w, r, "/lots", "停车场已创建 (ID="+strconv.FormatInt(id, 10)+")")
}

func (s *Server) handleUpdateLot(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "无效的 ID", http.StatusBadRequest)
		return
	}
	name := strings.TrimSpace(r.FormValue("name"))
	address := strings.TrimSpace(r.FormValue("address"))
	total, _ := strconv.Atoi(r.FormValue("total_spots"))
	if err := s.store.UpdateLot(id, name, address, total); err != nil {
		s.redirectWithFlash(w, r, "/lots", "更新失败: "+err.Error())
		return
	}
	s.redirectWithFlash(w, r, "/lots", "停车场已更新")
}

func (s *Server) handleDeleteLot(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "无效的 ID", http.StatusBadRequest)
		return
	}
	if err := s.store.DeleteLot(id); err != nil {
		s.redirectWithFlash(w, r, "/lots", "删除失败: "+err.Error())
		return
	}
	s.redirectWithFlash(w, r, "/lots", "停车场已删除")
}

// --- 车位 ---

func (s *Server) handleCreateSpot(w http.ResponseWriter, r *http.Request) {
	lotID, err := strconv.ParseInt(r.FormValue("lot_id"), 10, 64)
	if err != nil {
		s.redirectWithFlash(w, r, "/spots", "请选择停车场")
		return
	}
	code := strings.TrimSpace(r.FormValue("code"))
	status := models.SpotStatus(r.FormValue("status"))
	if status == "" {
		status = models.SpotAvailable
	}
	id, err := s.store.CreateSpot(lotID, code, status)
	if err != nil {
		s.redirectWithFlash(w, r, "/spots", "创建车位失败: "+err.Error())
		return
	}
	s.redirectWithFlash(w, r, "/spots", "车位已创建 (ID="+strconv.FormatInt(id, 10)+")")
}

func (s *Server) handleUpdateSpot(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "无效的 ID", http.StatusBadRequest)
		return
	}
	code := strings.TrimSpace(r.FormValue("code"))
	status := models.SpotStatus(r.FormValue("status"))
	if err := s.store.UpdateSpot(id, code, status); err != nil {
		s.redirectWithFlash(w, r, "/spots", "更新失败: "+err.Error())
		return
	}
	s.redirectWithFlash(w, r, "/spots", "车位已更新")
}

func (s *Server) handleDeleteSpot(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "无效的 ID", http.StatusBadRequest)
		return
	}
	if err := s.store.DeleteSpot(id); err != nil {
		s.redirectWithFlash(w, r, "/spots", "删除失败: "+err.Error())
		return
	}
	s.redirectWithFlash(w, r, "/spots", "车位已删除")
}

// --- 车辆 ---

func (s *Server) handleUpsertVehicle(w http.ResponseWriter, r *http.Request) {
	v := models.Vehicle{
		Plate:       r.FormValue("plate"),
		VehicleType: r.FormValue("vehicle_type"),
		Color:       strings.TrimSpace(r.FormValue("color")),
		OwnerName:   strings.TrimSpace(r.FormValue("owner_name")),
		OwnerPhone:  strings.TrimSpace(r.FormValue("owner_phone")),
	}
	id, err := s.store.UpsertVehicle(v)
	if err != nil {
		s.redirectWithFlash(w, r, "/vehicles", "保存失败: "+err.Error())
		return
	}
	s.redirectWithFlash(w, r, "/vehicles", "车辆已保存 (ID="+strconv.FormatInt(id, 10)+")")
}

func (s *Server) handleDeleteVehicle(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "无效的 ID", http.StatusBadRequest)
		return
	}
	if err := s.store.DeleteVehicle(id); err != nil {
		s.redirectWithFlash(w, r, "/vehicles", "删除失败: "+err.Error())
		return
	}
	s.redirectWithFlash(w, r, "/vehicles", "车辆已删除")
}

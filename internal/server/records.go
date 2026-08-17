package server

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"carmanageweb/internal/models"
	"carmanageweb/internal/store"
)

// --- 费用规则 ---

func (s *Server) handleCreateRule(w http.ResponseWriter, r *http.Request) {
	rule, err := parseRuleForm(r)
	if err != nil {
		s.redirectWithFlash(w, r, "/rules", "创建失败: "+err.Error())
		return
	}
	id, err := s.store.CreateFeeRule(rule)
	if err != nil {
		s.redirectWithFlash(w, r, "/rules", "创建失败: "+err.Error())
		return
	}
	s.redirectWithFlash(w, r, "/rules", "费用规则已创建 (ID="+strconv.FormatInt(id, 10)+")")
}

func (s *Server) handleUpdateRule(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "无效的 ID", http.StatusBadRequest)
		return
	}
	rule, err := parseRuleForm(r)
	if err != nil {
		s.redirectWithFlash(w, r, "/rules", "更新失败: "+err.Error())
		return
	}
	if err := s.store.UpdateFeeRule(id, rule); err != nil {
		s.redirectWithFlash(w, r, "/rules", "更新失败: "+err.Error())
		return
	}
	s.redirectWithFlash(w, r, "/rules", "费用规则已更新")
}

func (s *Server) handleDeleteRule(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "无效的 ID", http.StatusBadRequest)
		return
	}
	if err := s.store.DeleteFeeRule(id); err != nil {
		s.redirectWithFlash(w, r, "/rules", "删除失败: "+err.Error())
		return
	}
	s.redirectWithFlash(w, r, "/rules", "费用规则已删除")
}

// parseRuleForm 从表单解析费用规则。lot_id 为空表示全局规则。
func parseRuleForm(r *http.Request) (models.FeeRule, error) {
	rule := models.FeeRule{
		Name:              strings.TrimSpace(r.FormValue("name")),
		FreeMinutes:       atoiOr(r.FormValue("free_minutes"), 0),
		FirstBlockMinutes: atoiOr(r.FormValue("first_block_minutes"), 60),
		FirstBlockPrice:   atofOr(r.FormValue("first_block_price"), 0),
		UnitMinutes:       atoiOr(r.FormValue("unit_minutes"), 60),
		UnitPrice:         atofOr(r.FormValue("unit_price"), 0),
		DailyCap:          atofOr(r.FormValue("daily_cap"), 0),
		Active:            r.FormValue("active") == "1",
	}
	if lotStr := strings.TrimSpace(r.FormValue("lot_id")); lotStr != "" {
		if id, err := strconv.ParseInt(lotStr, 10, 64); err == nil && id > 0 {
			rule.LotID = &id
		}
	}
	return rule, nil
}

// --- 入场 / 出场 ---

func (s *Server) handleCheckIn(w http.ResponseWriter, r *http.Request) {
	spotID, err := strconv.ParseInt(r.FormValue("spot_id"), 10, 64)
	if err != nil {
		s.redirectWithFlash(w, r, "/records", "请选择车位")
		return
	}
	plate := strings.TrimSpace(r.FormValue("plate"))
	if plate == "" {
		// 回到来源页
		ref := r.FormValue("from")
		if ref == "" {
			ref = "/spots"
		}
		s.redirectWithFlash(w, r, ref, "车牌号不能为空")
		return
	}
	vehicleID, err := s.store.UpsertVehicle(models.Vehicle{Plate: plate})
	if err != nil {
		s.redirectWithFlash(w, r, "/records", "车辆保存失败: "+err.Error())
		return
	}
	recID, rule, err := s.store.CheckIn(spotID, vehicleID)
	if err != nil {
		ref := r.FormValue("from")
		if ref == "" {
			ref = "/spots"
		}
		var activeErr *store.VehicleAlreadyParkedError
		if errors.As(err, &activeErr) {
			s.redirectWithFlash(w, r, ref, "入场失败: "+activeErr.Error())
			return
		}
		s.redirectWithFlash(w, r, ref, "入场失败: "+err.Error())
		return
	}
	msg := "车辆 " + plate + " 已入场 (记录 #" + strconv.FormatInt(recID, 10) + ")"
	if rule != nil {
		msg += "，适用规则「" + rule.Name + "」"
	}
	s.redirectWithFlash(w, r, "/records", msg)
}

func (s *Server) handleCheckOut(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "无效的 ID", http.StatusBadRequest)
		return
	}
	fee, bd, err := s.store.CheckOut(id, zeroTime(), s.fee.Calc)
	if err != nil {
		s.redirectWithFlash(w, r, "/records", "出场失败: "+err.Error())
		return
	}
	msg := "已出场结算，费用 ¥" + strconv.FormatFloat(fee, 'f', 2, 64)
	if bd != nil {
		msg += "（时长 " + strconv.Itoa(bd.DurationMinutes) + " 分钟）"
	}
	s.redirectWithFlash(w, r, "/records", msg)
}

// handleDeleteRecord 删除停车记录（仅允许删除已完成的记录）。
func (s *Server) handleDeleteRecord(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "无效的 ID", http.StatusBadRequest)
		return
	}
	if err := s.store.DeleteRecord(id); err != nil {
		s.redirectWithFlash(w, r, "/records", "删除失败: "+err.Error())
		return
	}
	s.redirectWithFlash(w, r, "/records", "记录已删除")
}

// --- API ---

// handleFeeEstimate 费用预估接口，GET /api/fee/estimate?rule_id=&check_in=&check_out=
func (s *Server) handleFeeEstimate(w http.ResponseWriter, r *http.Request) {
	ruleID, err := strconv.ParseInt(r.URL.Query().Get("rule_id"), 10, 64)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "缺少或非法的 rule_id")
		return
	}
	rule, err := s.store.GetFeeRule(ruleID)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "费用规则不存在")
		return
	}
	checkIn, err := parseQueryTime(r, "check_in")
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "check_in 格式应为 RFC3339")
		return
	}
	checkOut, err := parseQueryTime(r, "check_out")
	if err != nil {
		checkOut = nowTime()
	}
	bd := s.fee.Calc(checkIn, checkOut, rule)
	writeJSON(w, http.StatusOK, bd)
}

// handleStatsAPI 仪表盘统计 JSON。
func (s *Server) handleStatsAPI(w http.ResponseWriter, r *http.Request) {
	stats, err := s.store.Stats()
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, stats)
}

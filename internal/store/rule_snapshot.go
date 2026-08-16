package store

import (
	"encoding/json"
	"fmt"

	"carmanageweb/internal/models"
)

func encodeFeeRuleSnapshot(rule *models.FeeRule) (string, error) {
	snapshot := models.NewFeeRuleSnapshot(rule)
	if snapshot == nil {
		return "", fmt.Errorf("不能为 nil 费用规则创建快照")
	}
	raw, err := json.Marshal(snapshot)
	if err != nil {
		return "", fmt.Errorf("编码费用规则快照: %w", err)
	}
	return string(raw), nil
}

func decodeFeeRuleSnapshot(raw string) (*models.FeeRule, *models.FeeRuleSnapshot, error) {
	if raw == "" {
		return nil, nil, nil
	}
	var snapshot models.FeeRuleSnapshot
	if err := json.Unmarshal([]byte(raw), &snapshot); err != nil {
		return nil, nil, fmt.Errorf("解析费用规则快照: %w", err)
	}
	rule := snapshot.FeeRule()
	if rule.FirstBlockMinutes <= 0 || rule.UnitMinutes <= 0 {
		return nil, nil, fmt.Errorf("费用规则快照的计费时长无效")
	}
	return rule, &snapshot, nil
}

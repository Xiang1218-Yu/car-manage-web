package models

import (
	"database/sql/driver"
	"fmt"
	"time"
)

// Time 是可被 database/sql 扫描的时间类型，行为类似 time.Time。
// SQLite 中时间以 RFC3339 字符串存储，Time 实现 Scanner/Valuer 完成转换，
// 避免在 store 层每处手动解析。JSON 序列化输出 RFC3339 字符串。
type Time time.Time

// Scan 实现 sql.Scanner：把 SQLite 返回的字符串解析为 time.Time。
func (t *Time) Scan(value interface{}) error {
	switch v := value.(type) {
	case nil:
		*t = Time(time.Time{})
		return nil
	case string:
		if v == "" {
			*t = Time(time.Time{})
			return nil
		}
		parsed, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return fmt.Errorf("解析时间 %q: %w", v, err)
		}
		*t = Time(parsed)
		return nil
	case time.Time:
		*t = Time(v)
		return nil
	}
	return fmt.Errorf("无法扫描 %T 到 models.Time", value)
}

// Value 实现 driver.Valuer，写入时格式化为 RFC3339。
func (t Time) Value() (driver.Value, error) {
	return time.Time(t).UTC().Format(time.RFC3339), nil
}

// MarshalJSON 输出 RFC3339 字符串，零值为空。
func (t Time) MarshalJSON() ([]byte, error) {
	tt := time.Time(t)
	if tt.IsZero() {
		return []byte(`""`), nil
	}
	return []byte(`"` + tt.UTC().Format(time.RFC3339) + `"`), nil
}

// 以下方法代理到 time.Time，供模板函数与业务逻辑使用。

func (t Time) IsZero() bool            { return time.Time(t).IsZero() }
func (t Time) Local() time.Time         { return time.Time(t).Local() }
func (t Time) UTC() time.Time          { return time.Time(t).UTC() }
func (t Time) Std() time.Time           { return time.Time(t) }
func (t Time) Format(layout string) string { return time.Time(t).Format(layout) }
func (t Time) Sub(u Time) time.Duration    { return time.Time(t).Sub(time.Time(u)) }
func (t Time) Before(u Time) bool          { return time.Time(t).Before(time.Time(u)) }
func (t Time) After(u Time) bool           { return time.Time(t).After(time.Time(u)) }
func (t Time) Equal(u Time) bool           { return time.Time(t).Equal(time.Time(u)) }
func (t Time) Add(d time.Duration) Time    { return Time(time.Time(t).Add(d)) }
func (t Time) AddDate(years, months, days int) Time {
	return Time(time.Time(t).AddDate(years, months, days))
}
func (t Time) Year() int        { return time.Time(t).Year() }
func (t Time) Month() time.Month { return time.Time(t).Month() }
func (t Time) Day() int         { return time.Time(t).Day() }
func (t Time) Hour() int        { return time.Time(t).Hour() }
func (t Time) Minute() int      { return time.Time(t).Minute() }
func (t Time) Location() *time.Location { return time.Time(t).Location() }
func (t Time) In(loc *time.Location) Time { return Time(time.Time(t).In(loc)) }

// Date 构造一个 Time（对应 time.Date）。
func NewDate(year int, month time.Month, day, hour, min, sec, nsec int, loc *time.Location) Time {
	return Time(time.Date(year, month, day, hour, min, sec, nsec, loc))
}

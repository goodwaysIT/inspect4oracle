// 此文件包含用于格式化各种sql.Null*类型的辅助函数。
package handler

import (
	"database/sql"
	"fmt"
)

// formatNullTime 辅助函数，用于格式化 sql.NullTime
func formatNullTime(nt sql.NullTime, layout string) string {
	if nt.Valid {
		return nt.Time.Format(layout)
	}
	return ""
}

// formatNullString 辅助函数，用于格式化 sql.NullString
func formatNullString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

// formatNullInt64 辅助函数，用于格式化 sql.NullInt64
// (已存在于 module_processor.go, 如果要统一管理则移到此处)
func formatNullInt64(ni sql.NullInt64) string {
	if ni.Valid {
		return fmt.Sprintf("%d", ni.Int64)
	}
	return ""
}

// formatNullFloat64 辅助函数，用于格式化 sql.NullFloat64
// (已存在于 module_processor.go, 如果要统一管理则移到此处)
func formatNullFloat64(nf sql.NullFloat64, format string) string {
	if nf.Valid {
		return fmt.Sprintf(format, nf.Float64)
	}
	return ""
}

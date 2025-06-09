package db

import (
	"database/sql"
	"fmt"
)

type ParameterInfo struct {
	Name  string
	Value sql.NullString
}

// GetParameterList 查询数据库参数列表
func GetParameterList(db *sql.DB) ([]ParameterInfo, error) {
	query := `SELECT NAME, VALUE FROM V$PARAMETER WHERE ISDEFAULT='FALSE'`

	var result []ParameterInfo
	err := ExecuteQueryAndScanToStructs(db, &result, query)
	if err != nil {
		return nil, fmt.Errorf("GetParameterList 使用 ExecuteQueryAndScanToStructs 失败: %w", err)
	}

	return result, nil
}

package db

import (
	"database/sql"
	"fmt"

	"reflect"
	"strings"

	"github.com/goodwaysIT/inspect4oracle/internal/logger" // 假设 logger 包路径
)

// ExecuteGenericQuery 执行一个通用的SQL查询并返回结果。
// It returns a slice of maps, where each map represents a row with column names as keys.
// It also returns a slice of column names and any error that occurred.
func ExecuteGenericQuery(db *sql.DB, query string, args ...interface{}) ([]map[string]interface{}, []string, error) {
	rows, err := db.Query(query, args...)
	if err != nil {
		logger.Errorf("Failed to execute query: %s, error: %v", query, err)
		return nil, nil, fmt.Errorf("failed to execute query '%s': %w", query, err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		logger.Errorf("Failed to get column names for query: %s, error: %v", query, err)
		return nil, nil, fmt.Errorf("failed to get column names for query '%s': %w", query, err)
	}

	var results []map[string]interface{}
	for rows.Next() {
		// 创建一个与列数相同的 interface{} 切片来接收值
		values := make([]interface{}, len(columns))
		// 创建一个 interface{} 切片，其元素将是指向 `values` 中元素的指针
		scanDest := make([]interface{}, len(columns))
		for i := range values {
			scanDest[i] = &values[i] // 每个元素都是一个指针
		}

		err = rows.Scan(scanDest...)
		if err != nil {
			logger.Warnf("Failed to scan row data: %s, error: %v. Skipping this row.", query, err)
			continue // 或者根据需要决定是否中止整个查询
		}

		rowData := make(map[string]interface{})
		for i, col := range columns {
			// 处理可能的 nil 值 (数据库 NULL)
			if values[i] == nil {
				rowData[col] = nil
			} else {
				// 尝试将 []byte (通常是字符串或数字的原始表示) 转换为 string
				// 其他类型如 int64, float64, time.Time 会被驱动正确处理
				if b, ok := values[i].([]byte); ok {
					rowData[col] = string(b)
				} else {
					rowData[col] = values[i]
				}
			}
		}
		results = append(results, rowData)
	}

	if err = rows.Err(); err != nil {
		logger.Errorf("Error while iterating over result set: %s, error: %v", query, err)
		return results, columns, fmt.Errorf("error iterating over result set for query '%s': %w", query, err)
	}

	logger.Debugf("Generic query executed successfully: %s, returned %d rows", query, len(results))
	return results, columns, nil
}

// ConvertRowToStruct 将 ExecuteGenericQuery 返回的单行结果 (map[string]interface{})
// 转换为指定的结构体。这只是一个辅助函数示例，实际应用中可能需要更复杂的反射或手动映射。
// Note: This function is very basic and has no error handling or complex type conversion.
// 在实际应用中，你可能需要根据具体需求调整或使用更健壮的库（如 sqlx）。
func ConvertRowToStruct(row map[string]interface{}, targetStruct interface{}) error {
	// 这是一个非常简化的示例，实际中你可能需要使用反射
	// or manually map to struct fields based on column names.
	// 例如，如果 targetStruct 是 *ParameterInfo:
	// if p, ok := targetStruct.(*ParameterInfo); ok {
	//     if name, ok := row["NAME"].(string); ok { p.Name = name }
	//     if value, ok := row["VALUE"].(string); ok { p.Value = value }
	// } else {
	//     return fmt.Errorf("目标结构体类型不匹配")
	// }
	// Since direct generic struct conversion is complex and error-prone (especially with type conversion and field name matching),
	// 通常建议在调用 ExecuteGenericQuery 后，在具体的查询函数中手动处理 map 到 struct 的转换。
	logger.Warn("ConvertRowToStruct is a basic example. It is recommended to manually convert map to struct in the caller for better control and error handling.")
	return fmt.Errorf("ConvertRowToStruct 功能受限，建议手动转换")
}

// findScanDestination attempts to find a matching field in a struct for a given column name.
// It first checks for a 'db' tag on struct fields, then falls back to case-insensitive field name matching.
// Returns the addressable interface of the field if found, and a boolean indicating success.
func findScanDestination(colName string, structType reflect.Type, structVal reflect.Value) (interface{}, bool) {
	upperColName := strings.ToUpper(colName)
	for j := 0; j < structVal.NumField(); j++ {
		fieldDesc := structType.Field(j) // StructField descriptor
		dbTag := fieldDesc.Tag.Get("db") // Get the value of the "db" tag

		matchedByTag := false
		if dbTag != "" {
			if strings.ToUpper(dbTag) == upperColName {
				matchedByTag = true
			}
		}

		matchedByName := false
		if !matchedByTag { // Only try to match by name if not matched by tag
			if strings.ToUpper(fieldDesc.Name) == upperColName {
				matchedByName = true
			}
		}

		if matchedByTag || matchedByName {
			fieldVal := structVal.Field(j)
			if fieldVal.CanAddr() {
				// Log which way it was matched, for debugging
				// if matchedByTag {
				// 	logger.Debugf("列 '%s' 匹配到字段 '%s' 通过 db 标签 '%s'", colName, fieldDesc.Name, dbTag)
				// } else {
				// 	logger.Debugf("列 '%s' 匹配到字段 '%s' 通过字段名", colName, fieldDesc.Name)
				// }
				return fieldVal.Addr().Interface(), true
			} else {
				logger.Warnf("字段 '%s' (列 '%s') 不可寻址，跳过扫描此列", fieldDesc.Name, colName)
			}
		}
	}
	return new(interface{}), false // Return a dummy scan target if no field is found
}

// ExecuteQueryAndScanToStructs executes a query and scans the results directly into a slice of structs.
// - db: The database connection.
// - destSlice: A pointer to a slice of structs (e.g., *[]MyStruct) where results will be stored.
// - query: The SQL query string.
// - args: Arguments for the query.
// This function uses reflection and maps columns to struct fields by comparing their uppercase names.
func ExecuteQueryAndScanToStructs(db *sql.DB, destSlice interface{}, query string, args ...interface{}) error {
	destVal := reflect.ValueOf(destSlice)
	if destVal.Kind() != reflect.Ptr {
		return fmt.Errorf("destSlice must be a pointer to a slice, got %T", destSlice)
	}
	sliceVal := destVal.Elem()
	if sliceVal.Kind() != reflect.Slice {
		return fmt.Errorf("destSlice must point to a slice, got %s", sliceVal.Kind())
	}

	structType := sliceVal.Type().Elem() // Get the type of the elements in the slice (e.g., MyStruct)
	if structType.Kind() != reflect.Struct {
		return fmt.Errorf("slice elements must be structs, got %s", structType.Kind())
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		logger.Errorf("Failed to execute query: %s, error: %v", query, err)
		return fmt.Errorf("failed to execute query '%s': %w", query, err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		logger.Errorf("Failed to get column names for query: %s, error: %v", query, err)
		return fmt.Errorf("Failed to get column names for query '%s': %w", query, err)
	}

	for rows.Next() {
		// Create a new instance of the struct type (e.g., a new MyStruct)
		newStructPtr := reflect.New(structType) // This is a pointer to the new struct (*MyStruct)
		newStructVal := newStructPtr.Elem()     // This is the struct value itself (MyStruct)

		scanDest := make([]interface{}, len(columns))
		for i, colName := range columns {
			var foundField bool
			scanDest[i], foundField = findScanDestination(colName, structType, newStructVal)
			if !foundField {
				logger.Debugf("列 '%s' 在目标结构体 '%s' 中没有匹配的字段或字段不可寻址，将扫描到临时变量", colName, structType.Name())
				// findScanDestination already returns new(interface{}) in this case
			}
		}

		if err := rows.Scan(scanDest...); err != nil {
			logger.Warnf("Failed to scan row data (query: %s): %v. Skipping this row.", query, err)
			continue // Skip this row if scanning fails
		}
		// Append the new, populated struct to the destination slice
		sliceVal.Set(reflect.Append(sliceVal, newStructVal))
	}

	if err = rows.Err(); err != nil {
		logger.Errorf("Error while iterating over result set: %s, error: %v", query, err)
		return fmt.Errorf("error iterating over result set for query '%s': %w", query, err)
	}

	logger.Debugf("Generic struct query executed successfully: %s, populated %d structs", query, sliceVal.Len())
	return nil
}

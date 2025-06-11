package db

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/goodwaysIT/inspect4oracle/internal/logger"

	go_ora "github.com/sijms/go-ora/v2"
)

// ConnectionDetails holds all necessary information for connecting to Oracle DB.
// This can be expanded later if more specific go-ora parameters are needed.
type ConnectionDetails struct {
	User           string
	Password       string
	Host           string
	Port           int
	DBName         string // SID or Service Name
	ConnectionType string // "SID" or "SERVICE_NAME"
}

// Connect establishes a connection to the Oracle database using the provided details.
// It returns a sql.DB object or an error if the connection fails.
func Connect(details ConnectionDetails) (*sql.DB, error) {
	// Set connection timeout to 30 seconds
	urlOptions := map[string]string{
		"CONNECTION TIMEOUT": "30",
	}

	// 使用 go-ora 的 BuildUrl 构建连接字符串
	connStr := go_ora.BuildUrl(
		details.Host,
		details.Port,
		details.DBName, // SID or Service Name
		details.User,
		details.Password,
		urlOptions,
	)

	// 使用 sijms/go-ora/v2 驱动打开连接
	db, err := sql.Open("oracle", connStr)
	if err != nil {
		return nil, fmt.Errorf("error opening database connection: %w", err)
	}

	err = db.Ping()
	if err != nil {
		db.Close() // Close the connection if ping fails
		return nil, fmt.Errorf("error pinging database: %w", err)
	}

	return db, nil
}

// PrivilegeCheckResult 表示权限检查的结果
type PrivilegeCheckResult struct {
	ViewName  string `json:"view_name"`
	HasAccess bool   `json:"has_access"`
	Error     string `json:"error,omitempty"`
}

// ValidatePrivileges 验证数据库连接是否具有查询关键系统视图的权限
func ValidatePrivileges(db *sql.DB) ([]PrivilegeCheckResult, error) {
	// 关键系统视图列表
	criticalViews := []string{
		// v$ views
		"v$active_session_history",
		"v$asm_diskgroup", // If ASM is used and needs checking
		"v$database",
		"v$instance",
		"v$session",
		"v$sql",
		"v$sqlarea",
		"v$sysmetric",
		"v$system_parameter",
		"v$temp_extent_pool",
		"v$version",
		// dba_ views
		"dba_data_files",
		"dba_free_space",
		"dba_objects",
		"dba_roles",
		"dba_role_privs",
		"dba_segments",
		"dba_sys_privs",
		"dba_tablespaces",
		"dba_temp_files",
		"dba_users",
	}

	results := make([]PrivilegeCheckResult, 0, len(criticalViews))

	// 检查每个视图的查询权限
	for _, view := range criticalViews {
		result := PrivilegeCheckResult{
			ViewName:  view,
			HasAccess: false,
		}

		// 构建查询语句
		query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE ROWNUM = 1", view)

		// 尝试执行查询
		var count int
		err := db.QueryRow(query).Scan(&count)

		if err != nil {
			// Check the error type, if it's insufficient privileges, log it and continue
			if strings.Contains(strings.ToUpper(err.Error()), "ORA-00942") || // 表或视图不存在
				strings.Contains(strings.ToUpper(err.Error()), "ORA-01031") { // 权限不足
				logger.Debug(fmt.Sprintf("Permission check failed for view '%s': %s", view, err.Error()))
				result.Error = fmt.Sprintf("视图 '%s' 权限不足或对象不存在", view)
			} else {
				logger.Debug(fmt.Sprintf("Permission check failed: %s", err.Error()))
				result.Error = err.Error()
			}
		} else {
			result.HasAccess = true
		}

		results = append(results, result)
	}

	return results, nil
}

// CheckDatabaseConnection 验证数据库连接并检查权限
func CheckDatabaseConnection(db *sql.DB) (bool, []PrivilegeCheckResult, error) {
	// 首先验证连接是否有效
	err := db.Ping()
	if err != nil {
		return false, nil, fmt.Errorf("database connection failed: %v", err)
	}

	// 检查权限
	privilegeResults, err := ValidatePrivileges(db)
	if err != nil {
		return true, privilegeResults, fmt.Errorf("permission check failed: %v", err)
	}

	// 检查是否有任何关键视图没有访问权限
	allAccessGranted := true
	for _, result := range privilegeResults {
		if !result.HasAccess {
			allAccessGranted = false
			break
		}
	}

	return allAccessGranted, privilegeResults, nil
}

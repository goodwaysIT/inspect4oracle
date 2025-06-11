package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/goodwaysIT/inspect4oracle/internal/db"
	"github.com/goodwaysIT/inspect4oracle/internal/logger"
)

// ValidateRequest defines the validation request structure
type ValidateRequest struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	Service  string `json:"service"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// ValidateResponse 定义验证响应结构体
type ValidateResponse struct {
	Success        bool                      `json:"success"`
	Message        string                    `json:"message"`
	PrivilegeCheck []db.PrivilegeCheckResult `json:"privilege_check,omitempty"`
}

// ValidateConnection 验证数据库连接
func ValidateConnection(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse JSON request body
	var reqData ValidateRequest
	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		logger.Error(fmt.Sprintf("Failed to decode request body: %s", err.Error()))
		sendJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// 验证必填字段
	if reqData.Host == "" || reqData.Username == "" || reqData.Password == "" || reqData.Service == "" {
		sendJSONError(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Set default port if not provided
	port := reqData.Port
	if port == "" {
		port = "1521"
	}

	// 验证端口号
	portInt, err := strconv.Atoi(port)
	if err != nil || portInt <= 0 || portInt > 65535 {
		sendJSONError(w, "Invalid port number", http.StatusBadRequest)
		return
	}

	// 尝试连接数据库
	dbConn, err := db.Connect(db.ConnectionDetails{
		User:           reqData.Username,
		Password:       reqData.Password,
		Host:           reqData.Host,
		Port:           portInt,
		DBName:         reqData.Service,
		ConnectionType: "SERVICE_NAME",
	})

	if err != nil {
		sendJSONError(w, fmt.Sprintf("Failed to connect to database: %v", err), http.StatusOK)
		return
	}
	defer dbConn.Close()

	// 验证数据库连接和权限
	allAccessGranted, privilegeResults, err := db.CheckDatabaseConnection(dbConn)
	if err != nil {
		// If the connection is successful but permission check fails, still return a partially successful result.
		sendJSONResponse(w, ValidateResponse{
			Success:        !allAccessGranted,
			Message:        "Failed to check database permissions",
			PrivilegeCheck: privilegeResults,
		}, http.StatusOK)
		return
	}

	// Check for sufficient privileges.
	if !allAccessGranted {
		sendJSONResponse(w, ValidateResponse{
			Success:        false,
			Message:        "Failed to check database permissions",
			PrivilegeCheck: privilegeResults,
		}, http.StatusOK)
		return
	}

	// 所有检查通过
	sendJSONResponse(w, ValidateResponse{
		Success:        true,
		Message:        "Database connection and permission validation successful",
		PrivilegeCheck: privilegeResults,
	}, http.StatusOK)
}

// sendJSONResponse sends a successful response in JSON format.
func sendJSONResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// sendJSONError sends an error response in JSON format.
func sendJSONError(w http.ResponseWriter, message string, statusCode int) {
	sendJSONResponse(w, ValidateResponse{
		Success: false,
		Message: message,
	}, statusCode)
}

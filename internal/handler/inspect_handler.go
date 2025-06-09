package handler

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/goodwaysIT/inspect4oracle/internal/db"
	"github.com/goodwaysIT/inspect4oracle/internal/logger"
)

// reportStore 用于存储生成的报告
var (
	reportStore      = make(map[string]ReportData)
	reportStoreMutex sync.RWMutex
)

// GetReportStatusHandler 处理获取报告状态的API请求，返回JSON
func GetReportStatusHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reportId := r.URL.Query().Get("id")
		log.Printf("DIAGNOSTIC: GetReportStatusHandler called for /api/report/status. Path: %s, Report ID: %s", r.URL.Path, reportId)
		if reportId == "" {
			logger.Error("API Error: Missing report ID for /api/report/status")
			http.Error(w, "Missing report ID", http.StatusBadRequest)
			return
		}

		reportStoreMutex.RLock()
		reportData, exists := reportStore[reportId]
		reportStoreMutex.RUnlock()

		if !exists {
			logger.Error(fmt.Sprintf("API Error: Report ID %s not found for /api/report/status", reportId))
			http.NotFound(w, r)
			return
		}

		// 返回报告数据作为JSON
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		responsePayload := map[string]interface{}{
			"success":        true,
			"reportId":       reportId,
			"dbInfo":         reportData.DBName,
			"dbConnection":   reportData.DBConnection,
			"generatedAt":    reportData.GeneratedAt,
			"modules":        reportData.Modules,
			"title":          reportData.Title,
			"reportSections": reportData.ReportSections,
		}
		if err := json.NewEncoder(w).Encode(responsePayload); err != nil {
			logger.Error(fmt.Sprintf("API Error: Failed to encode report status for ID %s: %v", reportId, err))
			http.Error(w, "Failed to encode report data", http.StatusInternalServerError)
		}
	}
}

// ViewReportHandler 处理报告查看请求 (渲染HTML)
func ViewReportHandler(content embed.FS) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reportId := r.URL.Query().Get("id")
		// Log that the handler was called and what reportId it received
		log.Printf("DIAGNOSTIC: ViewReportHandler called for /report.html. Path: %s, Report ID: %s", r.URL.Path, reportId)
		if reportId == "" {
			http.Error(w, "Missing report ID", http.StatusBadRequest)
			return
		}

		reportStoreMutex.RLock()
		reportData, exists := reportStore[reportId]
		reportStoreMutex.RUnlock()

		if !exists {
			http.NotFound(w, r)
			return
		}

		currentYear := time.Now().Format("2006")
		nonceBytes := make([]byte, 16)
		_, err := rand.Read(nonceBytes)
		if err != nil {
			log.Printf("警告: 生成 CSP Nonce 失败: %v", err)
		}
		cspNonce := base64.RawURLEncoding.EncodeToString(nonceBytes)
		if err != nil {
			cspNonce = ""
		}

		templateData := map[string]interface{}{
			"DbInfo":         reportData.BusinessName, // 使用业务名称
			"ActualDBName":   reportData.DBName,       // 如果需要在模板其他地方显示实际数据库名，可以添加这个
			"DbConnection":   reportData.DBConnection,
			"GeneratedAt":    reportData.GeneratedAt,
			"Modules":        reportData.Modules,
			"Title":          reportData.Title,
			"CopyrightYear":  currentYear,
			"ReportSections": reportData.ReportSections,
			"CSPNonce":       cspNonce,
		}

		tmpl, err := template.New("layout").Funcs(template.FuncMap{
			"safeJS": func(s interface{}) template.JS {
				return template.JS(fmt.Sprint(s))
			},
		}).ParseFS(content, "templates/layout.html", "templates/report.html")
		if err != nil {
			http.Error(w, "无法加载报告模板: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := tmpl.ExecuteTemplate(w, "layout", templateData); err != nil {
			http.Error(w, "无法执行模板: "+err.Error(), http.StatusInternalServerError)
			log.Printf("CRITICAL: 模板执行错误 (ViewReportHandler): %v.", err)
			return
		}
	}
}

// DBConnectionRequest 定义数据库连接请求结构体
type DBConnectionRequest struct {
	Business string   `json:"business"`
	Host     string   `json:"host"`
	Port     string   `json:"port"`
	Service  string   `json:"service"`
	Username string   `json:"username"`
	Password string   `json:"password"`
	Items    []string `json:"items"`
	Lang     string   `json:"lang"`
}

// parseInspectRequest 解析巡检请求中的参数
// 它支持 JSON, x-www-form-urlencoded, 和 multipart/form-data 类型的请求
// 返回 DBConnectionRequest 结构体和可能发生的错误
func parseInspectRequest(r *http.Request) (*DBConnectionRequest, error) {
	var req DBConnectionRequest

	contentType := r.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "application/json") {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.Error(fmt.Sprintf("解析 JSON 请求体失败: %v", err))
			return nil, fmt.Errorf("无效的请求体: %w", err)
		}
	} else if strings.HasPrefix(contentType, "application/x-www-form-urlencoded") || strings.HasPrefix(contentType, "multipart/form-data") {
		if strings.HasPrefix(contentType, "multipart/form-data") {
			if err := r.ParseMultipartForm(32 << 20); err != nil { // 32MB 最大内存
				logger.Error(fmt.Sprintf("解析 multipart 表单数据失败: %v", err))
				return nil, fmt.Errorf("解析 multipart 表单数据失败: %w", err)
			}
		} else {
			if err := r.ParseForm(); err != nil {
				logger.Error(fmt.Sprintf("解析表单数据失败: %v", err))
				return nil, fmt.Errorf("解析表单数据失败: %w", err)
			}
		}
		req.Host = r.FormValue("host")
		req.Port = r.FormValue("port")
		req.Service = r.FormValue("service")
		req.Username = r.FormValue("username")
		req.Password = r.FormValue("password")
		req.Lang = r.FormValue("lang")
		req.Business = r.FormValue("business")

		// 处理 items 参数，它可能以两种形式出现：
		// 1. items=item1,item2,item3 (单个字符串，逗号分隔)
		// 2. items=item1&items=item2&items=item3 (多个同名参数)
		if r.Form["items"] != nil {
			// 检查是否是单个逗号分隔的字符串
			if len(r.Form["items"]) == 1 && strings.Contains(r.Form["items"][0], ",") {
				req.Items = strings.Split(r.Form["items"][0], ",")
			} else {
				// 否则，视为多个同名参数
				req.Items = r.Form["items"]
			}
		} else {
			// 兼容旧的 items[] 写法，虽然标准做法是 items
			formItems := r.Form["items[]"]
			if len(formItems) > 0 {
				req.Items = formItems
			} else {
				// 如果 items 和 items[] 都没有，则尝试从单个 item 字段获取（如果存在）
				// 这主要为了兼容早期可能的错误表单提交方式
				singleItem := r.FormValue("item")
				if singleItem != "" {
					req.Items = []string{singleItem}
				}
			}
		}

	} else {
		logger.Error(fmt.Sprintf("不支持的 Content-Type: %s", contentType))
		return nil, fmt.Errorf("不支持的 Content-Type: %s", contentType)
	}
	return &req, nil
}

// validateInspectParameters 校验巡检请求参数
func validateInspectParameters(req *DBConnectionRequest) error {
	if req.Host == "" || req.Port == "" || req.Service == "" || req.Username == "" {
		return fmt.Errorf("主机、端口、服务名和用户名不能为空")
	}
	if len(req.Items) == 0 {
		return fmt.Errorf("巡检项不能为空")
	}
	// 可以在这里添加更多校验逻辑，例如端口号格式等
	return nil
}

// handleRequestValidation 解析并校验巡检请求
func handleRequestValidation(r *http.Request) (*DBConnectionRequest, error) {
	contentType := r.Header.Get("Content-Type")
	logger.Infof("InspectHandler received request with Content-Type: %s", contentType)

	req, err := parseInspectRequest(r)
	if err != nil {
		return nil, fmt.Errorf("failed to parse inspect request: %w", err)
	}

	logger.Infof("Parsed inspection request: Business='%s', Host='%s', Port='%s', Service='%s', Username='%s', ItemsCount=%d, Lang='%s'",
		req.Business, req.Host, req.Port, req.Service, req.Username, len(req.Items), req.Lang)

	if err := validateInspectParameters(req); err != nil {
		return nil, fmt.Errorf("invalid parameters for request (Business='%s', Host='%s', Port='%s', Service='%s', Username='%s', ItemsCount=%d, Lang='%s'): %w",
			req.Business, req.Host, req.Port, req.Service, req.Username, len(req.Items), req.Lang, err)
	}
	return req, nil
}

// establishDBConnection 建立数据库连接并获取基础信息
func establishDBConnection(req *DBConnectionRequest) (*sql.DB, *db.FullDBInfo, error) {
	portInt, convErr := strconv.Atoi(req.Port)
	if convErr != nil {
		return nil, nil, fmt.Errorf("invalid port number '%s': %w", req.Port, convErr)
	}

	dbConn, err := db.Connect(db.ConnectionDetails{
		User:           req.Username,
		Password:       req.Password, // 注意：这里仍然使用了 req.Password，实际应用中应考虑安全性
		Host:           req.Host,
		Port:           portInt,
		DBName:         req.Service,
		ConnectionType: "SERVICE_NAME",
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	fullDBInfo, err := db.GetDatabaseInfo(dbConn)
	if err != nil {
		dbConn.Close() // Ensure connection is closed if GetDatabaseInfo fails
		return nil, nil, fmt.Errorf("failed to get database info: %w", err)
	}
	return dbConn, fullDBInfo, nil
}

// processInspectionModules 处理所有选定的巡检模块
func processInspectionModules(items []string, dbConn *sql.DB, lang string, fullDBInfo *db.FullDBInfo) []ReportModule {
	var modules []ReportModule
	for _, item := range items {
		module, err := ProcessInspectionItem(item, dbConn, lang, fullDBInfo)
		if err != nil {
			logger.Error(fmt.Sprintf("处理巡检项 %s 时出错: %v", item, err))
			module = ReportModule{
				ID:   item,
				Name: item,
				Cards: []ReportCard{{
					Title: "Error",
					Value: fmt.Sprintf("处理巡检项时出错: %v", err),
				}},
			}
		}
		modules = append(modules, module)
	}
	return modules
}

// prepareReportData 准备报告的整体数据结构
func prepareReportData(req *DBConnectionRequest, fullDBInfo *db.FullDBInfo, modules []ReportModule, lang string) (ReportData, string) {
	reportSections := make([]ReportSection, 0, len(modules))
	for _, module := range modules {
		reportSections = append(reportSections, ReportSection{
			ID:   module.ID,
			Name: module.Name,
		})
	}

	dbInfoStr := fmt.Sprintf("%s (v%s)", fullDBInfo.Database.Name.String, fullDBInfo.Database.OverallVersion)
	if len(fullDBInfo.Instances) > 0 {
		dbInfoStr += fmt.Sprintf(" @ %s", fullDBInfo.Instances[0].HostName)
	}

	dbConnectionStr := fmt.Sprintf("%s:%s/%s", req.Host, req.Port, req.Service)
	reportID := generateReportID(req.Host, req.Port, req.Service)

	reportData := ReportData{
		Lang:           lang,
		Title:          "Oracle Database Inspection Report",
		BusinessName:   req.Business,
		DBName:         fullDBInfo.Database.Name.String,
		DBFullInfo:     dbInfoStr,
		DBConnection:   dbConnectionStr,
		GeneratedAt:    time.Now().Format("2006-01-02 15:04:05"),
		Modules:        modules,
		ReportSections: reportSections,
	}
	return reportData, reportID
}

// storeAndRespond 保存报告并发送HTTP响应
func storeAndRespond(w http.ResponseWriter, reportID string, reportData ReportData) {
	reportStoreMutex.Lock()
	reportStore[reportID] = reportData
	reportStoreMutex.Unlock()

	response := map[string]interface{}{
		"success":  true,
		"reportId": reportID,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error(fmt.Sprintf("Error encoding response: %v", err))
		// If encoding fails, it's hard to send a meaningful HTTP error
	}
}

// InspectHandler 处理巡检请求
func InspectHandler(debug bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		req, err := handleRequestValidation(r)
		if err != nil {
			logger.Error(fmt.Sprintf("API Error: %v", err))
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		dbConn, fullDBInfo, err := establishDBConnection(req)
		if err != nil {
			logger.Error(fmt.Sprintf("API Error: %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError) // Or appropriate status based on error type
			return
		}
		defer func() {
			if dbConn != nil {
				if err := dbConn.Close(); err != nil {
					logger.Error(fmt.Sprintf("Error closing database connection: %v", err))
				}
			}
		}()

		modules := processInspectionModules(req.Items, dbConn, req.Lang, fullDBInfo)
		reportData, reportID := prepareReportData(req, fullDBInfo, modules, req.Lang)
		storeAndRespond(w, reportID, reportData)
	}
}

// generateReportID 根据输入参数生成一个唯一的报告ID
func generateReportID(host, port, service string) string {
	timestamp := time.Now().UnixNano()
	data := fmt.Sprintf("%s-%s-%s-%d", host, port, service, timestamp)
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash[:16]) // 使用哈希的前16字节作为ID，转换为十六进制字符串
}

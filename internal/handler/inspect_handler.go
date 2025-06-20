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

// reportStore is used to store generated reports
var (
	reportStore      = make(map[string]ReportData)
	reportStoreMutex sync.RWMutex
)

// GetReportStatusHandler handles API requests to get report status, returns JSON.
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

// ViewReportHandler handles report view requests (renders HTML).
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
			log.Printf("Warning: Failed to generate CSP Nonce: %v", err)
		}
		cspNonce := base64.RawURLEncoding.EncodeToString(nonceBytes)
		if err != nil {
			cspNonce = ""
		}

		templateData := map[string]interface{}{
			"DbInfo":         reportData.BusinessName, // Use business name
			"ActualDBName":   reportData.DBName,       // Add this if you need to display the actual database name elsewhere in the template
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
			log.Printf("CRITICAL: Template execution error (ViewReportHandler): %v.", err)
			return
		}
	}
}

// DBConnectionRequest defines the database connection request structure
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

// parseInspectRequest parses parameters from the inspection request.
// It supports JSON, x-www-form-urlencoded, and multipart/form-data request types.
// Returns a DBConnectionRequest struct and any potential error.
func parseInspectRequest(r *http.Request) (*DBConnectionRequest, error) {
	var req DBConnectionRequest

	contentType := r.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "application/json") {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.Error(langText("解析JSON请求体失败: %v", "Failed to parse JSON request body: %v", "JSONリクエストボディの解析に失敗しました: %v", req.Lang), err)
			return nil, fmt.Errorf(langText("解析JSON请求体失败: %w", "Failed to parse JSON request body: %w", "JSONリクエストボディの解析に失敗しました: %w", req.Lang), err)
		}
	} else if strings.HasPrefix(contentType, "application/x-www-form-urlencoded") || strings.HasPrefix(contentType, "multipart/form-data") {
		if strings.HasPrefix(contentType, "multipart/form-data") {
			if err := r.ParseMultipartForm(32 << 20); err != nil { // 32MB 最大内存
				logger.Error(langText("解析multipart/form-data失败: %v", "Failed to parse multipart form data: %v", "multipart/form-dataの解析に失敗しました: %v", req.Lang), err)
				return nil, fmt.Errorf(langText("解析multipart/form-data失败: %w", "Failed to parse multipart form data: %w", "multipart/form-dataの解析に失敗しました: %w", req.Lang), err)
			}
		} else {
			if err := r.ParseForm(); err != nil {
				logger.Error(langText("解析表单数据失败: %v", "Failed to parse form data: %v", "フォームデータの解析に失敗しました: %v", req.Lang), err)
				return nil, fmt.Errorf(langText("解析表单数据失败: %w", "Failed to parse form data: %w", "フォームデータの解析に失敗しました: %w", req.Lang), err)
			}
		}
		req.Host = r.FormValue("host")
		req.Port = r.FormValue("port")
		req.Service = r.FormValue("service")
		req.Username = r.FormValue("username")
		req.Password = r.FormValue("password")
		req.Lang = r.FormValue("lang")
		req.Business = r.FormValue("business")

		// Handle the 'items' parameter, which can appear in two forms:
		// 1. items=item1,item2,item3 (single comma-separated string)
		// 2. items=item1&items=item2&items=item3 (multiple parameters with the same name)
		if r.Form["items"] != nil {
			// Check if it's a single comma-separated string
			if len(r.Form["items"]) == 1 && strings.Contains(r.Form["items"][0], ",") {
				req.Items = strings.Split(r.Form["items"][0], ",")
			} else {
				// Otherwise, treat as multiple parameters with the same name
				req.Items = r.Form["items"]
			}
		} else {
			// Compatible with the old items[] notation, although the standard is 'items'
			formItems := r.Form["items[]"]
			if len(formItems) > 0 {
				req.Items = formItems
			} else {
				// If neither 'items' nor 'items[]' are present, try to get from a single 'item' field (if it exists)
				// This is mainly for compatibility with potentially incorrect form submissions from earlier versions
				singleItem := r.FormValue("item")
				if singleItem != "" {
					req.Items = []string{singleItem}
				}
			}
		}

	} else {
		logger.Error(langText("不支持的内容类型: %s", "Unsupported Content-Type: %s", "サポートされていないコンテンツタイプ: %s", req.Lang), contentType)
		return nil, fmt.Errorf(langText("不支持的内容类型: %s", "Unsupported Content-Type: %s", "サポートされていないコンテンツタイプ: %s", req.Lang), contentType)
	}
	return &req, nil
}

// validateInspectParameters validates the inspection request parameters
func validateInspectParameters(req *DBConnectionRequest) error {
	if req.Host == "" || req.Port == "" || req.Service == "" || req.Username == "" {
		return fmt.Errorf(langText("主机、端口、服务名和用户名不能为空", "Host, port, service name and username cannot be empty", "ホスト、ポート、サービス名、ユーザー名は空にできません", req.Lang))
	}
	if len(req.Items) == 0 {
		return fmt.Errorf(langText("巡检项不能为空", "Inspection items cannot be empty", "検査項目は空にできません", req.Lang))
	}
	// More validation logic can be added here, e.g., port number format.
	return nil
}

// handleRequestValidation parses and validates the inspection request
func handleRequestValidation(r *http.Request) (*DBConnectionRequest, error) {
	contentType := r.Header.Get("Content-Type")
	logger.Infof("InspectHandler received request with Content-Type: %s", contentType)

	req, err := parseInspectRequest(r)
	if err != nil {
		return nil, fmt.Errorf(langText("解析巡检请求失败: %w", "failed to parse inspect request: %w", "検査リクエストの解析に失敗しました: %w", req.Lang), err)
	}

	logger.Infof("Parsed inspection request: Business='%s', Host='%s', Port='%s', Service='%s', Username='%s', ItemsCount=%d, Lang='%s'",
			req.Business, req.Host, req.Port, req.Service, req.Username, len(req.Items), req.Lang)

	if err := validateInspectParameters(req); err != nil {
		return nil, fmt.Errorf("invalid parameters for request (Business='%s', Host='%s', Port='%s', Service='%s', Username='%s', ItemsCount=%d, Lang='%s'): %w",
			req.Business, req.Host, req.Port, req.Service, req.Username, len(req.Items), req.Lang, err)
	}
	return req, nil
}

// establishDBConnection establishes a database connection and retrieves basic information
func establishDBConnection(req *DBConnectionRequest) (*sql.DB, *db.FullDBInfo, error) {
	portInt, convErr := strconv.Atoi(req.Port)
	if convErr != nil {
		return nil, nil, fmt.Errorf(langText("无效的端口号 '%s': %w", "invalid port number '%s': %w", "無効なポート番号 '%s': %w", req.Lang), req.Port, convErr)
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
		return nil, nil, fmt.Errorf(langText("连接数据库失败: %w", "failed to connect to database: %w", "データベースへの接続に失敗しました: %w", req.Lang), err)
	}

	fullDBInfo, err := db.GetDatabaseInfo(dbConn)
	if err != nil {
		dbConn.Close() // Ensure connection is closed if GetDatabaseInfo fails
		return nil, nil, fmt.Errorf(langText("获取数据库信息失败: %w", "failed to get database info: %w", "データベース情報の取得に失敗しました: %w", req.Lang), err)
	}
	return dbConn, fullDBInfo, nil
}

// processInspectionModules processes all selected inspection modules.
func processInspectionModules(items []string, dbConn *sql.DB, lang string, fullDBInfo *db.FullDBInfo) []ReportModule {
	var modules []ReportModule
	for _, item := range items {
		module, err := ProcessInspectionItem(item, dbConn, lang, fullDBInfo)
		if err != nil {
			logger.Error(langText("处理巡检项 %s 时出错: %v", "Error processing inspection item %s: %v", "検査項目 %s の処理中にエラーが発生しました: %v", lang), item, err)
			module = ReportModule{
				ID:   item,
				Name: item,
				Cards: []ReportCard{{
					Title: langText("错误", "Error", "エラー", lang),
					Value: fmt.Sprintf(langText("处理巡检项时出错: %v", "Error processing inspection item: %v", "検査項目の処理中にエラーが発生しました: %v", lang), err),
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
		Title:          langText("Oracle 数据库巡检报告", "Oracle Database Inspection Report", "Oracleデータベース検査レポート", lang),
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
		logger.Error(fmt.Sprintf("Failed to store report and send response: %v", err))
		// If encoding fails, it's hard to send a meaningful HTTP error
	}
}

// InspectHandler handles inspection requests.
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
					logger.Error(fmt.Sprintf("Failed to close database connection: %v", err))
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

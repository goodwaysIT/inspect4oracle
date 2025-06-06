package handler

import (
	"crypto/rand"
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

// InspectHandler 处理巡检请求
func InspectHandler(debug bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// 记录请求头
		headers, _ := json.Marshal(r.Header)
		logger.Info(fmt.Sprintf("请求头: %s", string(headers)))

		var (
			host, port, service, username, password, lang string
			businessNameFromRequest                       string // 用于存储从请求中获取的业务名称
			items                                         []string
		)

		// 检查 Content-Type 头
		contentType := r.Header.Get("Content-Type")
		if contentType == "application/json" {
			// 解析 JSON 请求体
			var req DBConnectionRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				logger.Error(fmt.Sprintf("解析 JSON 请求体失败: %v", err))
				http.Error(w, "Invalid request body", http.StatusBadRequest)
				return
			}
			// 使用 JSON 数据
			host = req.Host
			port = req.Port
			service = req.Service
			username = req.Username
			password = req.Password
			items = req.Items
			lang = req.Lang
			businessNameFromRequest = req.Business
		} else {
			// 解析表单数据
			if err := r.ParseForm(); err != nil {
				logger.Error(fmt.Sprintf("解析表单数据失败: %v", err))
				http.Error(w, "Failed to parse form data", http.StatusBadRequest)
				return
			}

			// 如果是 multipart 表单，解析 multipart 数据
			if strings.Contains(r.Header.Get("Content-Type"), "multipart/form-data") {
				if err := r.ParseMultipartForm(32 << 20); err != nil { // 32MB 最大内存
					logger.Error(fmt.Sprintf("解析 multipart 表单数据失败: %v", err))
					http.Error(w, "Failed to parse multipart form data", http.StatusBadRequest)
					return
				}
			}

			// 获取表单数据
			host = r.FormValue("host")
			port = r.FormValue("port")
			service = r.FormValue("service")
			username = r.FormValue("username")
			password = r.FormValue("password")
			lang = r.FormValue("lang")
			businessNameFromRequest = r.FormValue("business")

			// 获取 items
			// 首先尝试从 MultipartForm 获取
			if r.MultipartForm != nil && r.MultipartForm.Value != nil {
				if values, ok := r.MultipartForm.Value["items[]"]; ok && len(values) > 0 {
					items = values
				} else if values, ok := r.MultipartForm.Value["items"]; ok && len(values) > 0 {
					items = values
				}
			}

			// 如果 MultipartForm 中没有找到，尝试从 Form 获取
			if len(items) == 0 {
				if values, ok := r.Form["items[]"]; ok && len(values) > 0 {
					items = values
				} else if values, ok := r.Form["items"]; ok && len(values) > 0 {
					items = values
				}
			}

			// 记录接收到的 items 值
			logger.Info(fmt.Sprintf("接收到的 items: %v", items))

			// 记录接收到的数据
			logger.Info(fmt.Sprintf("接收到的表单数据: host=%s, port=%s, service=%s, username=%s, password=%v, items=%v",
				host, port, service, username, password != "", items))
		}

		// 验证必填字段
		if host == "" || port == "" || service == "" || username == "" || password == "" || len(items) == 0 {
			errMsg := fmt.Sprintf("Missing required fields: host=%s, port=%s, service=%s, username=%s, password=%v, items=%v",
				host, port, service, username, password != "", items)
			logger.Error(errMsg)
			http.Error(w, errMsg, http.StatusBadRequest)
			return
		}

		// 构建连接详情
		portInt, err := strconv.Atoi(port)
		if err != nil {
			http.Error(w, "Invalid port number", http.StatusBadRequest)
			return
		}

		// 连接数据库
		dbConn, err := db.Connect(db.ConnectionDetails{
			User:           username,
			Password:       password,
			Host:           host,
			Port:           portInt,
			DBName:         service,
			ConnectionType: "SERVICE_NAME",
		})
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to connect to database: %v", err), http.StatusInternalServerError)
			return
		}
		defer func() {
			if err := dbConn.Close(); err != nil {
				logger.Error(fmt.Sprintf("Error closing database connection: %v", err))
			}
		}()

		// 获取数据库信息
		fullDBInfo, err := db.GetDatabaseInfo(dbConn)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to get database info: %v", err), http.StatusInternalServerError)
			return
		}

		// 处理选中的巡检项
		var modules []ReportModule
		for _, item := range items {
			module, err := ProcessInspectionItem(item, dbConn, lang, fullDBInfo)
			if err != nil {
				logger.Error(fmt.Sprintf("处理巡检项 %s 时出错: %v", item, err))
				// 添加错误信息到报告中，而不是直接跳过
				module = ReportModule{
					ID:   item,
					Name: item,
					Cards: []ReportCard{{ // Changed to ReportCard type
						Title: "Error",
						Value: fmt.Sprintf("处理巡检项时出错: %v", err),
					}},
				}
			}
			modules = append(modules, module)
		}

		// 生成报告ID
		reportId := fmt.Sprintf("%s_%d", time.Now().Format("20060102150405"), time.Now().UnixNano()%1000)

		// 准备报告数据
		reportSections := make([]ReportSection, 0, len(modules))
		for _, module := range modules {
			reportSections = append(reportSections, ReportSection{
				ID:   module.ID,
				Name: module.Name,
			})
		}

		// 构建更完整的数据库信息字符串
		dbInfoStr := fmt.Sprintf("%s (v%s)", fullDBInfo.Database.Name.String, fullDBInfo.Database.OverallVersion)
		if len(fullDBInfo.Instances) > 0 {
			dbInfoStr += fmt.Sprintf(" @ %s", fullDBInfo.Instances[0].HostName)
		}

		// 构建数据库连接信息字符串
		dbConnectionStr := fmt.Sprintf("%s:%s/%s", host, port, service)

		reportData := ReportData{
			Lang:           lang,
			Title:          "Oracle Database Inspection Report", // 考虑使用 businessNameFromRequest 作为标题的一部分
			BusinessName:   businessNameFromRequest,
			DBName:         fullDBInfo.Database.Name.String, // 实际数据库名
			DBFullInfo:     dbInfoStr,             // 数据库完整信息
			DBConnection:   dbConnectionStr,
			GeneratedAt:    time.Now().Format("2006-01-02 15:04:05"),
			Modules:        modules,
			ReportSections: reportSections,
		}

		// 保存报告数据到内存或数据库
		reportStoreMutex.Lock()
		reportStore[reportId] = reportData
		reportStoreMutex.Unlock()

		// 返回报告ID
		response := map[string]interface{}{
			"success":  true,
			"reportId": reportId,
		}

		// 设置响应头
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

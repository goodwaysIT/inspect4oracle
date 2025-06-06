package handler

import (
	"embed"
	"html/template"
	"io/fs"
	"net/http"
	"time"
)

// IndexHandler 处理主页请求
func IndexHandler(content embed.FS) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 创建一个新的模板集
		tmpl := template.New("")

		// 从嵌入的文件系统中读取模板文件
		templateData, err := fs.ReadFile(content, "templates/index.html")
		if err != nil {
			http.Error(w, "无法读取模板文件: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// 解析模板
		tmpl, err = tmpl.Parse(string(templateData))
		if err != nil {
			http.Error(w, "模板解析错误: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// 执行模板
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		err = tmpl.Execute(w, nil)
		if err != nil {
			http.Error(w, "模板执行错误: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// ReportHandler 处理报告页面请求
func ReportHandler(content embed.FS) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 获取报告ID
		reportId := r.URL.Query().Get("id")
		if reportId == "" {
			http.Error(w, "Missing report ID", http.StatusBadRequest)
			return
		}

		// 从 reportStore 中获取报告数据
		reportStoreMutex.RLock()
		reportData, exists := reportStore[reportId]
		reportStoreMutex.RUnlock()

		if !exists {
			http.NotFound(w, r)
			return
		}

		// 创建一个新的模板集
		tmpl := template.New("")

		// 从嵌入的文件系统中读取模板文件
		templateData, err := fs.ReadFile(content, "templates/report.html")
		if err != nil {
			http.Error(w, "无法读取报告模板文件: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// 解析模板
		tmpl, err = tmpl.Parse(string(templateData))
		if err != nil {
			http.Error(w, "报告模板解析错误: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// 准备模板数据
		templateDataMap := map[string]interface{}{
			"DbInfo":      reportData.DBName,
			"GeneratedAt": time.Now().Format("2006-01-02 15:04:05"),
			"Modules":     reportData.Modules,
		}

		// 执行模板
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		err = tmpl.Execute(w, templateDataMap)
		if err != nil {
			http.Error(w, "报告模板执行错误: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

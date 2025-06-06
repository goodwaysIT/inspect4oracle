package main

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os" // Required for flag.Usage (os.Stderr, os.Args) and os.Exit
	"net/http"

	"github.com/goodwaysIT/inspect4oracle/internal/handler"
	"github.com/goodwaysIT/inspect4oracle/internal/logger"

	"github.com/gorilla/mux"
)

//go:embed static templates
var content embed.FS

const AppVersion = "0.1.0" // Application version constant

func main() {
	host := flag.String("host", "0.0.0.0", "IP address")
	port := flag.String("port", "8080", "Port")
	debug := flag.Bool("debug", false, "Debug mode")
	showVersion := flag.Bool("version", false, "Print version information and exit")

	// Custom usage message for -h/--help
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Inspect4Oracle - Oracle Database Inspection Tool\n\n")
		fmt.Fprintf(os.Stderr, "Version: %s\n\n", AppVersion)
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExample:\n")
		fmt.Fprintf(os.Stderr, "  %s -host 127.0.0.1 -port 9090 -debug\n", os.Args[0])
	}

	flag.Parse()
	if *showVersion {
		fmt.Printf("Inspect4Oracle version %s\n", AppVersion)
		os.Exit(0)
	}
	addr := fmt.Sprintf("%s:%s", *host, *port)

	// 初始化日志系统，防止 logger.Error 等为 nil 导致 panic
	logger.Init(*debug)

	r := mux.NewRouter()

	// 静态文件服务
	staticFS, err := fs.Sub(content, "static")
	if err != nil {
		log.Fatal(err)
	}
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

	// 注册根路由
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		handler.IndexHandler(content)(w, r)
	}).Methods("GET")

	// 报告页面路由
	r.HandleFunc("/report.html", handler.ViewReportHandler(content)).Methods("GET")

	// 创建子路由用于 API
	apiRouter := r.PathPrefix("/api").Subrouter()

	// API 日志中间件
	apiRouter.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Info(fmt.Sprintf("API 请求: %s %s", r.Method, r.URL.Path))
			// 设置 CORS 头
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	})

	// 注册 API 路由
	apiRouter.HandleFunc("/validate", handler.ValidateConnection).Methods("POST")
	apiRouter.HandleFunc("/inspect", func(w http.ResponseWriter, r *http.Request) {
		logger.Info("处理 /api/inspect 请求")
		handler.InspectHandler(*debug)(w, r)
	}).Methods("POST")
	// 移除冲突的 /api/report 路由，因为它功能模糊且与 /report.html 重叠
	// apiRouter.HandleFunc("/report", handler.ViewReportHandler(content)).Methods("GET")
	apiRouter.HandleFunc("/report/status", handler.GetReportStatusHandler()).Methods("GET") // 使用新的 GetReportStatusHandler 返回JSON

	// 主路由的日志中间件
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Info(fmt.Sprintf("页面请求: %s %s", r.Method, r.URL.Path))
			next.ServeHTTP(w, r)
		})
	})

	logger.Infof("--- Inspect4Oracle ---")
	logger.Infof("Version: %s", AppVersion)
	if *debug {
		logger.Info("Debug mode: enabled")
	} else {
		logger.Info("Debug mode: disabled")
	}
	logger.Infof("Server starting and listening on http://%s", addr)
	logger.Fatalf("Server failed to start: %v", http.ListenAndServe(addr, r))
}

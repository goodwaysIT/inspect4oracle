package main

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os" // Required for flag.Usage (os.Stderr, os.Args) and os.Exit

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

	// Static file serving
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

	// Create a subrouter for the API
	apiRouter := r.PathPrefix("/api").Subrouter()

	// API 日志中间件
	apiRouter.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Info(fmt.Sprintf("API Request: %s %s", r.Method, r.URL.Path))
			// Set CORS headers
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
		logger.Info("Handling /api/inspect request")
		handler.InspectHandler(*debug)(w, r)
	}).Methods("POST")
	// Remove conflicting /api/report route as its functionality is ambiguous and overlaps with /report.html
	// apiRouter.HandleFunc("/report", handler.ViewReportHandler(content)).Methods("GET")
	apiRouter.HandleFunc("/report/status", handler.GetReportStatusHandler()).Methods("GET") // Use the new GetReportStatusHandler to return JSON

	// Logging middleware for the main router
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Info(fmt.Sprintf("Page Request: %s %s", r.Method, r.URL.Path))
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

package handler

// ReportData 结构用于模板渲染
type ReportData struct {
	DBFullInfo     string          `json:"dbFullInfo,omitempty"` // Full database information, e.g., "ORCL (v19.3.0.0.0) @ dbhost.example.com"
	Lang           string          // 语言字段: "zh" 或 "en"
	Title          string          // 报告主标题
	BusinessName   string          // Business system name entered by the user
	DBName         string          // Name of the currently inspected database, used for download filenames and report titles
	DBConnection   string          // Database connection string, format: ip:port/servicename
	GeneratedAt    string          // 报告生成时间
	Modules        []ReportModule  // Data for each module included in the report
	ReportSections []ReportSection // List of modules for the left navigation menu
}

// ReportSection 结构用于定义报告的左侧导航菜单项
type ReportSection struct {
	ID   string // 对应 ReportModule 的 ID
	Name string // 对应 ReportModule 的 Name
}

// ParsedDSN struct stores information parsed from an Oracle connection string
type ParsedDSN struct {
	User        string
	Password    string
	Host        string
	Port        string
	SID         string
	ServiceName string
	IsValid     bool // Mark if DSN successfully parsed host and port
}

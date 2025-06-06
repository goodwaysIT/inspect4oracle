package handler

// ReportData 结构用于模板渲染
type ReportData struct {
	Lang           string          // 语言字段: "zh" 或 "en"
	Title          string          // 报告主标题
	BusinessName   string          // 用户输入的业务系统名称
	DBName         string          // 当前巡检的数据库名，用于下载文件名和报告标题部分
	DBConnection   string          // 数据库连接信息，格式为 ip:port/servicename
	GeneratedAt    string          // 报告生成时间
	Modules        []ReportModule  // 报告包含的各个模块数据
	ReportSections []ReportSection // 用于左侧导航菜单的模块列表
}

// ReportSection 结构用于定义报告的左侧导航菜单项
type ReportSection struct {
	ID   string // 对应 ReportModule 的 ID
	Name string // 对应 ReportModule 的 Name
}

// ParsedDSN 结构用于存储从Oracle连接字符串中解析出来的信息
type ParsedDSN struct {
	User        string
	Password    string
	Host        string
	Port        string
	SID         string
	ServiceName string
	IsValid     bool // 标记DSN是否成功解析出主机和端口
}

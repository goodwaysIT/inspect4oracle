package handler

import (
	"database/sql"
	"fmt"

	"github.com/goodwaysIT/inspect4oracle/internal/db"
	"github.com/goodwaysIT/inspect4oracle/internal/logger"
)

// moduleProcessFunc defines the standard signature for all module processing functions.
// They take a database connection, language, and pre-fetched full database info (which can be nil).
// They return slices of report cards, tables, charts, and an error.
type moduleProcessFunc func(dbConn *sql.DB, lang string, fullDBInfo *db.FullDBInfo) ([]ReportCard, []*ReportTable, []ReportChart, error)

// Adapter for processParametersModule
func adaptParametersModule(dbConn *sql.DB, lang string, _ *db.FullDBInfo) ([]ReportCard, []*ReportTable, []ReportChart, error) {
	// Original processParametersModule doesn't expect fullDBInfo, so we ignore it here.
	// It also returns a concrete []ReportChart which is usually nil for this module.
	return processParametersModule(dbConn, lang)
}

// Adapter for processDbinfoModule - its signature is already compatible
// func adaptDbinfoModule(dbConn *sql.DB, lang string, fullDBInfo *db.FullDBInfo) ([]ReportCard, []*ReportTable, []ReportChart, error) {
// 	 return processDbinfoModule(dbConn, lang, fullDBInfo)
// }

// Adapter for processStorageModule (assuming original doesn't take fullDBInfo)
func adaptStorageModule(dbConn *sql.DB, lang string, _ *db.FullDBInfo) ([]ReportCard, []*ReportTable, []ReportChart, error) {
	return processStorageModule(dbConn, lang)
}

// Adapter for processSessionsModule (assuming original doesn't take fullDBInfo)
func adaptSessionsModule(dbConn *sql.DB, lang string, _ *db.FullDBInfo) ([]ReportCard, []*ReportTable, []ReportChart, error) {
	return processSessionsModule(dbConn, lang)
}

// Adapter for processObjectsModule
func adaptObjectsModule(dbConn *sql.DB, lang string, _ *db.FullDBInfo) ([]ReportCard, []*ReportTable, []ReportChart, error) {
	return processObjectsModule(dbConn, lang)
}

// Adapter for processPerformanceModule (assuming original doesn't take fullDBInfo)
func adaptPerformanceModule(dbConn *sql.DB, lang string, _ *db.FullDBInfo) ([]ReportCard, []*ReportTable, []ReportChart, error) {
	return processPerformanceModule(dbConn, lang)
}

// Adapter for processSecurityModule - its signature is already compatible
// func adaptSecurityModule(dbConn *sql.DB, lang string, fullDBInfo *db.FullDBInfo) ([]ReportCard, []*ReportTable, []ReportChart, error) {
// 	 return processSecurityModule(dbConn, lang, fullDBInfo)
// }

// Adapter for processBackupModule (assuming original doesn't take fullDBInfo)
func adaptBackupModule(dbConn *sql.DB, lang string, _ *db.FullDBInfo) ([]ReportCard, []*ReportTable, []ReportChart, error) {
	return processBackupModule(dbConn, lang)
}

// moduleInfo holds information about a module, including its name and processing function.
// We use a struct to potentially extend this with more module-specific metadata later (e.g., icons, titles).
type moduleInfo struct {
	nameFunc  func(lang string) string
	processor moduleProcessFunc
	titleFunc func(lang string) string // Optional: for modules with specific titles
	icon      string                   // Optional: for module icon
}

// moduleProcessors maps inspection item keys to their respective moduleInfo.
var moduleProcessors = map[string]moduleInfo{
	"params": {
		nameFunc:  func(lang string) string { return langText("数据库参数", "Key Database Parameters", lang) },
		processor: adaptParametersModule,
	},
	"parameters": { // Alias for params
		nameFunc:  func(lang string) string { return langText("数据库参数", "Key Database Parameters", lang) },
		processor: adaptParametersModule,
	},
	"dbinfo": {
		nameFunc:  func(lang string) string { return langText("基本信息", "Basic Info", lang) },
		processor: processDbinfoModule, // Assumes processDbinfoModule is compatible or adapted
	},
	"storage": {
		nameFunc:  func(lang string) string { return langText("存储信息", "Storage Info", lang) },
		processor: adaptStorageModule,
	},
	"sessions": {
		nameFunc:  func(lang string) string { return langText("会话详情", "Session Details", lang) },
		processor: adaptSessionsModule,
	},
	"objects": {
		nameFunc: func(lang string) string { return langText("数据库对象", "Database Objects", lang) },
		titleFunc: func(lang string) string {
			return langText("数据库对象统计与状态", "Database Objects Statistics & Status", lang)
		},
		icon:      "fas fa-cube",
		processor: adaptObjectsModule,
	},
	"performance": {
		nameFunc:  func(lang string) string { return langText("数据库性能", "Database Performance", lang) },
		processor: adaptPerformanceModule,
	},
	"security": {
		nameFunc:  func(lang string) string { return langText("安全配置", "Security Configuration", lang) },
		processor: processSecurityModule, // Assumes processSecurityModule is compatible or adapted
	},
	"backup": {
		nameFunc:  func(lang string) string { return langText("备份与恢复", "Backup & Recovery", lang) },
		processor: adaptBackupModule,
	},
}

// ProcessInspectionItem 处理单个巡检项并返回报告模块。
// fullDBInfo 参数包含了从 dbinfo 模块预先获取的数据库的全面信息，供其他模块参考。
// 如果 fullDBInfo 为 nil (例如，在获取 dbinfo 本身时发生错误)，函数仍会尝试处理，但依赖 fullDBInfo 的模块可能会受影响。
func ProcessInspectionItem(item string, dbConn *sql.DB, lang string, fullDBInfo *db.FullDBInfo) (ReportModule, error) {
	module := ReportModule{ID: item, Cards: []ReportCard{}} // Initialize module

	pInfo, ok := moduleProcessors[item]
	if !ok {
		module.Name = fmt.Sprintf(langText("未知模块: %s", "Unknown Module: %s", lang), item)
		errMsg := fmt.Sprintf(langText("此模块 '%s' 的处理器未实现", "Handler for module '%s' is not implemented", lang), item)
		module.Cards = []ReportCard{{Title: langText("错误", "Error", lang), Value: errMsg}}
		return module, fmt.Errorf(errMsg)
	}

	module.Name = pInfo.nameFunc(lang)
	if pInfo.titleFunc != nil {
		module.Title = pInfo.titleFunc(lang)
	}
	if pInfo.icon != "" {
		module.Icon = pInfo.icon
	}

	// Log before calling the processor, especially for modules like backup that might take time
	if item == "backup" { // Specific logging for backup or other long-running modules
		logger.Infof("开始委派处理 %s 模块...", item)
	}

	cards, tables, charts, err := pInfo.processor(dbConn, lang, fullDBInfo)

	module.Cards = append(module.Cards, cards...)
	module.Tables = append(module.Tables, tables...)
	module.Charts = append(module.Charts, charts...)

	if err != nil {
		logger.Errorf("%s 模块处理返回错误: %v", item, err) // Generic error logging
		module.Error = err.Error()                  // Store error message in module
		// Note: The individual processor or its adapter is responsible for adding specific error cards if needed.
		// For critical errors that should halt further processing for this module, the processor should return the error.
		// If the error is not nil, the caller (InspectHandler) might decide how to proceed globally.
		// For some modules (like 'sessions', 'objects', 'performance', 'security' in the original switch),
		// the error from the processor didn't cause an immediate return of (module, err) from ProcessInspectionItem.
		// The new structure consistently sets module.Error. If a specific module's error should also be returned
		// from ProcessInspectionItem (as was the case for 'params', 'dbinfo', 'storage'), the processor should ensure this.
		// For now, we will return the error if it's not nil, mimicking the stricter cases.
		return module, err
	}

	return module, nil
}

// langText 是一个辅助函数，用于根据语言选择文本。
// 实际项目中，这个函数可能位于一个共享的 utils 或 i18n 包中。
func langText(zhText, enText, lang string) string {
	if lang == "zh" {
		return zhText
	}
	return enText // 默认为英文
}

// formatNullInt64AsGB 辅助函数，用于格式化 sql.NullInt64 并转换为 GB
func formatNullInt64AsGB(ni sql.NullInt64) string {
	if ni.Valid {
		return fmt.Sprintf("%.2f GB", float64(ni.Int64)/1024/1024/1024)
	}
	return "N/A"
}

// cardFromError 是一个辅助函数，用于从错误创建标准错误卡片
func cardFromError(titleKey, titleDefault string, err error, lang string) ReportCard {
	return ReportCard{
		Title: langText(titleKey, titleDefault, lang),
		Value: fmt.Sprintf(langText("获取信息失败: %v", "Failed to get information: %v", lang), err),
	}
}

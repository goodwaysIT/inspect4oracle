package handler

import (
	"database/sql"
	"fmt"

	"github.com/goodwaysIT/inspect4oracle/internal/db"
	"github.com/goodwaysIT/inspect4oracle/internal/logger"
)

// ProcessInspectionItem 处理单个巡检项并返回报告模块。
// fullDBInfo 参数包含了从 dbinfo 模块预先获取的数据库的全面信息，供其他模块参考。
// 如果 fullDBInfo 为 nil (例如，在获取 dbinfo 本身时发生错误)，函数仍会尝试处理，但依赖 fullDBInfo 的模块可能会受影响。
func ProcessInspectionItem(item string, dbConn *sql.DB, lang string, fullDBInfo *db.FullDBInfo) (ReportModule, error) {
	module := ReportModule{ID: item, Cards: []ReportCard{}} // 初始化模块，ID为巡检项名称，Cards明确类型

	// 默认情况下，如果 fullDBInfo 为 nil，后续依赖它的模块可能会出错或显示不完整信息
	// 各个 case 中需要妥善处理 fullDBInfo 可能为 nil 的情况

	switch item {
	case "params", "parameters":
		module.Name = langText("数据库参数", "Key Database Parameters", lang)
		cards, tables, _, err := processParametersModule(dbConn, lang)
		module.Cards = append(module.Cards, cards...)
		module.Tables = append(module.Tables, tables...)
		if err != nil {
			module.Error = err.Error() // Store error message in module
			// The error card is already added by the helper function.
			return module, err // Return error for logging/handling by caller
		}

	case "dbinfo":
		module.Name = langText("基本信息", "Basic Info", lang)
		cards, tables, _, err := processDbinfoModule(dbConn, lang, fullDBInfo)
		module.Cards = append(module.Cards, cards...)
		module.Tables = append(module.Tables, tables...)
		if err != nil {
			module.Error = err.Error()
			return module, err
		}

	case "storage":
		module.Name = langText("存储信息", "Storage Info", lang)
		cards, tables, charts, err := processStorageModule(dbConn, lang)
		module.Cards = append(module.Cards, cards...)
		module.Tables = append(module.Tables, tables...)
		module.Charts = append(module.Charts, charts...)
		if err != nil {
			module.Error = err.Error()
			return module, err
		}
	case "sessions":
		module.Name = langText("会话详情", "Session Details", lang)
		cards, tables, charts, processErr := processSessionsModule(dbConn, lang)
		module.Cards = append(module.Cards, cards...)
		module.Tables = append(module.Tables, tables...)
		module.Charts = append(module.Charts, charts...)
		if processErr != nil {
			module.Error = processErr.Error()
			// Error card might already be added by processSessionsModule,
			// but setting module.Error ensures it's reported.
			// The caller of ProcessInspectionItem might still want to log processErr.
		}
	case "objects":
		module.Name = langText("数据库对象", "Database Objects", lang)
		module.Title = langText("数据库对象统计与状态", "Database Objects Statistics & Status", lang)
		module.Icon = "fas fa-cube" // Example icon
		cards, tables, _, processErr := processObjectsModule(dbConn, lang)
		module.Cards = append(module.Cards, cards...)
		module.Tables = append(module.Tables, tables...)
		if processErr != nil {
			module.Error = processErr.Error()
		}
	case "performance":
		module.Name = langText("数据库性能", "Database Performance", lang)
		cards, tables, charts, processErr := processPerformanceModule(dbConn, lang)
		module.Cards = append(module.Cards, cards...)
		module.Tables = append(module.Tables, tables...)
		module.Charts = append(module.Charts, charts...)
		if processErr != nil {
			module.Error = processErr.Error()
			// 如果 processPerformanceModule 返回错误，通常它内部已经记录并创建了相应的错误卡片
			// 此处仅设置顶层模块错误状态
		}
	case "security":
		module.Name = langText("安全配置", "Security Configuration", lang)
		cards, tables, _, processErr := processSecurityModule(dbConn, lang, fullDBInfo) // charts is expected to be nil
		module.Cards = append(module.Cards, cards...)
		module.Tables = append(module.Tables, tables...)
		// module.Charts will remain empty or nil as security module doesn't produce charts
		if processErr != nil {
			module.Error = processErr.Error()
		}
	case "backup":
		module.Name = langText("备份与恢复", "Backup & Recovery", lang)
		logger.Infof("开始委派处理备份模块...")
		cards, tables, charts, err := processBackupModule(dbConn, lang)
		module.Cards = append(module.Cards, cards...)
		module.Tables = append(module.Tables, tables...)
		module.Charts = append(module.Charts, charts...)
		if err != nil {
			logger.Errorf("备份模块处理返回错误: %v", err)
			module.Error = err.Error()
		}

	default:
		module.Name = fmt.Sprintf(langText("未知模块: %s", "Unknown Module: %s", lang), item)
		errMsg := fmt.Sprintf(langText("此模块 '%s' 的处理器未实现", "Handler for module '%s' is not implemented", lang), item)
		module.Cards = []ReportCard{{Title: langText("错误", "Error", lang), Value: errMsg}}
		return module, fmt.Errorf(errMsg) // 返回错误，以便上层记录
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

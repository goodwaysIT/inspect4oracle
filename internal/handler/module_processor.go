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
		nameFunc:  func(lang string) string { return langText("数据库参数", "Key Database Parameters", "主要なデータベースパラメータ", lang) },
		processor: adaptParametersModule,
	},
	"parameters": { // Alias for params
		nameFunc:  func(lang string) string { return langText("Key Database Parameters", "Key Database Parameters", "Key Database Parameters", lang) },
		processor: adaptParametersModule,
	},
	"dbinfo": {
		nameFunc:  func(lang string) string { return langText("Basic Info", "Basic Info", "Basic Info", lang) },
		processor: processDbinfoModule, // Assumes processDbinfoModule is compatible or adapted
	},
	"storage": {
		nameFunc:  func(lang string) string { return langText("Storage Info", "Storage Info", "Storage Info", lang) },
		processor: adaptStorageModule,
	},
	"sessions": {
		nameFunc:  func(lang string) string { return langText("Session Details", "Session Details", "Session Details", lang) },
		processor: adaptSessionsModule,
	},
	"objects": {
		nameFunc: func(lang string) string { return langText("Database Objects", "Database Objects", "Database Objects", lang) },
		titleFunc: func(lang string) string {
			return langText("Database Objects Statistics & Status", "Database Objects Statistics & Status", "Database Objects Statistics & Status", lang)
		},
		icon:      "fas fa-cube",
		processor: adaptObjectsModule,
	},
	"performance": {
		nameFunc:  func(lang string) string { return langText("Database Performance", "Database Performance", "Database Performance", lang) },
		processor: adaptPerformanceModule,
	},
	"security": {
		nameFunc:  func(lang string) string { return langText("Security Configuration", "Security Configuration", "Security Configuration", lang) },
		processor: processSecurityModule, // Assumes processSecurityModule is compatible or adapted
	},
	"backup": {
		nameFunc:  func(lang string) string { return langText("Backup & Recovery", "Backup & Recovery", "Backup & Recovery", lang) },
		processor: adaptBackupModule,
	},
}

// ProcessInspectionItem processes a single inspection item and returns a report module.
// The fullDBInfo parameter contains comprehensive database information pre-fetched from the dbinfo module for reference by other modules.
// If fullDBInfo is nil (e.g., an error occurred while fetching dbinfo itself), the function will still attempt to process, but modules dependent on fullDBInfo may be affected.
func ProcessInspectionItem(item string, dbConn *sql.DB, lang string, fullDBInfo *db.FullDBInfo) (ReportModule, error) {
	module := ReportModule{ID: item, Cards: []ReportCard{}} // Initialize module

	pInfo, ok := moduleProcessors[item]
	if !ok {
		module.Name = fmt.Sprintf(langText("Unknown Module: %s", "Unknown Module: %s", "Unknown Module: %s", lang), item)
		errMsg := fmt.Sprintf(langText("Handler for module '%s' is not implemented", "Handler for module '%s' is not implemented", "Handler for module '%s' is not implemented", lang), item)
		module.Cards = []ReportCard{{Title: langText("Error", "Error", "Error", lang), Value: errMsg}}
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
		logger.Infof("Starting to delegate processing for module %s...", item)
	}

	cards, tables, charts, err := pInfo.processor(dbConn, lang, fullDBInfo)

	module.Cards = append(module.Cards, cards...)
	module.Tables = append(module.Tables, tables...)
	module.Charts = append(module.Charts, charts...)

	if err != nil {
		logger.Errorf("Error processing module %s: %v", item, err) // Generic error logging
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

// langText is a helper function for selecting text based on language.
// In a real project, this function might be located in a shared utils or i18n package.
func langText(zhText, enText, jpText, lang string) string {
	switch lang {
	case "zh":
		return zhText
	case "jp":
		return jpText
	default:
		return enText // Default to English
	}
}

// formatNullInt64AsGB is a helper function to format sql.NullInt64 and convert it to GB.
func formatNullInt64AsGB(ni sql.NullInt64) string {
	if ni.Valid {
		return fmt.Sprintf("%.2f GB", float64(ni.Int64)/1024/1024/1024)
	}
	return "N/A"
}

// cardFromError is a helper function to create a standard error card from an error.
func cardFromError(titleZh, titleEn, titleJp string, err error, lang string) ReportCard {
	return ReportCard{
		Title: langText(titleZh, titleEn, titleJp, lang),
		Value: fmt.Sprintf(langText("Failed to get data: %v", "Failed to get data: %v", "Failed to get data: %v", lang), err),
	}
}

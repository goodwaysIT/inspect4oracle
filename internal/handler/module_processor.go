package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"strconv"
	"strings"

	"github.com/goodwaysIT/inspect4oracle/internal/db"
	"github.com/goodwaysIT/inspect4oracle/internal/logger"
)

// chartColor 定义了图表的边框和背景颜色
type chartColor struct {
	BorderColor     string
	BackgroundColor string
}

// performanceChartColors 为性能图表提供一组预定义的颜色
var performanceChartColors = []chartColor{
	{BorderColor: "#007bff", BackgroundColor: "rgba(0, 123, 255, 0.2)"},   // Blue
	{BorderColor: "#28a745", BackgroundColor: "rgba(40, 167, 69, 0.2)"},   // Green
	{BorderColor: "#dc3545", BackgroundColor: "rgba(220, 53, 69, 0.2)"},   // Red
	{BorderColor: "#ffc107", BackgroundColor: "rgba(255, 193, 7, 0.2)"},   // Yellow
	{BorderColor: "#17a2b8", BackgroundColor: "rgba(23, 162, 184, 0.2)"},  // Cyan
	{BorderColor: "#6f42c1", BackgroundColor: "rgba(111, 66, 193, 0.2)"},  // Indigo
	{BorderColor: "#fd7e14", BackgroundColor: "rgba(253, 126, 20, 0.2)"},  // Orange
	{BorderColor: "#20c997", BackgroundColor: "rgba(32, 201, 151, 0.2)"},  // Teal
	{BorderColor: "#6c757d", BackgroundColor: "rgba(108, 117, 125, 0.2)"}, // Grey
	{BorderColor: "#e83e8c", BackgroundColor: "rgba(232, 62, 140, 0.2)"},  // Pink
}

// processParametersModule handles the "parameters" inspection item.
func processParametersModule(dbConn *sql.DB, lang string) (cards []ReportCard, tables []*ReportTable, charts []ReportChart, err error) {
	params, dbErr := db.GetParameterList(dbConn)
	if dbErr != nil {
		cards = append(cards, ReportCard{Title: langText("错误", "Error", lang), Value: fmt.Sprintf(langText("获取参数失败: %v", "Failed to get parameters: %v", lang), dbErr)})
		return cards, nil, nil, dbErr
	}
	logger.Debugf("获取到参数: %v", params)
	if len(params) > 0 {
		paramTable := &ReportTable{
			Name:    langText("参数列表", "Parameter List", lang),
			Headers: []string{langText("参数名", "Parameter Name", lang), langText("值", "Value", lang)},
			Rows:    [][]string{},
		}
		for _, p := range params {
			row := []string{p.Name, p.Value.String}
			paramTable.Rows = append(paramTable.Rows, row)
		}
		tables = append(tables, paramTable)
	}
	return cards, tables, nil, nil
}

// processDbinfoModule handles the "dbinfo" inspection item.
func processDbinfoModule(dbConn *sql.DB, lang string, preFetchedInfo *db.FullDBInfo) (cards []ReportCard, tables []*ReportTable, charts []ReportChart, err error) {
	dbInfoToProcess := preFetchedInfo
	var fetchErr error

	if dbInfoToProcess == nil {
		dbInfoToProcess, fetchErr = db.GetDatabaseInfo(dbConn)
		if fetchErr != nil {
			cards = append(cards, ReportCard{Title: langText("错误", "Error", lang), Value: fmt.Sprintf(langText("获取数据库信息失败: %v", "Failed to get database info: %v", lang), fetchErr)})
			return cards, nil, nil, fetchErr
		}
	}

	dbCards := []ReportCard{
		{Title: langText("数据库名称", "DB Name", lang), Value: dbInfoToProcess.Database.Name.String},
		{Title: langText("DBID", "DBID", lang), Value: formatNullInt64(dbInfoToProcess.Database.DBID)},
		{Title: langText("创建时间", "Created", lang), Value: dbInfoToProcess.Database.Created.String},
		{Title: langText("版本", "Overall Version", lang), Value: dbInfoToProcess.Database.OverallVersion},
		{Title: langText("日志模式", "Log Mode", lang), Value: dbInfoToProcess.Database.LogMode},
		{Title: langText("打开模式", "Open Mode", lang), Value: dbInfoToProcess.Database.OpenMode},
		{Title: langText("容器数据库", "CDB", lang), Value: dbInfoToProcess.Database.CDB.String},
		{Title: langText("保护模式", "Protection Mode", lang), Value: dbInfoToProcess.Database.ProtectionMode},
		{Title: langText("闪回", "Flashback", lang), Value: dbInfoToProcess.Database.FlashbackOn},
		{Title: langText("强制日志", "Force Logging", lang), Value: dbInfoToProcess.Database.ForceLogging.String},
		{Title: langText("角色", "DB Role", lang), Value: dbInfoToProcess.Database.DatabaseRole},
		{Title: langText("平台名称", "Platform Name", lang), Value: dbInfoToProcess.Database.PlatformName},
		{Title: langText("数据库唯一名称", "DB Unique Name", lang), Value: dbInfoToProcess.Database.DBUniqueName.String},
		{Title: langText("字符集", "Character Set", lang), Value: dbInfoToProcess.Database.CharacterSet.String},
		{Title: langText("全国字符集", "National Character Set", lang), Value: dbInfoToProcess.Database.NationalCharacterSet.String},
	}
	cards = append(cards, dbCards...)

	if len(dbInfoToProcess.Instances) > 0 {
		instanceTable := &ReportTable{
			Name:    langText("实例详情", "Instance Details", lang),
			Headers: []string{langText("实例号", "Inst ID", lang), langText("实例名", "Instance Name", lang), langText("主机名", "Host Name", lang), langText("版本", "Version", lang), langText("启动时间", "Startup Time", lang), langText("状态", "Status", lang)},
			Rows:    [][]string{},
		}
		for _, inst := range dbInfoToProcess.Instances {
			row := []string{
				strconv.Itoa(inst.InstanceNumber),
				inst.InstanceName,
				inst.HostName,
				inst.Version,
				inst.StartupTime,
				inst.Status,
			}
			instanceTable.Rows = append(instanceTable.Rows, row)
		}
		tables = append(tables, instanceTable)
	}
	return cards, tables, nil, nil
}

// processStorageModule handles the "storage" inspection item.
func processStorageModule(dbConn *sql.DB, lang string) (cards []ReportCard, tables []*ReportTable, charts []ReportChart, err error) {
	storageData, dbErr := db.GetStorageInfo(dbConn)
	if dbErr != nil {
		cards = append(cards, ReportCard{Title: langText("错误", "Error", lang), Value: fmt.Sprintf(langText("获取存储信息失败: %v", "Failed to get storage info: %v", lang), dbErr)})
		return cards, nil, nil, dbErr
	}

	if len(storageData.ControlFiles) > 0 {
		cfTable := &ReportTable{
			Name:    langText("控制文件", "Control Files", lang),
			Headers: []string{langText("文件路径", "File Path", lang), langText("大小(MB)", "Size(MB)", lang)},
			Rows:    [][]string{},
		}
		for _, cf := range storageData.ControlFiles {
			row := []string{cf.Name, fmt.Sprintf("%.2f", cf.SizeMB)}
			cfTable.Rows = append(cfTable.Rows, row)
		}
		tables = append(tables, cfTable)
	} else {
		cards = append(cards, ReportCard{Title: langText("控制文件", "Control Files", lang), Value: langText("未找到", "Not Found", lang)})
	}

	if len(storageData.RedoLogs) > 0 {
		redoTable := &ReportTable{
			Name:    langText("Redo 日志组", "Redo Log Groups", lang),
			Headers: []string{langText("组号", "Group#", lang), langText("线程号", "Thread#", lang), langText("成员数", "Members", lang), langText("大小(MB)", "Size(MB)", lang), langText("文件", "Members", lang), langText("状态", "Status", lang), langText("归档", "Archived", lang), langText("类型", "Type", lang)},
			Rows:    [][]string{},
		}
		for _, rl := range storageData.RedoLogs {
			row := []string{
				strconv.Itoa(rl.GroupNo),
				strconv.Itoa(rl.ThreadNo),
				strconv.Itoa(rl.Members),
				fmt.Sprintf("%.2f", rl.SizeMB),
				rl.Member,
				rl.Status,
				rl.Archived,
				rl.Type,
			}
			redoTable.Rows = append(redoTable.Rows, row)
		}
		tables = append(tables, redoTable)
	}

	if len(storageData.Tablespaces) > 0 {
		tsTable := &ReportTable{
			Name:    langText("表空间使用情况", "Tablespace Usage", lang),
			Headers: []string{langText("状态", "Status", lang), langText("表空间名称", "Tablespace", lang), langText("类型", "Type", lang), langText("extent管理", "Extent Management", lang), langText("segment管理", "Segment Management", lang), langText("已用(MB)", "Used(MB)", lang), langText("总量(MB)", "Total(MB)", lang), langText("使用率(%)", "Used %", lang), langText("可扩展大小(MB)", "Autoextend Size(MB)", lang)},
			Rows:    [][]string{},
		}
		for _, ts := range storageData.Tablespaces {
			row := []string{
				ts.Status,
				ts.Name,
				ts.Type,
				ts.ExtentManagement,
				ts.SegmentSpaceManagement,
				fmt.Sprintf("%.2f", ts.UsedMB),
				fmt.Sprintf("%.2f", ts.TotalMB),
				ts.UsedPercent,
				fmt.Sprintf("%.2f", ts.CanExtendMB),
			}
			tsTable.Rows = append(tsTable.Rows, row)
		}
		tables = append(tables, tsTable)
	}
	// Note: The original code for ASM Disk Groups and Datafile IO stats was not fully visible in the snippet.
	// If those sections exist, they should also be part of this helper or their own helpers.
	return cards, tables, charts, nil
}

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
		}				tsTable.Rows = append(tsTable.Rows, row)
			}
			module.Tables = append(module.Tables, tsTable)
		}

		// 数据文件信息
		if len(storageData.DataFiles) > 0 {
			dfTable := &ReportTable{
				Name:    langText("数据文件", "Data Files", lang),
				Headers: []string{langText("文件编号", "File ID", lang), langText("文件名", "File Name", lang), langText("表空间", "Tablespace", lang), langText("大小(MB)", "Size(MB)", lang), langText("状态", "Status", lang), langText("自动扩展", "Autoextend", lang)},
				Rows:    [][]string{},
			}
			for _, df := range storageData.DataFiles {
				row := []string{
					fmt.Sprintf("%d", df.FileID),
					df.FileName,
					df.TablespaceName,
					fmt.Sprintf("%.2f", df.SizeMB),
					df.Status,
					df.Autoextensible,
				}
				dfTable.Rows = append(dfTable.Rows, row)
			}
			module.Tables = append(module.Tables, dfTable)
		}

		// 归档日志摘要表格 (如果存在)
		if len(storageData.ArchivedLogsSummary) > 0 {
			archTable := &ReportTable{
				Name:    langText("归档日志摘要 (近7日)", "Archived Log Summary (Last 7 Days)", lang),
				Headers: []string{langText("日期", "Day", lang), langText("日志数量", "Log Count", lang), langText("总大小(MB)", "Total Size(MB)", lang)},
				Rows:    [][]string{},
			}
			for _, al := range storageData.ArchivedLogsSummary {
				row := []string{
					al.Day,
					strconv.Itoa(al.LogCount),
					fmt.Sprintf("%.2f", al.TotalSizeMB),
				}
				archTable.Rows = append(archTable.Rows, row)
			}
			module.Tables = append(module.Tables, archTable)
		}

		// ASM 磁盘组表格 (如果存在)
		if len(storageData.ASMDiskgroups) > 0 {
			asmTable := &ReportTable{
				Name:    langText("ASM 磁盘组", "ASM Diskgroups", lang),
				Headers: []string{langText("磁盘组名称", "Diskgroup", lang), langText("总大小(MB)", "Total(MB)", lang), langText("空闲(MB)", "Free(MB)", lang), langText("使用率(%)", "Used %", lang), langText("状态", "State", lang), langText("冗余类型", "Redundancy", lang)},
				Rows:    [][]string{},
			}
			for _, adg := range storageData.ASMDiskgroups {
				row := []string{
					adg.Name,
					strconv.FormatInt(adg.TotalMB, 10),
					strconv.FormatInt(adg.FreeMB, 10),
					fmt.Sprintf("%.2f", adg.UsedPercent),
					adg.State,
					adg.RedundancyType,
				}
				asmTable.Rows = append(asmTable.Rows, row)
			}
			module.Tables = append(module.Tables, asmTable)
		}
	case "sessions":
		// 初始化会话模块
		module.Name = langText("会话详情", "Session Details", lang)
		logger.Debugf("开始处理会话模块（完全恢复），语言: %s", lang)

		// 从数据库获取数据
		var sessionData *db.AllSessionInfo
		var overviewErr, eventErr, historyErr error

		sessionData, overviewErr, eventErr, historyErr = db.GetSessionDetails(dbConn)

		// 1. 会话总览 - 表格 (已恢复)
		if overviewErr != nil {
			logger.Warnf("处理会话模块 - 会话总览数据获取失败: %v", overviewErr)
			module.Cards = append(module.Cards, ReportCard{Title: langText("会话总览", "Session Overview", lang), Value: fmt.Sprintf(langText("获取数据失败: %v", "Failed to get data: %v", lang), overviewErr)})
		} else if sessionData != nil && len(sessionData.Overview) > 0 {
			sessionOverviewTable := ReportTable{
				Name:    langText("会话总览", "Session Overview", lang),
				Headers: []string{langText("实例", "Inst", lang), langText("用户名", "Username", lang), langText("机器", "Machine", lang), langText("状态", "Status", lang), langText("会话数", "Count", lang)},
				Rows:    [][]string{},
			}
			for _, so := range sessionData.Overview {
				var usernameStr, machineStr string
				if so.Username.Valid {
					usernameStr = so.Username.String
				} else {
					usernameStr = ""
				}
				if so.Machine.Valid {
					machineStr = so.Machine.String
				} else {
					machineStr = ""
				}
				row := []string{strconv.Itoa(so.InstID), usernameStr, machineStr, so.Status, strconv.Itoa(so.SessionCount)}
				sessionOverviewTable.Rows = append(sessionOverviewTable.Rows, row)
			}
			module.Tables = append(module.Tables, &sessionOverviewTable)
		} else {
			module.Cards = append(module.Cards, ReportCard{Title: langText("会话总览", "Session Overview", lang), Value: langText("无会话总览数据", "No session overview data available.", lang)})
		}

		// 2. 按等待事件统计的会话数 - 表格 (已恢复并修正)
		if eventErr != nil {
			logger.Warnf("处理会话模块 - 按等待事件统计数据获取失败: %v", eventErr)
			module.Cards = append(module.Cards, ReportCard{Title: langText("按等待事件统计会话数", "Session Count by Wait Event", lang), Value: fmt.Sprintf(langText("获取数据失败: %v", "Failed to get data: %v", lang), eventErr)})
		} else if sessionData != nil && len(sessionData.ByEvent) > 0 {
			sessionByEventTable := ReportTable{
				Name:    langText("按等待事件统计会话数", "Session Count by Wait Event", lang),
				Headers: []string{langText("等待事件", "Wait Event", lang), langText("会话数", "Count", lang)},
				Rows:    [][]string{},
			}
			for _, sbe := range sessionData.ByEvent {
				eventNameStr := sbe.Event
				if eventNameStr == "" {
					eventNameStr = langText("未知事件", "Unknown Event", lang)
				}
				row := []string{eventNameStr, strconv.Itoa(sbe.SessionCount)}
				sessionByEventTable.Rows = append(sessionByEventTable.Rows, row)
			}
			module.Tables = append(module.Tables, &sessionByEventTable)
		} else {
			module.Cards = append(module.Cards, ReportCard{Title: langText("按等待事件统计会话数", "Session Count by Wait Event", lang), Value: langText("无按等待事件统计的会话数据", "No session count by wait event data available.", lang)})
		}

		// 3. 最近一天会话数 - Chart.js Line Chart
		if historyErr != nil {
			logger.Warnf("处理会话模块 - 获取最近活动会话历史数据失败: %v (这可能是由于ASH未启用或许可问题，图表将不会显示)", historyErr)
			// 如果获取历史数据失败，可以添加一个卡片提示，而不是完全不显示任何东西
			module.Cards = append(module.Cards, ReportCard{
				Title: langText("最近活动会话历史", "Recent Active Session History", lang),
				Value: fmt.Sprintf(langText("获取图表数据失败: %v。图表无法生成。", "Failed to get chart data: %v. Chart cannot be generated.", lang), historyErr),
			})
			// module.ChartDataJSON = "{}" // 确保图表数据为空JSON对象
		} else if sessionData != nil && len(sessionData.HistoryForChart) > 0 {
			labels := make([]string, len(sessionData.HistoryForChart))
			dataCounts := make([]int, len(sessionData.HistoryForChart))
			for i, h := range sessionData.HistoryForChart {
				labels[i] = h.SampleTime
				dataCounts[i] = h.SessionCount
			}

			// 1. Prepare ChartDataset
			var rptDatasets []ChartDataset
			// Convert labels and dataCounts to []ChartDataPoint
			var points []ChartDataPoint
			for i, label := range labels { // Assuming labels and dataCounts have same length
				points = append(points, ChartDataPoint{X: label, Y: dataCounts[i]})
			}

			rptDatasets = append(rptDatasets, ChartDataset{
				Label:       langText("数据库会话数", "Database Session Count", lang),
				Data:        points,
				BorderColor: "rgb(75, 192, 192)",
				Fill:        false,
				// tension: 0.1, // tension is a Chart.js specific option, not directly in our ChartDataset struct
			})

			// 2. Prepare ChartJSOptions
			rptOptions := ChartJSOptions{
				Responsive:          true,
				MaintainAspectRatio: false, // Set to false to allow custom sizing via CSS or canvas attributes
				Plugins: ChartPluginsOptions{
					Title: ChartPluginTitleOptions{
						Display: true,
						Text:    langText("最近24小时活动会话数趋势", "Active Session Trend (Last 24h)", lang),
					},
					Legend: ChartPluginLegendOptions{
						Display:  true,
						Position: "top", // Example position
					},
				},
				Scales: ChartScalesOptions{
					X: ChartScaleOptions{
						Type: "time", // 确保 Type 仍然是 "time"
						Time: &ChartTimeScaleOptions{
							TooltipFormat: "yyyy-MM-dd HH:mm", // 鼠标悬停提示框中的时间格式 (24小时制)
							DisplayFormats: &ChartTimeDisplayFormats{ // X轴刻度上显示的时间格式
								Minute: "HH:mm",       // 当X轴单位为分钟时，显示为 "14:30" (24小时制)
								Hour:   "MM-dd HH:mm", // 当X轴单位为小时时，显示为 "05-30 14:00" (24小时制)
								Day:    "yyyy-MM-dd",  // 当X轴单位为天时
								Week:   "yyyy-MM-dd",
								Month:  "yyyy-MM",
								Year:   "yyyy",
							},
						},
						Title: ChartScaleTitleOptions{
							Display: true,
							// 您可以保持或修改这里的标题文本，例如 langText("时间", "Time", lang)
							Text: langText("时间 (HH:MM)", "Time (HH:MM)", lang),
						},
					},
					Y: ChartScaleOptions{
						BeginAtZero: true,
						Title: ChartScaleTitleOptions{
							Display: true,
							Text:    langText("活动会话数", "Active Sessions", lang),
						},
					},
				},
			}

			// 3. Serialize chart data and options to JSON
			chartJSData := ChartJSData{Datasets: rptDatasets}
			datasetsJSON, err := json.Marshal(chartJSData)
			if err != nil {
				logger.Errorf("无法序列化会话图表数据集为JSON: %v", err)
				// Optionally, handle the error by not adding the chart or adding a chart with an error message
			}
			optionsJSON, err := json.Marshal(rptOptions)
			if err != nil {
				logger.Errorf("无法序列化会话图表选项为JSON: %v", err)
				// Optionally, handle the error
			}

			// 4. Create ReportChart with JSON strings
			sessionChart := ReportChart{
				ChartID:      "sessionHistoryChart",
				Type:         "line",
				DatasetsJSON: template.HTML(datasetsJSON),
				OptionsJSON:  template.HTML(optionsJSON),
			}
			module.Charts = append(module.Charts, sessionChart)
			logger.Debugf("已为会话模块添加图表: %s (ID: %s)", rptOptions.Plugins.Title.Text, sessionChart.ChartID)
		} else {
			// 如果没有历史数据，也添加一个卡片提示
			module.Cards = append(module.Cards, ReportCard{
				Title: langText("最近活动会话历史", "Recent Active Session History", lang),
				Value: langText("无最近活动会话历史数据可供显示。", "No recent active session history data available to display.", lang),
			})
			// module.ChartDataJSON = "{}" // 确保图表数据为空JSON对象
		}
	case "objects":
		module.Name = langText("对象概述", "Object Overview", lang)
		logger.Infof("开始处理对象模块... 语言: %s", lang)

		objectData, overviewErr, topSegmentsErr, invalidObjectsErr := db.GetObjectDetails(dbConn)

		if overviewErr != nil {
			logger.Errorf("处理对象模块 - 获取对象概述失败: %v", overviewErr)
			module.Cards = append(module.Cards, ReportCard{
				Title: langText("对象概述错误", "Object Overview Error", lang),
				Value: fmt.Sprintf(langText("获取对象概述数据失败: %v", "Failed to get object overview data: %v", lang), overviewErr),
			})
		}
		if topSegmentsErr != nil {
			logger.Errorf("处理对象模块 - 获取Top段信息失败: %v", topSegmentsErr)
			module.Cards = append(module.Cards, ReportCard{
				Title: langText("Top段信息错误", "Top Segments Error", lang),
				Value: fmt.Sprintf(langText("获取Top段数据失败: %v", "Failed to get top segments data: %v", lang), topSegmentsErr),
			})
		}
		if invalidObjectsErr != nil {
			logger.Errorf("处理对象模块 - 获取无效对象信息失败: %v", invalidObjectsErr)
			module.Cards = append(module.Cards, ReportCard{
				Title: langText("无效对象信息错误", "Invalid Objects Error", lang),
				Value: fmt.Sprintf(langText("获取无效对象数据失败: %v", "Failed to get invalid objects data: %v", lang), invalidObjectsErr),
			})
		}

		// 处理对象概述数据
		if overviewErr == nil {
			if objectData != nil && len(objectData.Overview) > 0 {
				overviewTable := &ReportTable{
					Name:    langText("对象概述", "Object Overview", lang),
					Headers: []string{langText("所有者", "Owner", lang), langText("对象类型", "Object Type", lang), langText("对象数量", "Object Count", lang)},
					Rows:    [][]string{},
				}
				for _, ov := range objectData.Overview {
					row := []string{ov.Owner, ov.ObjectType, strconv.Itoa(ov.ObjectCount)}
					overviewTable.Rows = append(overviewTable.Rows, row)
				}
				module.Tables = append(module.Tables, overviewTable)
			} else {
				module.Cards = append(module.Cards, ReportCard{
					Title: langText("对象概述", "Object Overview", lang),
					Value: langText("未查询到对象概述数据。", "No object overview data found.", lang),
				})
			}
		} // overviewErr is handled by the error card logic above

		// 处理Top段数据
		if topSegmentsErr == nil {
			if objectData != nil && len(objectData.TopSegments) > 0 {
				topSegmentsTable := &ReportTable{
					Name:    langText("Top段 (按大小)", "Top Segments (by Size)", lang),
					Headers: []string{langText("所有者", "Owner", lang), langText("段类型", "Segment Type", lang), langText("段名", "Segment Name", lang), langText("大小 (MB)", "Size (MB)", lang)},
					Rows:    [][]string{},
				}
				for _, ts := range objectData.TopSegments {
					row := []string{ts.Owner, ts.SegmentType, ts.SegmentName, fmt.Sprintf("%.2f", ts.SizeMB)}
					topSegmentsTable.Rows = append(topSegmentsTable.Rows, row)
				}
				module.Tables = append(module.Tables, topSegmentsTable)
			} else {
				module.Cards = append(module.Cards, ReportCard{
					Title: langText("Top段", "Top Segments", lang),
					Value: langText("未查询到Top段数据。", "No top segments data found.", lang),
				})
			}
		} // topSegmentsErr is handled by the error card logic above

		// 处理无效对象数据 (已存在逻辑)
		if objectData != nil && len(objectData.InvalidObjects) > 0 {
			invalidObjectsTable := &ReportTable{
				Name:    langText("无效对象列表", "Invalid Objects List", lang),
				Headers: []string{langText("所有者", "Owner", lang), langText("对象名", "Object Name", lang), langText("对象类型", "Object Type", lang), langText("创建时间", "Created", lang), langText("最后DDL时间", "Last DDL Time", lang)},
				Rows:    [][]string{},
			}
			for _, obj := range objectData.InvalidObjects {
				row := []string{obj.Owner, obj.ObjectName, obj.ObjectType, obj.Created, obj.LastDDLTime}
				invalidObjectsTable.Rows = append(invalidObjectsTable.Rows, row)
			}
			module.Tables = append(module.Tables, invalidObjectsTable)
		} else if invalidObjectsErr == nil {
			// No error, but no invalid objects found
			module.Cards = append(module.Cards, ReportCard{
				Title: langText("无效对象检查", "Invalid Objects Check", lang),
				Value: langText("未发现无效对象。", "No invalid objects found.", lang),
			})
		}

		//module.ChartDataJSON = "{}" // Objects module currently does not have a chart
	case "performance":
		module.Name = langText("性能指标", "Performance Metrics", lang)
		module.Title = langText("数据库性能指标 (最近24小时)", "Database Performance Metrics (Last 24 Hours)", lang)
		module.Icon = "fas fa-chart-line"

		sysMetrics, err := db.GetSysMetricSummary(dbConn)
		if err != nil {
			logger.Errorf("处理性能模块 - 获取 SysMetricSummary 失败: %v", err)
			module.Error = fmt.Sprintf(langText("获取性能指标数据失败: %v", "Failed to get performance metrics data: %v", lang), err)
			module.Cards = append(module.Cards, cardFromError("性能数据错误", "Performance Data Error", err, lang))
			return module, err
		}

		if len(sysMetrics) == 0 {
			module.Description = langText("在过去24小时内未找到相关的性能指标数据。", "No relevant performance metrics data found for the past 24 hours.", lang)
			module.Cards = append(module.Cards, ReportCard{Title: langText("数据提示", "Data Note", lang), Value: langText("未找到性能指标数据", "No performance metrics data found", lang)})
			return module, nil
		}

		// 按 MetricName 分组数据
		metricsByName := make(map[string][]db.SysMetricSummary)
		for _, sm := range sysMetrics {
			if sm.MetricName.Valid {
				metricsByName[sm.MetricName.String] = append(metricsByName[sm.MetricName.String], sm)
			}
		}

		module.Charts = []ReportChart{}
		chartIndex := 0 // 初始化图表颜色索引

		// 为每个 MetricName 创建图表
		// 定义一个期望的顺序，或者根据实际获取的指标名称来迭代
		desiredMetricsOrder := []string{
			"DB Time Per Sec",
			"Average Active Sessions",
			"CPU Usage Per Sec",
			"Host CPU Utilization (%)",
			"Executions Per Sec",
			"User Commits Per Sec",
			"Physical Read Total Bytes Per Sec",
			"Physical Write Total Bytes Per Sec",
			"Redo Generated Per Sec",
			"Network Traffic Volume Per Sec",
			"SQL Service Response Time",
			// 可以按需添加其他指标，例如：
			// "CPU Usage Per Txn",
			// "Executions Per Txn",
			// "User Rollbacks Per Sec",
			// "DB Block Changes Per Sec",
			// "GC CR Block Received Per Second",
			// "GC Current Block Received Per Second",
			// "Background Time Per Sec",
		}

		for _, metricName := range desiredMetricsOrder {
			metricData, ok := metricsByName[metricName]
			if !ok || len(metricData) == 0 {
				logger.Warnf("性能模块: 未找到指标 '%s' 的数据或数据为空", metricName)
				continue // 如果没有这个指标的数据，则跳过
			}

			dataPoints := []ChartDataPoint{}
			var metricUnit string
			// 对每个指标的数据点按时间排序，确保图表正确显示
			// sort.Slice(metricData, func(i, j int) bool {
			// 	 return metricData[i].EndTime.Before(metricData[j].EndTime)
			// })

			for _, data := range metricData {
				if data.ValueStr.Valid {
					dataPoints = append(dataPoints, ChartDataPoint{
						X: data.EndTimeStr.String, // 格式化为 YYYY-MM-DD HH:MM
						Y: data.ValueStr.String,
					})
				}
				if data.MetricUnit.Valid && metricUnit == "" { // 获取一次单位即可
					metricUnit = data.MetricUnit.String
				}
			}

			if len(dataPoints) == 0 {
				logger.Warnf("性能模块: 指标 '%s' 没有有效的数据点", metricName)
				continue
			}

			chartTitle := metricName
			yAxisLabel := metricUnit
			if yAxisLabel == "" {
				yAxisLabel = langText("值", "Value", lang)
			}

			// 生成一个对前端更友好的 chartId
			cleanMetricName := strings.ToLower(metricName)
			cleanMetricName = strings.ReplaceAll(cleanMetricName, " ", "_")
			cleanMetricName = strings.ReplaceAll(cleanMetricName, "(%)", "percent") // 处理百分号
			cleanMetricName = strings.ReplaceAll(cleanMetricName, "/", "_per_")     // 处理斜杠

			// Prepare datasets and options for this specific chart
			selectedColor := performanceChartColors[chartIndex%len(performanceChartColors)]
			currentChartDatasets := []ChartDataset{
				{
					Label:           metricName,
					Data:            dataPoints,
					BorderColor:     selectedColor.BorderColor,
					BackgroundColor: selectedColor.BackgroundColor,
					Fill:            true, // 根据需要可以设为 false
				},
			}

			perfChartOptions := ChartJSOptions{
				Responsive:          true,
				MaintainAspectRatio: false,
				Plugins: ChartPluginsOptions{
					Title: ChartPluginTitleOptions{
						Display: true,
						Text:    chartTitle,
					},
					Legend: ChartPluginLegendOptions{
						Display:  len(currentChartDatasets) > 1, // Only display legend if multiple datasets
						Position: "top",
					},
				},
				Scales: ChartScalesOptions{
					X: ChartScaleOptions{
						Type: "time", // 确保 Type 仍然是 "time"
						Time: &ChartTimeScaleOptions{
							TooltipFormat: "yyyy-MM-dd HH:mm", // 鼠标悬停提示框中的时间格式 (24小时制)
							DisplayFormats: &ChartTimeDisplayFormats{ // X轴刻度上显示的时间格式
								Minute: "HH:mm",       // 当X轴单位为分钟时，显示为 "14:30" (24小时制)
								Hour:   "MM-dd HH:mm", // 当X轴单位为小时时，显示为 "05-30 14:00" (24小时制)
								Day:    "yyyy-MM-dd",  // 当X轴单位为天时
								Week:   "yyyy-MM-dd",
								Month:  "yyyy-MM",
								Year:   "yyyy",
							},
						},
						Title: ChartScaleTitleOptions{
							Display: true,
							// 您可以保持或修改这里的标题文本，例如 langText("时间", "Time", lang)
							Text: langText("时间 (HH:MM)", "Time (HH:MM)", lang),
						},
					},
					Y: ChartScaleOptions{
						BeginAtZero: true,
						Title: ChartScaleTitleOptions{
							Display: true,
							Text:    yAxisLabel, // Use the metric's unit or 'Value'
						},
					},
				},
			}

			perfChartJSData := ChartJSData{Datasets: currentChartDatasets}
			datasetsJSON, err := json.Marshal(perfChartJSData)
			if err != nil {
				logger.Errorf("无法序列化性能图表 '%s' 的数据集为JSON: %v", metricName, err)
				continue // Skip this chart if serialization fails
			}
			optionsJSON, err := json.Marshal(perfChartOptions)
			if err != nil {
				logger.Errorf("无法序列化性能图表 '%s' 的选项为JSON: %v", metricName, err)
				continue // Skip this chart if serialization fails
			}

			chart := ReportChart{
				ChartID:      "performance_" + cleanMetricName,
				Type:         "line",
				DatasetsJSON: template.HTML(string(datasetsJSON)),
				OptionsJSON:  template.HTML(string(optionsJSON)),
			}
			module.Charts = append(module.Charts, chart)
			chartIndex++ // 递增图表颜色索引
		}

		if len(module.Charts) == 0 && module.Error == "" {
			module.Description = langText("虽然获取到了性能指标原始数据，但未能生成任何图表。可能是因为期望的指标数据缺失或无效。", "Although raw performance metrics data was retrieved, no charts could be generated. This might be due to missing or invalid data for the expected metrics.", lang)
			module.Cards = append(module.Cards, ReportCard{Title: langText("图表提示", "Chart Note", lang), Value: langText("未能生成性能图表", "Could not generate performance charts", lang)})
		}
	case "security":
		module.Name = langText("安全配置", "Security Configuration", lang)
		logger.Infof("开始处理安全模块... 语言: %s", lang)
		module.Cards = []ReportCard{}

		// 1. 获取非系统用户信息
		nonSystemUsers, err := db.GetNonSystemUsers(dbConn)
		if err != nil {
			logger.Errorf("处理安全模块 - 获取非系统用户信息失败: %v", err)
			module.Cards = append(module.Cards, ReportCard{
				Title: langText("非系统用户信息错误", "Non-System User Info Error", lang),
				Value: fmt.Sprintf(langText("获取非系统用户信息失败: %v", "Failed to get non-system user info: %v", lang), err),
			})
		} else {
			if len(nonSystemUsers) > 0 {
				usersTable := &ReportTable{
					Name: langText("非系统用户账户", "Non-System User Accounts", lang),
					Headers: []string{
						langText("用户名", "Username", lang),
						langText("账户状态", "Account Status", lang),
						langText("锁定日期", "Lock Date", lang),
						langText("过期日期", "Expiry Date", lang),
						langText("默认表空间", "Default Tablespace", lang),
						langText("临时表空间", "Temp Tablespace", lang),
						langText("Profile", "Profile", lang),
						langText("创建日期", "Created", lang),
						langText("上次登录", "Last Login", lang),
					},
					Rows: [][]string{},
				}
				for _, u := range nonSystemUsers {
					lockDateStr := "N/A"
					if u.LockDate.Valid {
						lockDateStr = u.LockDate.Time.Format("2006-01-02 15:04:05")
					}
					expiryDateStr := "N/A"
					if u.ExpiryDate.Valid {
						expiryDateStr = u.ExpiryDate.Time.Format("2006-01-02 15:04:05")
					}
					createdStr := u.Created.Format("2006-01-02 15:04:05")
					lastLoginStr := "N/A"
					if u.LastLogin.Valid {
						lastLoginStr = u.LastLogin.Time.Format("2006-01-02 15:04:05")
					}
					row := []string{
						u.Username,
						u.AccountStatus,
						lockDateStr,
						expiryDateStr,
						u.DefaultTablespace,
						u.TemporaryTablespace,
						u.Profile,
						createdStr,
						lastLoginStr,
					}
					usersTable.Rows = append(usersTable.Rows, row)
				}
				module.Tables = append(module.Tables, usersTable)
			} else {
				module.Cards = append(module.Cards, ReportCard{
					Title: langText("非系统用户账户", "Non-System User Accounts", lang),
					Value: langText("未发现符合条件的非系统用户账户。", "No non-system user accounts found matching criteria.", lang),
				})
			}
		}

		// 2. 获取 Profile 配置信息
		profiles, err := db.GetProfiles(dbConn)
		if err != nil {
			logger.Errorf("处理安全模块 - 获取 Profile 配置信息失败: %v", err)
			module.Cards = append(module.Cards, ReportCard{
				Title: langText("Profile 配置错误", "Profile Configuration Error", lang),
				Value: fmt.Sprintf(langText("获取 Profile 配置信息失败: %v", "Failed to get Profile configuration: %v", lang), err),
			})
		} else {
			if len(profiles) > 0 {
				profilesTable := &ReportTable{
					Name: langText("Profile 配置 (密码策略与DEFAULT)", "Profile Configuration (Password Policies & DEFAULT)", lang),
					Headers: []string{
						langText("Profile 名称", "Profile Name", lang),
						langText("资源名称", "Resource Name", lang),
						langText("限制值", "Limit", lang),
					},
					Rows: [][]string{},
				}
				for _, p := range profiles {
					row := []string{
						p.Profile,
						p.ResourceName,
						p.Limit,
					}
					profilesTable.Rows = append(profilesTable.Rows, row)
				}
				module.Tables = append(module.Tables, profilesTable)
			} else {
				module.Cards = append(module.Cards, ReportCard{
					Title: langText("Profile 配置", "Profile Configuration", lang),
					Value: langText("未发现相关的 Profile 配置信息。", "No relevant Profile configurations found.", lang),
				})
			}
		}

		// 3. 获取非系统角色列表
		nonSystemRoles, err := db.GetNonSystemRoles(dbConn)
		if err != nil {
			logger.Errorf("处理安全模块 - 获取非系统角色列表失败: %v", err)
			module.Cards = append(module.Cards, ReportCard{
				Title: langText("非系统角色错误", "Non-System Roles Error", lang),
				Value: fmt.Sprintf(langText("获取非系统角色列表失败: %v", "Failed to get non-system roles: %v", lang), err),
			})
		} else {
			if len(nonSystemRoles) > 0 {
				rolesTable := &ReportTable{
					Name: langText("非系统角色", "Non-System Roles", lang),
					Headers: []string{
						langText("角色名称", "Role Name", lang),
						langText("认证类型", "Authentication Type", lang),
					},
					Rows: [][]string{},
				}
				for _, r := range nonSystemRoles {
					row := []string{r.RoleName, r.AuthenticationType}
					rolesTable.Rows = append(rolesTable.Rows, row)
				}
				module.Tables = append(module.Tables, rolesTable)
			} else {
				module.Cards = append(module.Cards, ReportCard{
					Title: langText("非系统角色", "Non-System Roles", lang),
					Value: langText("未发现非系统角色。", "No non-system roles found.", lang),
				})
			}
		}

		// 4. 获取拥有特权角色的用户列表
		usersWithPrivRoles, err := db.GetUsersWithPrivilegedRoles(dbConn)
		if err != nil {
			logger.Errorf("处理安全模块 - 获取用户特权角色信息失败: %v", err)
			module.Cards = append(module.Cards, ReportCard{
				Title: langText("用户特权角色错误", "User Privileged Roles Error", lang),
				Value: fmt.Sprintf(langText("获取用户特权角色信息失败: %v", "Failed to get user privileged roles: %v", lang), err),
			})
		} else {
			if len(usersWithPrivRoles) > 0 {
				userPrivRolesTable := &ReportTable{
					Name: langText("拥有特权角色的用户", "Users with Privileged Roles", lang),
					Headers: []string{
						langText("用户/角色 (Grantee)", "User/Role (Grantee)", lang),
						langText("授予的角色", "Granted Role", lang),
						langText("Admin Option", "Admin Option", lang),
						langText("Default Role", "Default Role", lang),
					},
					Rows: [][]string{},
				}
				for _, upr := range usersWithPrivRoles {
					row := []string{upr.Grantee, upr.GrantedRole, upr.AdminOption, upr.DefaultRole}
					userPrivRolesTable.Rows = append(userPrivRolesTable.Rows, row)
				}
				module.Tables = append(module.Tables, userPrivRolesTable)
			} else {
				module.Cards = append(module.Cards, ReportCard{
					Title: langText("拥有特权角色的用户", "Users with Privileged Roles", lang),
					Value: langText("未发现拥有指定特权角色的非系统用户。", "No non-system users found with specified privileged roles.", lang),
				})
			}
		}

		// 5. 获取拥有系统权限的用户列表
		usersWithSysPrivs, err := db.GetUsersWithSystemPrivileges(dbConn)
		if err != nil {
			logger.Errorf("处理安全模块 - 获取用户系统权限信息失败: %v", err)
			module.Cards = append(module.Cards, ReportCard{
				Title: langText("用户系统权限错误", "User System Privileges Error", lang),
				Value: fmt.Sprintf(langText("获取用户系统权限信息失败: %v", "Failed to get user system privileges: %v", lang), err),
			})
		} else {
			if len(usersWithSysPrivs) > 0 {
				userSysPrivsTable := &ReportTable{
					Name: langText("拥有系统权限的用户", "Users with System Privileges", lang),
					Headers: []string{
						langText("用户 (Grantee)", "User (Grantee)", lang),
						langText("系统权限", "System Privilege", lang),
						langText("Admin Option", "Admin Option", lang),
					},
					Rows: [][]string{},
				}
				for _, usp := range usersWithSysPrivs {
					row := []string{usp.Grantee, usp.Privilege, usp.AdminOption}
					userSysPrivsTable.Rows = append(userSysPrivsTable.Rows, row)
				}
				module.Tables = append(module.Tables, userSysPrivsTable)
			} else {
				module.Cards = append(module.Cards, ReportCard{
					Title: langText("拥有系统权限的用户", "Users with System Privileges", lang),
					Value: langText("未发现拥有直接系统权限的非系统用户。", "No non-system users found with direct system privileges.", lang),
				})
			}
		}

		// 6. 获取角色授予角色的信息
		roleToRoleGrants, err := db.GetRoleToRoleGrants(dbConn)
		if err != nil {
			logger.Errorf("处理安全模块 - 获取角色授予角色信息失败: %v", err)
			module.Cards = append(module.Cards, ReportCard{
				Title: langText("角色授予角色错误", "Role to Role Grants Error", lang),
				Value: fmt.Sprintf(langText("获取角色授予角色信息失败: %v", "Failed to get role to role grants: %v", lang), err),
			})
		} else {
			if len(roleToRoleGrants) > 0 {
				r2rTable := &ReportTable{
					Name: langText("角色授予的角色 (非系统授予者)", "Roles Granted by Roles (Non-System Grantor)", lang),
					Headers: []string{
						langText("授予者角色", "Grantor Role", lang),
						langText("被授予角色", "Granted Role", lang),
						langText("Admin Option", "Admin Option", lang),
					},
					Rows: [][]string{},
				}
				for _, r2r := range roleToRoleGrants {
					row := []string{r2r.Role, r2r.GrantedRole, r2r.AdminOption}
					r2rTable.Rows = append(r2rTable.Rows, row)
				}
				module.Tables = append(module.Tables, r2rTable)
			} else {
				module.Cards = append(module.Cards, ReportCard{
					Title: langText("角色授予的角色", "Roles Granted by Roles", lang),
					Value: langText("未发现非系统角色授予其他角色的记录。", "No records found where non-system roles grant other roles.", lang),
				})
			}
		}

		// TODO: Further security checks like audit settings etc. can be added here

	case "backup":
		module.Name = langText("备份与恢复", "Backup & Recovery", lang)
		logger.Info("开始处理备份与恢复模块...")

		backupData := db.GetAllBackupDetails(dbConn)

		// 1. 归档模式
		if backupData.ArchivelogModeError != nil {
			logger.Errorf("处理备份模块 - 获取归档模式失败: %v", backupData.ArchivelogModeError)
			module.Cards = append(module.Cards, cardFromError("归档模式错误", "Archivelog Mode Error", backupData.ArchivelogModeError, lang))
		} else {
			module.Cards = append(module.Cards, ReportCard{
				Title: langText("数据库日志模式", "Database Log Mode", lang),
				Value: backupData.ArchivelogMode.LogMode,
			})
		}

		// 2. 闪回数据库状态
		if backupData.FlashbackStatusError != nil {
			logger.Errorf("处理备份模块 - 获取闪回状态失败: %v", backupData.FlashbackStatusError)
			module.Cards = append(module.Cards, cardFromError("闪回状态错误", "Flashback Status Error", backupData.FlashbackStatusError, lang))
		} else {
			flashbackCardValue := fmt.Sprintf("%s", backupData.FlashbackStatus.FlashbackOn)
			if backupData.FlashbackStatus.FlashbackOn == "YES" {
				if backupData.FlashbackStatus.OldestFlashbackTime.Valid {
					flashbackCardValue += fmt.Sprintf(langText(" (最早可至: %s)", " (Oldest: %s)", lang), backupData.FlashbackStatus.OldestFlashbackTime.Time.Format("2006-01-02 15:04:05"))
				}
				if backupData.FlashbackStatus.RetentionTarget.Valid {
					flashbackCardValue += fmt.Sprintf(langText(", 保留目标: %d 分钟", ", Retention: %d mins", lang), backupData.FlashbackStatus.RetentionTarget.Int64)
				}
			}
			module.Cards = append(module.Cards, ReportCard{
				Title: langText("闪回数据库状态", "Flashback Database Status", lang),
				Value: flashbackCardValue,
			})
		}

		// 3. RMAN 备份作业 (过去7天)
		if backupData.RMANJobsError != nil {
			logger.Errorf("处理备份模块 - 获取RMAN作业失败: %v", backupData.RMANJobsError)
			module.Cards = append(module.Cards, cardFromError("RMAN备份作业错误", "RMAN Backup Jobs Error", backupData.RMANJobsError, lang))
		} else {
			if len(backupData.RMANJobs) > 0 {
				rmanTable := &ReportTable{
					Name: langText("最近RMAN备份作业 (过去7天)", "Recent RMAN Backup Jobs (Last 7 Days)", lang),
					Headers: []string{
						langText("会话键", "Session Key", lang), langText("开始时间", "Start Time", lang), langText("结束时间", "End Time", lang),
						langText("状态", "Status", lang), langText("输入", "Input", lang), langText("输出", "Output", lang), langText("耗时", "Duration", lang),
						langText("优化?", "Optimized?", lang), langText("压缩率", "Compression Ratio", lang),
					},
					Rows: [][]string{},
				}
				for _, job := range backupData.RMANJobs {
					startTime := "N/A"
					if job.StartTime.Valid {
						startTime = job.StartTime.Time.Format("2006-01-02 15:04:05")
					}
					endTime := "N/A"
					if job.EndTime.Valid {
						endTime = job.EndTime.Time.Format("2006-01-02 15:04:05")
					}
					row := []string{
						fmt.Sprintf("%d", job.SessionKey),
						startTime,
						endTime,
						job.Status,
						job.InputBytesDisplay,
						job.OutputBytesDisplay,
						job.TimeTakenDisplay,
						job.Optimized,
						fmt.Sprintf("%.2f", job.CompressionRatio),
					}
					rmanTable.Rows = append(rmanTable.Rows, row)
				}
				module.Tables = append(module.Tables, rmanTable)
			} else {
				module.Cards = append(module.Cards, ReportCard{
					Title: langText("RMAN备份作业", "RMAN Backup Jobs", lang),
					Value: langText("过去7天内未发现RMAN备份作业记录。", "No RMAN backup jobs found in the last 7 days.", lang),
				})
			}
		}

		// 4. 回收站对象
		if backupData.RecycleBinError != nil {
			logger.Errorf("处理备份模块 - 获取回收站对象失败: %v", backupData.RecycleBinError)
			module.Cards = append(module.Cards, cardFromError("回收站错误", "Recycle Bin Error", backupData.RecycleBinError, lang))
		} else {
			if len(backupData.RecycleBinItems) > 0 {
				rbTable := &ReportTable{
					Name: langText("回收站对象 (可恢复)", "Recycle Bin Objects (Restorable)", lang),
					Headers: []string{
						langText("所有者", "Owner", lang), langText("对象名", "Object Name", lang), langText("原始名", "Original Name", lang),
						langText("类型", "Type", lang), langText("删除时间", "Drop Time", lang), langText("空间(块)", "Space (Blocks)", lang), langText("可恢复?", "Can Undrop?", lang),
					},
					Rows: [][]string{},
				}
				for _, item := range backupData.RecycleBinItems {
					row := []string{
						item.Owner,
						item.ObjectName,
						item.OriginalName,
						item.Type,
						item.Droptime.String,
						fmt.Sprintf("%d", item.Space),
						item.CanUndrop,
					}
					rbTable.Rows = append(rbTable.Rows, row)
				}
				module.Tables = append(module.Tables, rbTable)
			} else {
				module.Cards = append(module.Cards, ReportCard{
					Title: langText("回收站", "Recycle Bin", lang),
					Value: langText("回收站中未发现可恢复的对象。", "No restorable objects found in the recycle bin.", lang),
				})
			}
		}

		// 5. Data Pump 作业
		if backupData.DataPumpJobsError != nil {
			logger.Errorf("处理备份模块 - 获取Data Pump作业失败: %v", backupData.DataPumpJobsError)
			module.Cards = append(module.Cards, cardFromError("Data Pump作业错误", "Data Pump Jobs Error", backupData.DataPumpJobsError, lang))
		} else {
			if len(backupData.DataPumpJobs) > 0 {
				dpTable := &ReportTable{
					Name: langText("Data Pump 作业", "Data Pump Jobs", lang),
					Headers: []string{
						langText("作业名", "Job Name", lang), langText("所有者", "Owner", lang), langText("操作", "Operation", lang),
						langText("模式", "Mode", lang), langText("状态", "State", lang), langText("附加会话", "Attached Sessions", lang),
					},
					Rows: [][]string{},
				}
				for _, job := range backupData.DataPumpJobs {
					attachedSessions := "N/A"
					if job.AttachedSessions.Valid {
						attachedSessions = fmt.Sprintf("%d", job.AttachedSessions.Int64)
					}
					row := []string{
						job.JobName,
						job.OwnerName,
						job.Operation,
						job.JobMode,
						job.State,
						attachedSessions,
					}
					dpTable.Rows = append(dpTable.Rows, row)
				}
				module.Tables = append(module.Tables, dpTable)
			} else {
				module.Cards = append(module.Cards, ReportCard{
					Title: langText("Data Pump 作业", "Data Pump Jobs", lang),
					Value: langText("未发现活动的或最近的Data Pump作业。", "No active or recent Data Pump jobs found.", lang),
				})
			}
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

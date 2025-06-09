package handler

import (
	"database/sql"
	"encoding/json" // For chart data serialization
	"fmt"
	"html/template"
	"strconv"

	"github.com/goodwaysIT/inspect4oracle/internal/db"
	"github.com/goodwaysIT/inspect4oracle/internal/logger"
)

// processSessionsModule handles the "sessions" inspection item.
func processSessionsModule(dbConn *sql.DB, lang string) (cards []ReportCard, tables []*ReportTable, charts []ReportChart, err error) {
	logger.Debugf("开始处理会话模块，语言: %s", lang)

	var sessionData *db.AllSessionInfo
	var overviewErr, eventErr, historyErr error

	sessionData, overviewErr, eventErr, historyErr = db.GetSessionDetails(dbConn)

	// 1. 会话总览 - 表格
	if overviewErr != nil {
		logger.Warnf("处理会话模块 - 会话总览数据获取失败: %v", overviewErr)
		cards = append(cards, ReportCard{Title: langText("会话总览", "Session Overview", lang), Value: fmt.Sprintf(langText("获取数据失败: %v", "Failed to get data: %v", lang), overviewErr)})
		if err == nil {
			err = overviewErr
		}
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
		tables = append(tables, &sessionOverviewTable)
	} else {
		cards = append(cards, ReportCard{Title: langText("会话总览", "Session Overview", lang), Value: langText("无会话总览数据", "No session overview data available.", lang)})
	}

	// 2. 按等待事件统计的会话数 - 表格
	if eventErr != nil {
		logger.Warnf("处理会话模块 - 按等待事件统计数据获取失败: %v", eventErr)
		cards = append(cards, ReportCard{Title: langText("按等待事件统计会话数", "Session Count by Wait Event", lang), Value: fmt.Sprintf(langText("获取数据失败: %v", "Failed to get data: %v", lang), eventErr)})
		if err == nil {
			err = eventErr
		}
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
		tables = append(tables, &sessionByEventTable)
	} else {
		cards = append(cards, ReportCard{Title: langText("按等待事件统计会话数", "Session Count by Wait Event", lang), Value: langText("无按等待事件统计的会话数据", "No session count by wait event data available.", lang)})
	}

	// 3. 最近一天会话数 - Chart.js Line Chart
	if historyErr != nil {
		logger.Warnf("处理会话模块 - 获取最近活动会话历史数据失败: %v (这可能是由于ASH未启用或许可问题，图表将不会显示)", historyErr)
		cards = append(cards, ReportCard{
			Title: langText("最近活动会话历史", "Recent Active Session History", lang),
			Value: fmt.Sprintf(langText("获取图表数据失败: %v。图表无法生成。", "Failed to get chart data: %v. Chart cannot be generated.", lang), historyErr),
		})
		if err == nil {
			err = historyErr
		}
	} else if sessionData != nil && len(sessionData.HistoryForChart) > 0 {
		labels := make([]string, len(sessionData.HistoryForChart))
		dataCounts := make([]int, len(sessionData.HistoryForChart))
		for i, h := range sessionData.HistoryForChart {
			labels[i] = h.SampleTime
			dataCounts[i] = h.SessionCount
		}

		var rptDatasets []ChartDataset
		var points []ChartDataPoint
		for i, label := range labels {
			points = append(points, ChartDataPoint{X: label, Y: dataCounts[i]})
		}

		rptDatasets = append(rptDatasets, ChartDataset{
			Label:       langText("数据库会话数", "Database Session Count", lang),
			Data:        points,
			BorderColor: "rgb(75, 192, 192)",
			Fill:        false,
		})

		rptOptions := ChartJSOptions{
			Responsive:          true,
			MaintainAspectRatio: false,
			Plugins: ChartPluginsOptions{
				Title: ChartPluginTitleOptions{
					Display: true,
					Text:    langText("最近24小时活动会话数趋势", "Active Session Trend (Last 24h)", lang),
				},
				Legend: ChartPluginLegendOptions{
					Display:  true,
					Position: "top",
				},
			},
			Scales: ChartScalesOptions{
				X: ChartScaleOptions{
					Type: "time",
					Time: &ChartTimeScaleOptions{
						TooltipFormat: "yyyy-MM-dd HH:mm",
						DisplayFormats: &ChartTimeDisplayFormats{
							Minute: "HH:mm",
							Hour:   "MM-dd HH:mm",
							Day:    "yyyy-MM-dd",
							Week:   "yyyy-MM-dd",
							Month:  "yyyy-MM",
							Year:   "yyyy",
						},
					},
					Title: ChartScaleTitleOptions{
						Display: true,
						Text:    langText("时间 (HH:MM)", "Time (HH:MM)", lang),
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

		chartJSData := ChartJSData{Datasets: rptDatasets}
		datasetsJSON, jsonErr := json.Marshal(chartJSData)
		if jsonErr != nil {
			logger.Errorf("无法序列化会话图表数据集为JSON: %v", jsonErr)
			cards = append(cards, ReportCard{Title: langText("会话图表错误", "Session Chart Error", lang), Value: fmt.Sprintf(langText("无法生成图表数据: %v", "Failed to generate chart data: %v", lang), jsonErr)})
			if err == nil {
				err = jsonErr
			}
		}
		optionsJSON, jsonErr := json.Marshal(rptOptions)
		if jsonErr != nil {
			logger.Errorf("无法序列化会话图表选项为JSON: %v", jsonErr)
			cards = append(cards, ReportCard{Title: langText("会话图表配置错误", "Session Chart Config Error", lang), Value: fmt.Sprintf(langText("无法生成图表配置: %v", "Failed to generate chart config: %v", lang), jsonErr)})
			if err == nil {
				err = jsonErr
			}
		}

		if err == nil { // Only add chart if no critical error occurred during its data generation
			sessionChart := ReportChart{
				ChartID:      "sessionTrendChart",
				Type:         "line",
				DatasetsJSON: template.HTML(string(datasetsJSON)),
				OptionsJSON:  template.HTML(string(optionsJSON)),
			}
			charts = append(charts, sessionChart)
		}
	} else {
		cards = append(cards, ReportCard{Title: langText("最近活动会话历史", "Recent Active Session History", lang), Value: langText("无活动会话历史数据可供绘图", "No active session history data available for chart.", lang)})
	}
	return cards, tables, charts, err
}

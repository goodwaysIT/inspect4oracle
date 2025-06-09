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

// generateSessionOverview creates report cards and a table for session overview data.
func generateSessionOverview(sessionData *db.AllSessionInfo, fetchErr error, lang string) (cards []ReportCard, table *ReportTable, processingErr error) {
	if fetchErr != nil {
		logger.Warnf("会话总览数据获取失败: %v", fetchErr)
		cards = append(cards, ReportCard{Title: langText("会话总览", "Session Overview", lang), Value: fmt.Sprintf(langText("获取数据失败: %v", "Failed to get data: %v", lang), fetchErr)})
		return cards, nil, fetchErr
	}

	if sessionData != nil && len(sessionData.Overview) > 0 {
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
		table = &sessionOverviewTable
	} else {
		cards = append(cards, ReportCard{Title: langText("会话总览", "Session Overview", lang), Value: langText("无会话总览数据", "No session overview data available.", lang)})
	}
	return cards, table, nil
}

// generateSessionByEvent creates report cards and a table for session count by wait event data.
func generateSessionByEvent(sessionData *db.AllSessionInfo, fetchErr error, lang string) (cards []ReportCard, table *ReportTable, processingErr error) {
	if fetchErr != nil {
		logger.Warnf("按等待事件统计数据获取失败: %v", fetchErr)
		cards = append(cards, ReportCard{Title: langText("按等待事件统计会话数", "Session Count by Wait Event", lang), Value: fmt.Sprintf(langText("获取数据失败: %v", "Failed to get data: %v", lang), fetchErr)})
		return cards, nil, fetchErr
	}

	if sessionData != nil && len(sessionData.ByEvent) > 0 {
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
		table = &sessionByEventTable
	} else {
		cards = append(cards, ReportCard{Title: langText("按等待事件统计会话数", "Session Count by Wait Event", lang), Value: langText("无按等待事件统计的会话数据", "No session count by wait event data available.", lang)})
	}
	return cards, table, nil
}

// generateSessionHistoryChart creates report cards and a chart for recent active session history.
func generateSessionHistoryChart(sessionData *db.AllSessionInfo, fetchErr error, lang string) (cards []ReportCard, chart *ReportChart, processingErr error) {
	if fetchErr != nil {
		logger.Warnf("获取最近活动会话历史数据失败: %v (这可能是由于ASH未启用或许可问题，图表将不会显示)", fetchErr)
		cards = append(cards, ReportCard{
			Title: langText("最近活动会话历史", "Recent Active Session History", lang),
			Value: fmt.Sprintf(langText("获取图表数据失败: %v。图表无法生成。", "Failed to get chart data: %v. Chart cannot be generated.", lang), fetchErr),
		})
		return cards, nil, fetchErr
	}

	if sessionData != nil && len(sessionData.HistoryForChart) > 0 {
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
			return cards, nil, jsonErr // Return JSON marshaling error
		}
		optionsJSON, jsonErr := json.Marshal(rptOptions)
		if jsonErr != nil {
			logger.Errorf("无法序列化会话图表选项为JSON: %v", jsonErr)
			cards = append(cards, ReportCard{Title: langText("会话图表配置错误", "Session Chart Config Error", lang), Value: fmt.Sprintf(langText("无法生成图表配置: %v", "Failed to generate chart config: %v", lang), jsonErr)})
			return cards, nil, jsonErr // Return JSON marshaling error
		}

		sessionChart := ReportChart{
			ChartID:      "sessionTrendChart",
			Type:         "line",
			DatasetsJSON: template.HTML(string(datasetsJSON)),
			OptionsJSON:  template.HTML(string(optionsJSON)),
		}
		chart = &sessionChart
	} else {
		cards = append(cards, ReportCard{Title: langText("最近活动会话历史", "Recent Active Session History", lang), Value: langText("无活动会话历史数据可供绘图", "No active session history data available for chart.", lang)})
	}
	return cards, chart, nil
}

// processSessionsModule handles the "sessions" inspection item.
func processSessionsModule(dbConn *sql.DB, lang string) (allCards []ReportCard, allTables []*ReportTable, allCharts []ReportChart, overallErr error) {
	logger.Debugf("开始处理会话模块，语言: %s", lang)

	sessionData, overviewFetchErr, eventFetchErr, historyFetchErr := db.GetSessionDetails(dbConn)

	// Helper to manage overall error, ensuring we capture the first non-nil error.
	setOverallErr := func(e error) {
		if overallErr == nil && e != nil {
			overallErr = e
		}
	}

	// 1. Session Overview
	overviewCards, overviewTable, overviewProcErr := generateSessionOverview(sessionData, overviewFetchErr, lang)
	allCards = append(allCards, overviewCards...)
	if overviewTable != nil {
		allTables = append(allTables, overviewTable)
	}
	setOverallErr(overviewProcErr)

	// 2. Session By Event
	byEventCards, byEventTable, byEventProcErr := generateSessionByEvent(sessionData, eventFetchErr, lang)
	allCards = append(allCards, byEventCards...)
	if byEventTable != nil {
		allTables = append(allTables, byEventTable)
	}
	setOverallErr(byEventProcErr)

	// 3. Session History Chart
	historyCards, historyChart, historyProcErr := generateSessionHistoryChart(sessionData, historyFetchErr, lang)
	allCards = append(allCards, historyCards...)
	if historyChart != nil {
		allCharts = append(allCharts, *historyChart) // ReportChart is not a pointer in allCharts slice
	}
	setOverallErr(historyProcErr)

	return allCards, allTables, allCharts, overallErr
}

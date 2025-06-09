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
	{BorderColor: "#ffc107", BackgroundColor: "rgba(255, 193, 7, 0.2)"},   // Yellow
	{BorderColor: "#dc3545", BackgroundColor: "rgba(220, 53, 69, 0.2)"},   // Red
	{BorderColor: "#17a2b8", BackgroundColor: "rgba(23, 162, 184, 0.2)"},  // Cyan
	{BorderColor: "#6c757d", BackgroundColor: "rgba(108, 117, 125, 0.2)"}, // Grey
	{BorderColor: "#fd7e14", BackgroundColor: "rgba(253, 126, 20, 0.2)"},  // Orange
	{BorderColor: "#6610f2", BackgroundColor: "rgba(102, 16, 242, 0.2)"},  // Indigo
	{BorderColor: "#e83e8c", BackgroundColor: "rgba(232, 62, 140, 0.2)"},  // Pink
	{BorderColor: "#20c997", BackgroundColor: "rgba(32, 201, 151, 0.2)"},  // Teal
}

// processPerformanceModule 处理性能模块的逻辑，获取指标数据并生成图表
func processPerformanceModule(dbConn *sql.DB, lang string) ([]ReportCard, []*ReportTable, []ReportChart, error) {
	var cards []ReportCard
	var tables []*ReportTable // Performance module currently doesn't generate tables, but we keep the signature consistent
	var charts []ReportChart
	var overallErr error

	logger.Info("开始处理性能模块...")

	// 1. 获取所有性能指标数据
	metricsBundle := db.GetAllPerformanceMetrics(dbConn)
	metricsData := metricsBundle.SysMetricsSummary
	err := metricsBundle.SysMetricsError
	if err != nil {
		errMsg := langText("获取性能指标数据失败", "Failed to retrieve performance metrics data", lang)
		logger.Errorf("%s: %v", errMsg, err)
		cards = append(cards, cardFromError(langText("性能指标错误", "Performance Metrics Error", lang), errMsg, err, lang))
		overallErr = err
		// 不直接返回，允许后续处理其他可能的metrics数据，如果GetAllPerformanceMetrics将来返回多种metrics
	}

	if len(metricsData) == 0 && overallErr == nil { // 如果有错误，则错误卡片已添加
		logger.Warn("性能模块: 未获取到任何性能指标数据")
		cards = append(cards, ReportCard{
			Title: langText("性能指标", "Performance Metrics", lang),
			Value: langText("未获取到任何性能指标数据。", "No performance metrics data was retrieved.", lang),
		})
		// 不返回，因为可能还有其他metrics数据或错误
	}

	// Group metrics by metric name
	metricsMap := make(map[string][]db.SysMetricSummary)
	for _, metric := range metricsData {
		if !metric.MetricName.Valid {
			logger.Warnf("Skipping metric with nil MetricName")
			continue
		}
		metricsMap[metric.MetricName.String] = append(metricsMap[metric.MetricName.String], metric)
	}

	// Create a chart for each metric
	colorIndex := 0
	for metricNameStr, metricValues := range metricsMap {
		if len(metricValues) == 0 {
			continue
		}

		dataPoints := make([]ChartDataPoint, 0, len(metricValues))
		var metricUnit string

		for i, mv := range metricValues {
			if !mv.BeginTimeStr.Valid || !mv.ValueStr.Valid {
				logger.Warnf("Skipping metric data point with nil BeginTimeStr or ValueStr for metric: %s", metricNameStr)
				continue
			}

			val, errConv := strconv.ParseFloat(mv.ValueStr.String, 64)
			if errConv != nil {
				logger.Warnf("Failed to parse metric value '%s' to float for metric '%s': %v. Skipping data point.", mv.ValueStr.String, metricNameStr, errConv)
				continue
			}
			dataPoints = append(dataPoints, ChartDataPoint{X: mv.BeginTimeStr.String, Y: val})

			if i == 0 && mv.MetricUnit.Valid {
				metricUnit = mv.MetricUnit.String
			}
		}

		if len(dataPoints) == 0 {
			logger.Warnf("No valid data points to plot for metric: %s", metricNameStr)
			continue
		}

		chartID := "chart-perf-" + strings.ToLower(strings.ReplaceAll(metricNameStr, " ", "-"))
		chartID = strings.ReplaceAll(chartID, "(", "")
		chartID = strings.ReplaceAll(chartID, ")", "")
		chartID = strings.ReplaceAll(chartID, "%", "percent")
		chartID = strings.ReplaceAll(chartID, "/", "_per_")

		chartTitle := metricNameStr
		yAxisLabel := metricUnit
		if yAxisLabel == "" {
			yAxisLabel = langText("值", "Value", lang)
		}
		if metricUnit != "" { // Append unit to title if present
			chartTitle = fmt.Sprintf("%s (%s)", metricNameStr, metricUnit)
		}

		selectedColor := performanceChartColors[colorIndex%len(performanceChartColors)]
		currentChartDatasets := []ChartDataset{
			{
				Label:           metricNameStr,
				Data:            dataPoints,
				BorderColor:     selectedColor.BorderColor,
				BackgroundColor: selectedColor.BackgroundColor,
				Fill:            true,
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
					Display:  len(currentChartDatasets) > 1,
					Position: "top",
				},
			},
			Scales: ChartScalesOptions{
				X: ChartScaleOptions{
					Type: "time",
					Time: &ChartTimeScaleOptions{
						Unit:          "hour", // Default to hour, can be adjusted
						TooltipFormat: "yyyy-MM-dd HH:mm",
						DisplayFormats: &ChartTimeDisplayFormats{
							Minute: "HH:mm",
							Hour:   "MM-dd HH:mm",
							Day:    "yyyy-MM-dd",
						},
					},
					Title: ChartScaleTitleOptions{
						Display: true,
						Text:    langText("时间", "Time", lang),
					},
				},
				Y: ChartScaleOptions{
					BeginAtZero: true,
					Title: ChartScaleTitleOptions{
						Display: true,
						Text:    yAxisLabel,
					},
				},
			},
		}

		perfChartJSData := ChartJSData{Datasets: currentChartDatasets}
		datasetsJSON, errMarshal := json.Marshal(perfChartJSData)
		if errMarshal != nil {
			logger.Errorf("无法序列化性能图表 '%s' 的数据集为JSON: %v", metricNameStr, errMarshal)
			overallErr = fmt.Errorf("failed to marshal chart datasets for %s: %w", metricNameStr, errMarshal) // Capture error
			continue
		}
		optionsJSON, errMarshal := json.Marshal(perfChartOptions)
		if errMarshal != nil {
			logger.Errorf("无法序列化性能图表 '%s' 的选项为JSON: %v", metricNameStr, errMarshal)
			overallErr = fmt.Errorf("failed to marshal chart options for %s: %w", metricNameStr, errMarshal) // Capture error
			continue
		}

		chart := ReportChart{
			ChartID:      chartID,
			Type:         "line",
			DatasetsJSON: template.HTML(string(datasetsJSON)),
			OptionsJSON:  template.HTML(string(optionsJSON)),
		}
		charts = append(charts, chart)
		colorIndex++
	}

	if len(charts) == 0 && overallErr == nil && len(metricsData) > 0 { // Add note only if data was present but no charts made
		cards = append(cards, ReportCard{
			Title: langText("图表提示", "Chart Note", lang),
			Value: langText("虽然获取到了性能指标原始数据，但未能生成任何图表。可能是因为期望的指标数据缺失或无效。", "Although raw performance metrics data was retrieved, no charts could be generated. This might be due to missing or invalid data for the expected metrics.", lang),
		})
	}

	logger.Info("性能模块处理完成.")
	return cards, tables, charts, overallErr
}

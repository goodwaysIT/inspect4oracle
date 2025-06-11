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

// chartColor defines the border and background colors for charts.
type chartColor struct {
	BorderColor     string
	BackgroundColor string
}

// performanceChartColors provides a predefined set of colors for performance charts.
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

// generatePerformanceChart generates a single performance chart for a given metric.
func generatePerformanceChart(metricNameStr string, metricValues []db.SysMetricSummary, lang string, colorIndex int) (*ReportChart, error) {
	if len(metricValues) == 0 {
		return nil, nil // No data, no chart, no error
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
		return nil, nil // No valid data, no chart, no error
	}

	chartID := "chart-perf-" + strings.ToLower(strings.ReplaceAll(metricNameStr, " ", "-"))
	chartID = strings.ReplaceAll(chartID, "(", "")
	chartID = strings.ReplaceAll(chartID, ")", "")
	chartID = strings.ReplaceAll(chartID, "%", "percent")
	chartID = strings.ReplaceAll(chartID, "/", "_per_")

	chartTitle := metricNameStr
	yAxisLabel := metricUnit
	if yAxisLabel == "" {
		yAxisLabel = langText("值", "Value", "値", lang)
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
					Text:    langText("时间", "Time", "時間", lang),
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
		logger.Errorf("Failed to serialize dataset for performance chart '%s' to JSON: %v", metricNameStr, errMarshal)
		return nil, fmt.Errorf("failed to marshal chart datasets for %s: %w", metricNameStr, errMarshal)
	}
	optionsJSON, errMarshal := json.Marshal(perfChartOptions)
	if errMarshal != nil {
		logger.Errorf("Failed to serialize options for performance chart '%s' to JSON: %v", metricNameStr, errMarshal)
		return nil, fmt.Errorf("failed to marshal chart options for %s: %w", metricNameStr, errMarshal)
	}

	return &ReportChart{
		ChartID:      chartID,
		Type:         "line",
		DatasetsJSON: template.HTML(string(datasetsJSON)),
		OptionsJSON:  template.HTML(string(optionsJSON)),
	}, nil
}

// processPerformanceModule handles the logic for the performance module, fetching metric data and generating charts.
func processPerformanceModule(dbConn *sql.DB, lang string) ([]ReportCard, []*ReportTable, []ReportChart, error) {
	var cards []ReportCard
	var tables []*ReportTable // Performance module currently doesn't generate tables, but we keep the signature consistent
	var charts []ReportChart
	var overallErr error

	logger.Info("Starting to process performance module...")

	// 1. Get all performance metrics data
	metricsBundle := db.GetAllPerformanceMetrics(dbConn)
	metricsData := metricsBundle.SysMetricsSummary
	err := metricsBundle.SysMetricsError
	if err != nil {
		logger.Errorf("%s: %v", langText("Failed to retrieve performance metrics data", "Failed to retrieve performance metrics data", "Failed to retrieve performance metrics data", lang), err)
		cards = append(cards, cardFromError("Performance Metrics Error", "Performance Metrics Error", "Performance Metrics Error", err, lang))
		overallErr = err
		// Do not return directly, allow subsequent processing of other possible metrics data if GetAllPerformanceMetrics returns multiple metrics in the future.
	}

	if len(metricsData) == 0 && overallErr == nil { // If there is an error, the error card has already been added
		logger.Warn("Performance module: No performance metric data was retrieved.")
		cards = append(cards, ReportCard{
			Title: langText("性能指标", "Performance Metrics", "パフォーマンスメトリクス", lang),
			Value: langText("No performance metrics data was retrieved.", "No performance metrics data was retrieved.", "パフォーマンスメトリクスデータが取得されませんでした。", lang),
		})
		// Do not return, as there may be other metrics data or errors
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
		chart, errGenChart := generatePerformanceChart(metricNameStr, metricValues, lang, colorIndex)
		if errGenChart != nil {
			logger.Errorf("Failed to generate performance chart '%s': %v", metricNameStr, errGenChart)
			if overallErr == nil {
				overallErr = errGenChart
			} else {
				overallErr = fmt.Errorf("%v; %w", overallErr, errGenChart)
			}
			continue // Continue to next metric even if one chart fails
		}
		if chart != nil {
			charts = append(charts, *chart)
			colorIndex++
		}
	}

	if len(charts) == 0 && overallErr == nil && len(metricsData) > 0 { // Add note only if data was present but no charts made
		cards = append(cards, ReportCard{
			Title: langText("图表提示", "Chart Note", "チャートノート", lang),
			Value: langText("Although raw performance metrics data was retrieved, no charts could be generated. This might be due to missing or invalid data for the expected metrics.", "Although raw performance metrics data was retrieved, no charts could be generated. This might be due to missing or invalid data for the expected metrics.", "Although raw performance metrics data was retrieved, no charts could be generated. This might be due to missing or invalid data for the expected metrics.", lang),
		})
	}

	logger.Info("Performance module processing completed.")
	return cards, tables, charts, overallErr
}

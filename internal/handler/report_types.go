// Package handler defines types used for report generation.
package handler

import "html/template"

// ReportCard defines a simple key-value card for display.
// Note: This was previously an anonymous struct in module_processor.go, promoting to a named type.
type ReportCard struct {
	Title string `json:"title"`
	Value string `json:"value"`
}

// ReportTable defines the structure for a table in a report.
type ReportTable struct {
	Name    string     `json:"name"`            // 表格名称
	Headers []string   `json:"headers"`         // 表头
	Rows    [][]string `json:"rows"`            // 表格数据行
	Notes   string     `json:"notes,omitempty"` // 表格的额外注释
}

// ChartDataPoint represents a single point in a chart.
type ChartDataPoint struct {
	X interface{} `json:"x"` // Can be a timestamp (string/int) or category name (string)
	Y interface{} `json:"y"` // Typically a numerical value
}

// ChartDataset represents a single dataset in a chart.
type ChartDataset struct {
	Label           string           `json:"label,omitempty"`
	Data            []ChartDataPoint `json:"data"`
	BorderColor     string           `json:"borderColor,omitempty"`
	BackgroundColor string           `json:"backgroundColor,omitempty"`
	Fill            bool             `json:"fill,omitempty"`
	YAxisID         string           `json:"yAxisID,omitempty"`
}

// ChartScaleTitleOptions defines options for the title of a scale (axis).
type ChartScaleTitleOptions struct {
	Display bool   `json:"display,omitempty"`
	Text    string `json:"text,omitempty"`
}

// ChartTimeDisplayFormats defines the display formats for different time units.
type ChartTimeDisplayFormats struct {
	Minute string `json:"minute,omitempty"`
	Hour   string `json:"hour,omitempty"`
	Day    string `json:"day,omitempty"`
	Week   string `json:"week,omitempty"`
	Month  string `json:"month,omitempty"`
	Year   string `json:"year,omitempty"`
}

// ChartTimeScaleOptions defines options specific to a time scale.
type ChartTimeScaleOptions struct {
	Unit           string                   `json:"unit,omitempty"`           // e.g., "day", "month", "year", "minute"
	TooltipFormat  string                   `json:"tooltipFormat,omitempty"`  // Format for the tooltip
	DisplayFormats *ChartTimeDisplayFormats `json:"displayFormats,omitempty"` // Pointer to allow omission
}

// ChartScaleOptions defines options for an individual scale (axis).
type ChartScaleOptions struct {
	Type        string                 `json:"type,omitempty"`    // e.g., "linear", "logarithmic", "category", "time"
	Display     interface{}            `json:"display,omitempty"` // Can be bool or "auto"
	BeginAtZero bool                   `json:"beginAtZero,omitempty"`
	Title       ChartScaleTitleOptions `json:"title,omitempty"`
	Time        *ChartTimeScaleOptions `json:"time,omitempty"` // Pointer to allow omission if not a time scale
}

// ChartScalesOptions defines options for all scales (axes).
type ChartScalesOptions struct {
	X ChartScaleOptions `json:"x,omitempty"`
	Y ChartScaleOptions `json:"y,omitempty"`
}

// ChartPluginTitleOptions defines options for the chart title plugin.
type ChartPluginTitleOptions struct {
	Display bool   `json:"display,omitempty"`
	Text    string `json:"text,omitempty"`
}

// ChartPluginLegendOptions defines options for the chart legend plugin.
type ChartPluginLegendOptions struct {
	Display  bool   `json:"display,omitempty"`
	Position string `json:"position,omitempty"` // e.g., "top", "bottom", "left", "right"
}

// ChartPluginTooltipOptions defines options for the chart tooltip plugin.
type ChartPluginTooltipOptions struct {
	Enabled   bool   `json:"enabled,omitempty"`
	Mode      string `json:"mode,omitempty"` // e.g., "index", "point", "nearest"
	Intersect bool   `json:"intersect,omitempty"`
}

// ChartPluginsOptions defines options for various chart plugins.
type ChartPluginsOptions struct {
	Title   ChartPluginTitleOptions   `json:"title,omitempty"`
	Legend  ChartPluginLegendOptions  `json:"legend,omitempty"`
	Tooltip ChartPluginTooltipOptions `json:"tooltip,omitempty"`
}

// ChartJSOptions contains detailed configuration for how a Chart.js chart should be displayed.
type ChartJSOptions struct {
	Responsive          bool                `json:"responsive,omitempty"`
	MaintainAspectRatio bool                `json:"maintainAspectRatio,omitempty"`
	Plugins             ChartPluginsOptions `json:"plugins,omitempty"`
	Scales              ChartScalesOptions  `json:"scales,omitempty"`
}

// ChartJSData is the top-level structure for Chart.js data object.
type ChartJSData struct {
	// Labels   []string         `json:"labels,omitempty"` // Use for charts where X values are categories not part of dataset points
	Datasets []ChartDataset `json:"datasets"`
}

// ReportChart represents data and configuration for a single chart.
type ReportChart struct {
	ChartID      string        `json:"chartId"`
	Type         string        `json:"type"`                  // e.g., "line", "bar"
	DatasetsJSON template.HTML `json:"datasetsJson"`          // JSON string for chart datasets
	OptionsJSON  template.HTML `json:"optionsJson,omitempty"` // JSON string for chart options
}

// ReportModule defines the structure for a report module.
// The ID field was added based on its usage in module_processor.go
type ReportModule struct {
	ID          string         `json:"id"`                    // 模块内部唯一ID (e.g., "dbinfo", "performance")
	Name        string         `json:"name"`                  // 模块内部名称 (e.g., "dbinfo") - Potentially redundant with ID, review usage
	Title       string         `json:"title"`                 // 模块显示标题 (e.g., "数据库基本信息")
	Icon        string         `json:"icon,omitempty"`        // Font Awesome icon class (e.g., "fas fa-database")
	Cards       []ReportCard   `json:"cards,omitempty"`       // 信息卡片列表
	Tables      []*ReportTable `json:"tables,omitempty"`      // 表格列表 (Changed to slice of pointers based on module_processor.go usage)
	Charts      []ReportChart  `json:"charts,omitempty"`      // 图表列表 (New field)
	Error       string         `json:"error,omitempty"`       // 模块级别错误信息
	Description string         `json:"description,omitempty"` // 模块描述或摘要信息
}

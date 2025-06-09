// Package db handles database interactions for performance metrics.
package db

import (
	"database/sql"

	"github.com/goodwaysIT/inspect4oracle/internal/logger"
)

// SysMetricSummary holds data from DBA_HIST_SYSMETRIC_SUMMARY.
// METRIC_ID and INTSIZE_CSEC are also available but often not directly used for high-level charts.
type SysMetricSummary struct {
	SnapID               sql.NullString `json:"snap_id"`         // Changed to NullString to avoid driver panic
	DBID                 sql.NullString `json:"dbid"`            // Changed to NullString
	InstanceNumber       sql.NullString `json:"instance_number"` // Changed to NullString
	BeginTimeStr         sql.NullString `json:"begin_time_str"`  // Diagnostic: scan as string
	EndTimeStr           sql.NullString `json:"end_time_str"`    // Diagnostic: scan as string
	MetricName           sql.NullString `json:"metric_name"`
	MetricUnit           sql.NullString `json:"metric_unit"`
	ValueStr             sql.NullString `json:"value_str"`              // Raw string for AVERAGE
	MaxValStr            sql.NullString `json:"max_val_str"`            // Raw string for MAXVAL
	StandardDeviationStr sql.NullString `json:"standard_deviation_str"` // Raw string for STANDARD_DEVIATION
	// Parsed values can be added here later if needed, e.g.:
	// Value             sql.NullFloat64 `json:"value,omitempty"`
	// MaxVal            sql.NullFloat64 `json:"max_val,omitempty"`
	// StandardDeviation sql.NullFloat64 `json:"standard_deviation,omitempty"`
}

// PerformanceMetricsBundle holds all performance-related data fetched for the report.
type PerformanceMetricsBundle struct {
	SysMetricsSummary []SysMetricSummary `json:"sys_metrics_summary"`
	SysMetricsError   error              `json:"sys_metrics_error"`
}

// GetSysMetricSummary retrieves data from DBA_HIST_SYSMETRIC_SUMMARY for the last 24 hours
// for a predefined set of important metrics.
func GetSysMetricSummary(db *sql.DB) ([]SysMetricSummary, error) {
	query := `
	SELECT
	    BEGIN_TIME,
	    END_TIME,
	    METRIC_NAME,
	    METRIC_UNIT,
	    AVERAGE as VALUE -- Using AVERAGE as the primary value for the summary
	FROM
	    DBA_HIST_SYSMETRIC_SUMMARY
	WHERE
	    END_TIME >= SYSDATE - 1
	    AND METRIC_NAME IN (
	        'DB Time Per Sec',
	        'Average Active Sessions',
	        'CPU Usage Per Sec',
	        'Host CPU Utilization (%)',
	        'Executions Per Sec',
	        'User Commits Per Sec',
	        'Physical Read Total Bytes Per Sec',
	        'Physical Write Total Bytes Per Sec',
	        'Redo Generated Per Sec',
	        'Network Traffic Volume Per Sec',
	        'SQL Service Response Time',
	        'Database CPU Time Ratio',
	        'Physical Reads Per Sec',
	        'Physical Writes Per Sec',
	        'Logons Cumulative',
	        'User Rollbacks Per Sec',
	        'DB Block Changes Per Sec',
	        'GC CR Block Received Per Second',
	        'Logical Reads Per Sec',
	        'PGA Cache Hit %',
	        'Total PGA Used for Workareas'
	    )
	ORDER BY
	    METRIC_NAME, BEGIN_TIME`

	rows, err := db.Query(query)
	if err != nil {
		logger.Errorf("Error querying DBA_HIST_SYSMETRIC_SUMMARY: %v", err)
		return nil, err
	}
	defer rows.Close()

	var metrics []SysMetricSummary
	var firstScanError error // To store the first error encountered during scanning
	for rows.Next() {
		var m SysMetricSummary
		err := rows.Scan(
			&m.BeginTimeStr,
			&m.EndTimeStr,
			&m.MetricName,
			&m.MetricUnit,
			&m.ValueStr,
		)
		if err != nil {
			// Attempt to log more details, note that m.SnapID etc. might not be populated if scan failed early for them
			// SnapID is now NullString, so using .String for logging
			logger.Errorf("Error scanning DBA_HIST_SYSMETRIC_SUMMARY row (MetricName: %s, BeginTimeStr: %s, EndTimeStr: %s): %v", m.MetricName.String, m.BeginTimeStr.String, m.EndTimeStr.String, err)
			if firstScanError == nil { // Store the first error
				firstScanError = err
			}
			continue // Skipping problematic row but error is now tracked
		}
		metrics = append(metrics, m)
	}

	if err = rows.Err(); err != nil { // This checks for errors encountered during iteration (e.g., connection issue)
		logger.Errorf("Error after iterating DBA_HIST_SYSMETRIC_SUMMARY rows: %v", err)
		// If rows.Err() is not nil, it's usually more critical than a single scan error.
		// Return this error, potentially masking firstScanError if it was also set.
		return metrics, err
	}

	// If there was a scan error but iteration completed without other errors, return the first scan error.
	if firstScanError != nil {
		logger.Warnf("Returning partial metrics from DBA_HIST_SYSMETRIC_SUMMARY due to scan error(s). First error: %v", firstScanError)
		return metrics, firstScanError
	}

	return metrics, nil
}

// GetAllPerformanceMetrics aggregates all performance related metrics.
// Currently, it only fetches SysMetricSummary.
func GetAllPerformanceMetrics(db *sql.DB) PerformanceMetricsBundle {
	var bundle PerformanceMetricsBundle
	bundle.SysMetricsSummary, bundle.SysMetricsError = GetSysMetricSummary(db)
	return bundle
}

package db

import (
	"database/sql"
	"fmt"

	"github.com/goodwaysIT/inspect4oracle/internal/logger"
)

// SessionOverview contains overview information for current sessions
// Corresponds to: select inst_id,username,machine,status,count(*) as session_count from gv$session group by inst_id,username,machine,status;
type SessionOverview struct {
	InstID       int            `json:"inst_id"`
	Username     sql.NullString `json:"username"` // Can be NULL
	Machine      sql.NullString `json:"machine"`  // Can be NULL
	Status       string         `json:"status"`
	SessionCount int            `json:"session_count"`
}

// SessionEventCount contains the session count grouped by wait event
// Corresponds to: select event,count(*) as session_count from gv$session group by event order by 1;
type SessionEventCount struct {
	Event        string `json:"event"`
	SessionCount int    `json:"session_count"`
}

// SessionHistoryPoint contains the session count at a specific point in time, for charting
// Corresponds to: select to_char(sample_time,'yyyy-mm-dd hh24:mi') as sample_time,count(*) as session_count
// from gv$active_session_history where sample_time > sysdate - 1
// group by to_char(sample_time,'yyyy-mm-dd hh24:mi') order by 1;
type SessionHistoryPoint struct {
	SampleTime   string `json:"sample_time"`
	SessionCount int    `json:"session_count"`
}

// AllSessionInfo contains all session-related information to be passed to the handler
type AllSessionInfo struct {
	Overview        []SessionOverview
	ByEvent         []SessionEventCount
	HistoryForChart []SessionHistoryPoint
}

// getCurrentSessionOverview gets the current session overview
func getCurrentSessionOverview(db *sql.DB) ([]SessionOverview, error) {
	query := `
SELECT 
    inst_id AS InstID, 
    username AS Username, 
    machine AS Machine, 
    status AS Status, 
    count(*) AS SessionCount 
FROM gv$session 
GROUP BY inst_id, username, machine, status 
ORDER BY inst_id, username, machine, status`
	var overview []SessionOverview
	err := ExecuteQueryAndScanToStructs(db, &overview, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get current session overview: %w", err)
	}
	logger.Infof("Successfully fetched overview for %d sessions.", len(overview))
	logger.Debugf("Session overview info: %v", overview)
	return overview, nil
}

// getSessionCountByEvent gets the session count grouped by wait event
func getSessionCountByEvent(db *sql.DB) ([]SessionEventCount, error) {
	query := `
SELECT 
    event AS Event, 
    count(*) AS SessionCount 
FROM gv$session 
GROUP BY event 
ORDER BY SessionCount DESC, event` // Order by count desc for better readability
	var byEvent []SessionEventCount
	err := ExecuteQueryAndScanToStructs(db, &byEvent, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get session count by wait event: %w", err)
	}
	logger.Infof("Successfully fetched wait event statistics for %d sessions.", len(byEvent))
	logger.Debugf("Wait event statistics: %v", byEvent)
	return byEvent, nil
}

// getDailySessionHistory gets the session history for the last day, for charting
func getDailySessionHistory(db *sql.DB) ([]SessionHistoryPoint, error) {
	query := `
SELECT 
    to_char(sample_time, 'yyyy-mm-dd hh24:mi') AS SampleTime, 
    count(*) AS SessionCount 
FROM gv$active_session_history 
WHERE sample_time > sysdate - INTERVAL '1' DAY 
GROUP BY to_char(sample_time, 'yyyy-mm-dd hh24:mi') 
ORDER BY SampleTime`
	var history []SessionHistoryPoint // Ensure history is always initialized
	err := ExecuteQueryAndScanToStructs(db, &history, query)
	if err != nil {
		// Log a warning if ASH is not available or licensed.
		logger.Warnf("Failed to get session history (ASH) (could be due to ASH not being enabled or license issues): %v", err)
		// Return the (empty) history slice and the error, so caller knows an attempt was made.
		return history, fmt.Errorf("failed to get session history (ASH): %w", err)
	}
	logger.Infof("Successfully fetched %d session history points.", len(history))
	//logger.Debugf("Session history info: %v", history)
	return history, nil
}

// GetSessionDetails gets all session-related information
// Returns AllSessionInfo and a separate error status for each sub-query
func GetSessionDetails(db *sql.DB) (allInfo *AllSessionInfo, overviewErr error, eventErr error, historyErr error) {
	logger.Info("Starting to fetch session module information...")
	allInfo = &AllSessionInfo{}

	allInfo.Overview, overviewErr = getCurrentSessionOverview(db)
	if overviewErr != nil {
		logger.Warnf("Error fetching session overview: %v", overviewErr)
		// overviewErr is returned directly
	}

	allInfo.ByEvent, eventErr = getSessionCountByEvent(db)
	if eventErr != nil {
		logger.Warnf("Error fetching session count by event: %v", eventErr)
		// eventErr is returned directly
	}

	allInfo.HistoryForChart, historyErr = getDailySessionHistory(db)
	if historyErr != nil {
		logger.Warnf("Error fetching session history for chart: %v", historyErr)
		// historyErr is returned directly. ASH data is often considered optional.
	}

	logger.Info("Session module information fetching complete.")
	// The function now returns individual errors. The caller can decide how to handle them.
	// A general error is no longer aggregated here unless a specific design requires it.
	return
}

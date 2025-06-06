package db

import (
	"database/sql"
	"fmt"

	"github.com/goodwaysIT/inspect4oracle/internal/logger"
)

// SessionOverview 包含当前会话的概述信息
// Corresponds to: select inst_id,username,machine,status,count(*) as session_count from gv$session group by inst_id,username,machine,status;
type SessionOverview struct {
	InstID       int            `json:"inst_id"`
	Username     sql.NullString `json:"username"` // Can be NULL
	Machine      sql.NullString `json:"machine"`  // Can be NULL
	Status       string         `json:"status"`
	SessionCount int            `json:"session_count"`
}

// SessionEventCount 包含按等待事件统计的会话数
// Corresponds to: select event,count(*) as session_count from gv$session group by event order by 1;
type SessionEventCount struct {
	Event        string `json:"event"`
	SessionCount int    `json:"session_count"`
}

// SessionHistoryPoint 包含特定时间点的会话数，用于图表
// Corresponds to: select to_char(sample_time,'yyyy-mm-dd hh24:mi') as sample_time,count(*) as session_count
// from gv$active_session_history where sample_time > sysdate - 1
// group by to_char(sample_time,'yyyy-mm-dd hh24:mi') order by 1;
type SessionHistoryPoint struct {
	SampleTime   string `json:"sample_time"`
	SessionCount int    `json:"session_count"`
}

// AllSessionInfo 包含所有会话相关的信息，用于传递给处理器
type AllSessionInfo struct {
	Overview        []SessionOverview
	ByEvent         []SessionEventCount
	HistoryForChart []SessionHistoryPoint
}

// getCurrentSessionOverview 获取当前会话概述
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
		return nil, fmt.Errorf("获取当前会话概述失败: %w", err)
	}
	logger.Infof("成功获取 %d 个会话的概述信息。", len(overview))
	logger.Debugf("会话概述信息: %v", overview)
	return overview, nil
}

// getSessionCountByEvent 获取按等待事件统计的会话数
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
		return nil, fmt.Errorf("获取按等待事件统计的会话数失败: %w", err)
	}
	logger.Infof("成功获取 %d 个会话的等待事件统计信息。", len(byEvent))
	logger.Debugf("等待事件统计信息: %v", byEvent)
	return byEvent, nil
}

// getDailySessionHistory 获取最近一天的会话历史记录，用于图表
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
		logger.Warnf("获取会话历史(ASH)失败 (可能是ASH未启用或许可问题): %v", err)
		// Return the (empty) history slice and the error, so caller knows an attempt was made.
		return history, fmt.Errorf("获取会话历史(ASH)失败: %w", err)
	}
	logger.Infof("成功获取 %d 个会话的等待事件统计信息。", len(history))
	//logger.Debugf("会话历史信息: %v", history)
	return history, nil
}

// GetSessionDetails 获取所有会话相关信息
// 返回 AllSessionInfo 以及每个子查询的独立错误状态
func GetSessionDetails(db *sql.DB) (allInfo *AllSessionInfo, overviewErr error, eventErr error, historyErr error) {
	logger.Info("开始获取会话模块信息...")
	allInfo = &AllSessionInfo{}

	allInfo.Overview, overviewErr = getCurrentSessionOverview(db)
	if overviewErr != nil {
		logger.Warnf("获取会话概述时出错: %v", overviewErr)
		// overviewErr is returned directly
	}

	allInfo.ByEvent, eventErr = getSessionCountByEvent(db)
	if eventErr != nil {
		logger.Warnf("获取按事件统计的会话时出错: %v", eventErr)
		// eventErr is returned directly
	}

	allInfo.HistoryForChart, historyErr = getDailySessionHistory(db)
	if historyErr != nil {
		logger.Warnf("获取会话历史图表数据时出错: %v", historyErr)
		// historyErr is returned directly. ASH data is often considered optional.
	}

	logger.Info("会话模块信息获取完成。")
	// The function now returns individual errors. The caller can decide how to handle them.
	// A general error is no longer aggregated here unless a specific design requires it.
	return
}

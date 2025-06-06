// Package db handles database querying functionalities for backup and recovery information.
package db

import (
	"database/sql"
	"fmt"

	"github.com/goodwaysIT/inspect4oracle/internal/logger"
)

// ArchivelogModeInfo 存储数据库的日志模式。
type ArchivelogModeInfo struct {
	LogMode string `json:"log_mode"` // ARCHIVELOG or NOARCHIVELOG
}

// GetArchivelogMode 获取数据库当前的日志模式。
func GetArchivelogMode(db *sql.DB) (ArchivelogModeInfo, error) {
	var info ArchivelogModeInfo
	query := `SELECT LOG_MODE AS LogMode FROM V$DATABASE`
	err := db.QueryRow(query).Scan(&info.LogMode)
	if err != nil {
		return info, fmt.Errorf("获取数据库日志模式失败: %w", err)
	}
	logger.Infof("成功获取数据库日志模式: %s", info.LogMode)
	return info, nil
}

// RMANBackupJobInfo 存储 RMAN 备份作业的详细信息。
type RMANBackupJobInfo struct {
	SessionKey         int64        `json:"session_key"`
	StartTime          sql.NullTime `json:"start_time"`
	EndTime            sql.NullTime `json:"end_time"`
	InputBytesDisplay  string       `json:"input_bytes_display"`
	OutputBytesDisplay string       `json:"output_bytes_display"`
	Status             string       `json:"status"`
	TimeTakenDisplay   string       `json:"time_taken_display"`
	Optimized          string       `json:"optimized"`         // For V$RMAN_BACKUP_JOB_DETAILS if available
	CompressionRatio   float64      `json:"compression_ratio"` // For V$RMAN_BACKUP_JOB_DETAILS if available
}

// GetRecentRMANBackupJobs 获取最近的 RMAN 备份作业 (例如过去7天)。
func GetRecentRMANBackupJobs(db *sql.DB) ([]RMANBackupJobInfo, error) {
	query := `
SELECT 
    SESSION_KEY AS SessionKey, 
    START_TIME AS StartTime, 
    END_TIME AS EndTime, 
    INPUT_BYTES_DISPLAY AS InputBytesDisplay, 
    OUTPUT_BYTES_DISPLAY AS OutputBytesDisplay, 
    STATUS AS Status, 
    TIME_TAKEN_DISPLAY AS TimeTakenDisplay,
    OPTIMIZED AS Optimized,
    COMPRESSION_RATIO AS CompressionRatio
FROM V$RMAN_BACKUP_JOB_DETAILS 
WHERE START_TIME >= SYSDATE - 7
ORDER BY START_TIME DESC`

	var jobs []RMANBackupJobInfo
	err := ExecuteQueryAndScanToStructs(db, &jobs, query)
	if err != nil {
		// V$RMAN_BACKUP_JOB_DETAILS 可能不存在或无权限，尝试 V$BACKUP_SET 作为备选
		logger.Warnf("查询 V$RMAN_BACKUP_JOB_DETAILS 失败 (%v)，尝试 V$BACKUP_SET", err)
		queryBackupSet := `
SELECT 
    RECID AS SessionKey, 
    START_TIME AS StartTime, 
    COMPLETION_TIME AS EndTime, 
    NULL AS InputBytesDisplay, -- V$BACKUP_SET 没有直接的 INPUT_BYTES_DISPLAY
    TO_CHAR(BYTES) AS OutputBytesDisplay, -- 近似值
    'COMPLETED' AS Status, -- V$BACKUP_SET 通常只记录完成的
    TO_CHAR(ELAPSED_SECONDS) || ' seconds' AS TimeTakenDisplay,
    NULL AS Optimized,
    NULL AS CompressionRatio 
FROM V$BACKUP_SET 
WHERE COMPLETION_TIME >= SYSDATE - 7 AND BACKUP_TYPE != 'L' -- 排除纯归档日志备份，关注数据文件备份
ORDER BY COMPLETION_TIME DESC`
		err = ExecuteQueryAndScanToStructs(db, &jobs, queryBackupSet)
		if err != nil {
			return nil, fmt.Errorf("获取 RMAN 备份作业信息失败 (尝试了 V$RMAN_BACKUP_JOB_DETAILS 和 V$BACKUP_SET): %w", err)
		}
	}
	logger.Infof("成功获取 %d 条 RMAN 备份作业信息。", len(jobs))
	return jobs, nil
}

// FlashbackStatusInfo 存储闪回数据库的状态。
type FlashbackStatusInfo struct {
	FlashbackOn         string        `json:"flashback_on"`
	OldestFlashbackSCN  sql.NullInt64 `json:"oldest_flashback_scn"`
	OldestFlashbackTime sql.NullTime  `json:"oldest_flashback_time"`
	RetentionTarget     sql.NullInt64 `json:"retention_target"` // DB_FLASHBACK_RETENTION_TARGET in minutes
}

// GetFlashbackStatus 获取闪回数据库的状态。
func GetFlashbackStatus(db *sql.DB) (FlashbackStatusInfo, error) {
	var info FlashbackStatusInfo
	query := `
SELECT 
    d.FLASHBACK_ON AS FlashbackOn, 
    l.OLDEST_FLASHBACK_SCN AS OldestFlashbackSCN, 
    l.OLDEST_FLASHBACK_TIME AS OldestFlashbackTime,
    TO_NUMBER(p.VALUE) AS RetentionTarget
FROM V$DATABASE d 
LEFT JOIN V$FLASHBACK_DATABASE_LOG l ON 1=1
LEFT JOIN V$PARAMETER p ON p.NAME = 'db_flashback_retention_target'`

	err := db.QueryRow(query).Scan(&info.FlashbackOn, &info.OldestFlashbackSCN, &info.OldestFlashbackTime, &info.RetentionTarget)
	if err != nil && err != sql.ErrNoRows {
		return info, fmt.Errorf("获取闪回数据库状态失败: %w", err)
	}
	logger.Infof("成功获取闪回数据库状态: FlashbackOn=%s", info.FlashbackOn)
	return info, nil
}

// RecycleBinObjectInfo 存储回收站中的对象信息。
type RecycleBinObjectInfo struct {
	Owner        string         `json:"owner"`
	ObjectName   string         `json:"object_name"`
	OriginalName string         `json:"original_name"`
	Type         string         `json:"type"`
	TsName       string         `json:"ts_name"`
	Createtime   sql.NullString `json:"createtime"`
	Droptime     sql.NullString `json:"droptime"`
	Space        int64          `json:"space"` // Blocks
	CanUndrop    string         `json:"can_undrop"`
}

// GetRecycleBinObjects 获取回收站中的对象 (仅可恢复的)。
func GetRecycleBinObjects(db *sql.DB) ([]RecycleBinObjectInfo, error) {
	query := `
SELECT 
    OWNER AS Owner, 
    OBJECT_NAME AS ObjectName, 
    ORIGINAL_NAME AS OriginalName, 
    TYPE AS Type, 
    TS_NAME AS TsName, 
    CREATETIME AS Createtime, 
    DROPTIME AS Droptime, 
    SPACE AS Space, 
    CAN_UNDROP AS CanUndrop
FROM DBA_RECYCLEBIN 
WHERE CAN_UNDROP = 'YES' AND TYPE != 'INDEX' -- 排除索引，因为恢复表时会自动重建
ORDER BY DROPTIME DESC`

	var objects []RecycleBinObjectInfo
	err := ExecuteQueryAndScanToStructs(db, &objects, query)
	if err != nil {
		return nil, fmt.Errorf("获取回收站对象信息失败: %w", err)
	}
	logger.Infof("成功获取 %d 条回收站对象信息。", len(objects))
	return objects, nil
}

// DataPumpJobInfo 存储 Data Pump 作业的信息。
type DataPumpJobInfo struct {
	JobName          string        `json:"job_name"`
	OwnerName        string        `json:"owner_name"`
	Operation        string        `json:"operation"`
	JobMode          string        `json:"job_mode"`
	State            string        `json:"state"`
	AttachedSessions sql.NullInt64 `json:"attached_sessions"`
}

// GetDataPumpJobs 获取当前或最近的 Data Pump 作业。
func GetDataPumpJobs(db *sql.DB) ([]DataPumpJobInfo, error) {
	query := `
SELECT 
    JOB_NAME AS JobName, 
    OWNER_NAME AS OwnerName, 
    OPERATION AS Operation, 
    JOB_MODE AS JobMode, 
    STATE AS State, 
    ATTACHED_SESSIONS AS AttachedSessions
FROM DBA_DATAPUMP_JOBS 
ORDER BY OWNER_NAME, JOB_NAME`

	var jobs []DataPumpJobInfo
	err := ExecuteQueryAndScanToStructs(db, &jobs, query)
	if err != nil {
		return nil, fmt.Errorf("获取 Data Pump 作业信息失败: %w", err)
	}
	logger.Infof("成功获取 %d 条 Data Pump 作业信息。", len(jobs))
	return jobs, nil
}

// AllBackupInfo 包含所有备份相关的信息的集合，用于传递给处理器。
type AllBackupInfo struct {
	ArchivelogMode       ArchivelogModeInfo
	RMANJobs             []RMANBackupJobInfo
	FlashbackStatus      FlashbackStatusInfo
	RecycleBinItems      []RecycleBinObjectInfo
	DataPumpJobs         []DataPumpJobInfo
	ArchivelogModeError  error
	RMANJobsError        error
	FlashbackStatusError error
	RecycleBinError      error
	DataPumpJobsError    error
}

// GetAllBackupDetails 汇总所有备份相关的信息。
func GetAllBackupDetails(db *sql.DB) AllBackupInfo {
	var backupInfo AllBackupInfo

	backupInfo.ArchivelogMode, backupInfo.ArchivelogModeError = GetArchivelogMode(db)
	backupInfo.RMANJobs, backupInfo.RMANJobsError = GetRecentRMANBackupJobs(db)
	backupInfo.FlashbackStatus, backupInfo.FlashbackStatusError = GetFlashbackStatus(db)
	backupInfo.RecycleBinItems, backupInfo.RecycleBinError = GetRecycleBinObjects(db)
	backupInfo.DataPumpJobs, backupInfo.DataPumpJobsError = GetDataPumpJobs(db)

	return backupInfo
}

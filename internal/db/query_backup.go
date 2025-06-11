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

// GetArchivelogMode gets the current log mode of the database.
func GetArchivelogMode(db *sql.DB) (ArchivelogModeInfo, error) {
	var info ArchivelogModeInfo
	query := `SELECT LOG_MODE AS LogMode FROM V$DATABASE`
	err := db.QueryRow(query).Scan(&info.LogMode)
	if err != nil {
		return info, fmt.Errorf("failed to get database log mode: %w", err)
	}
	logger.Infof("Successfully retrieved database log mode: %s", info.LogMode)
	return info, nil
}

// RMANBackupJobInfo stores detailed information about RMAN backup jobs.
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

// GetRecentRMANBackupJobs gets recent RMAN backup jobs (e.g., last 7 days).
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
		logger.Warnf("Failed to query V$RMAN_BACKUP_JOB_DETAILS (%v), trying V$BACKUP_SET", err)
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
WHERE COMPLETION_TIME >= SYSDATE - 7 AND BACKUP_TYPE != 'L' -- Exclude pure archive log backups, focus on data file backups
ORDER BY COMPLETION_TIME DESC`
		err = ExecuteQueryAndScanToStructs(db, &jobs, queryBackupSet)
		if err != nil {
			return nil, fmt.Errorf("failed to get RMAN backup job information (tried V$RMAN_BACKUP_JOB_DETAILS and V$BACKUP_SET): %w", err)
		}
	}
	logger.Infof("Successfully retrieved %d RMAN backup job records.", len(jobs))
	return jobs, nil
}

// FlashbackStatusInfo 存储闪回数据库的状态。
type FlashbackStatusInfo struct {
	FlashbackOn         string        `json:"flashback_on"`
	OldestFlashbackSCN  sql.NullInt64 `json:"oldest_flashback_scn"`
	OldestFlashbackTime sql.NullTime  `json:"oldest_flashback_time"`
	RetentionTarget     sql.NullInt64 `json:"retention_target"` // DB_FLASHBACK_RETENTION_TARGET in minutes
}

// GetFlashbackStatus gets the status of the flashback database.
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
		return info, fmt.Errorf("failed to get flashback database status: %w", err)
	}
	logger.Infof("Successfully retrieved flashback database status: FlashbackOn=%s", info.FlashbackOn)
	return info, nil
}

// RecycleBinObjectInfo stores information about objects in the recycle bin.
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

// GetRecycleBinObjects gets objects from the recycle bin (only recoverable ones).
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
		return nil, fmt.Errorf("failed to get recycle bin object information: %w", err)
	}
	logger.Infof("Successfully retrieved %d recycle bin object records.", len(objects))
	return objects, nil
}

// DataPumpJobInfo stores information about Data Pump jobs.
type DataPumpJobInfo struct {
	JobName          string        `json:"job_name"`
	OwnerName        string        `json:"owner_name"`
	Operation        string        `json:"operation"`
	JobMode          string        `json:"job_mode"`
	State            string        `json:"state"`
	AttachedSessions sql.NullInt64 `json:"attached_sessions"`
}

// GetDataPumpJobs gets current or recent Data Pump jobs.
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
		return nil, fmt.Errorf("failed to get Data Pump job information: %w", err)
	}
	logger.Infof("Successfully retrieved %d Data Pump job records.", len(jobs))
	return jobs, nil
}

// AllBackupInfo contains a collection of all backup-related information to be passed to the processor.
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

// GetAllBackupDetails aggregates all backup-related information.
func GetAllBackupDetails(db *sql.DB) AllBackupInfo {
	var backupInfo AllBackupInfo

	backupInfo.ArchivelogMode, backupInfo.ArchivelogModeError = GetArchivelogMode(db)
	backupInfo.RMANJobs, backupInfo.RMANJobsError = GetRecentRMANBackupJobs(db)
	backupInfo.FlashbackStatus, backupInfo.FlashbackStatusError = GetFlashbackStatus(db)
	backupInfo.RecycleBinItems, backupInfo.RecycleBinError = GetRecycleBinObjects(db)
	backupInfo.DataPumpJobs, backupInfo.DataPumpJobsError = GetDataPumpJobs(db)

	return backupInfo
}

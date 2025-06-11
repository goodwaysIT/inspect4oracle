package db

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/goodwaysIT/inspect4oracle/internal/logger"
)

// StorageInfo contains all storage-related information
type StorageInfo struct {
	ControlFiles        []ControlFileInfo
	RedoLogs            []RedoLogInfo
	DataFiles           []DataFileInfo
	Tablespaces         []TablespaceInfo
	ArchivedLogsSummary []ArchivedLogSummary // Daily archive volume for the past 7 days
	ASMDiskgroups       []ASMDiskgroupInfo   // If ASM is used
}

// ControlFileInfo stores information about control files
type ControlFileInfo struct {
	Name   string  `db:"NAME"` // Assuming column name is NAME
	SizeMB float64 `db:"SIZE_MB"`
}

// RedoLogInfo stores information about Redo log groups and members
type RedoLogInfo struct {
	GroupNo    int     `db:"GROUP_NO"`
	ThreadNo   int     `db:"THREAD_NO"`
	SequenceNo int64   `db:"SEQUENCE_NO"` // Assuming column name is SEQUENCE_NO
	Members    int     `db:"MEMBERS"`     // Assuming column name is MEMBERS
	SizeMB     float64 `db:"SIZE_MB"`
	Member     string  `db:"MEMBER"`   // Assuming column name is MEMBER
	Status     string  `db:"STATUS"`   // Assuming column name is STATUS
	Archived   string  `db:"ARCHIVED"` // Assuming column name is ARCHIVED
	Type       string  `db:"TYPE"`     // Assuming column name is TYPE
}

// DataFileInfo stores information about data files
type DataFileInfo struct {
	FileID         int     `db:"FILE_ID"`
	FileName       string  `db:"FILE_NAME"`
	TablespaceName string  `db:"TABLESPACE_NAME"`
	SizeMB         float64 `db:"SIZE_MB"`
	Status         string  `db:"STATUS"`         // Assuming column name is STATUS
	Autoextensible string  `db:"AUTOEXTENSIBLE"` // Assuming column name is AUTOEXTENSIBLE
}

// TablespaceInfo stores tablespace usage information
type TablespaceInfo struct {
	Status                 string  `db:"STATUS"`
	Name                   string  `db:"TABLESPACE_NAME"`
	Type                   string  `db:"CONTENTS"` // PERMANENT, TEMPORARY, UNDO
	ExtentManagement       string  `db:"EXTENT_MANAGEMENT"`
	SegmentSpaceManagement string  `db:"SEGMENT_SPACE_MANAGEMENT"`
	UsedMB                 float64 `db:"USED_MB"`
	TotalMB                float64 `db:"CURRENT_SIZE_MB"` // Mapped to current_size_MB from query
	UsedPercent            string  `db:"PCT_USED"`
	CanExtendMB            float64 `db:"CANEXTEND_SIZE_MB"`
}

// ArchivedLogSummary stores summary information for daily archived logs
type ArchivedLogSummary struct {
	Day         string // YYYY-MM-DD
	LogCount    int
	TotalSizeMB float64
}

// ASMDiskgroupInfo stores information about ASM diskgroups
type ASMDiskgroupInfo struct {
	Name           string
	TotalMB        int64
	FreeMB         int64
	UsedPercent    float64
	State          string
	RedundancyType string
}

// GetStorageInfo fetches all storage-related information
// dbVersion is used to determine if certain queries are applicable (e.g., ASM-related views are widely used after specific versions)
// logMode is used to determine whether to query archive logs (e.g., archive logs are irrelevant in NOARCHIVELOG mode)
func GetStorageInfo(db *sql.DB) (*StorageInfo, error) {
	logger.Info("Starting to fetch storage information...")
	startTime := time.Now()
	storageInfo := &StorageInfo{}
	var err error

	// 1. Control Files
	storageInfo.ControlFiles, err = getControlFiles(db)
	if err != nil {
		logger.Warnf("Failed to get control file info: %v. Continuing with other storage items.", err)
		// Do not interrupt, log the error and continue
	}

	// 2. Redo Logs
	storageInfo.RedoLogs, err = getRedoLogs(db)
	if err != nil {
		logger.Warnf("Failed to get redo log info: %v. Continuing with other storage items.", err)
	}

	// 3. Data Files
	storageInfo.DataFiles, err = getDataFiles(db)
	if err != nil {
		logger.Warnf("Failed to get data file info: %v. Continuing with other storage items.", err)
	}

	// 4. Tablespace Usage
	storageInfo.Tablespaces, err = getTablespaceUsage(db)
	if err != nil {
		logger.Warnf("Failed to get tablespace usage: %v. Continuing with other storage items.", err)
	}

	// 5. Archived Log Summary (only meaningful in ARCHIVELOG mode)

	storageInfo.ArchivedLogsSummary, err = getArchivedLogSummary(db)
	if err != nil {
		logger.Warnf("Failed to get archived log summary: %v. Continuing with other storage items.", err)
	}

	// 6. ASM Diskgroups (usually require specific permissions and only when using ASM)
	// Simple version check, in practice more complex logic may be needed to determine the ASM environment
	// For example, check the 'asm instance' parameter or try to query v$asm_diskgroup, and skip if it fails
	isASM, errAsmCheck := checkASMInstance(db)
	if errAsmCheck != nil {
		logger.Warnf("Failed to check ASM environment: %v. Skipping ASM diskgroup query.", errAsmCheck)
	} else if isASM {
		storageInfo.ASMDiskgroups, err = getASMDiskgroupInfo(db)
		if err != nil {
			logger.Warnf("Failed to get ASM diskgroup info: %v. Continuing with other storage items.", err)
		}
	} else {
		logger.Info("ASM environment not detected or cannot be confirmed, skipping ASM diskgroup query.")
	}

	logger.Infof("Finished fetching storage information, elapsed time: %s", time.Since(startTime))
	return storageInfo, nil // Return the collected information, even if some queries fail
}

func getControlFiles(db *sql.DB) ([]ControlFileInfo, error) {
	var files []ControlFileInfo
	query := "SELECT NAME, round(BLOCK_SIZE*FILE_SIZE_BLKS/1024/1024) AS SIZE_MB FROM V$CONTROLFILE"
	err := ExecuteQueryAndScanToStructs(db, &files, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get control file info (generic scan): %w", err)
	}
	logger.Infof("Successfully fetched info for %d control files.", len(files))
	logger.Debugf("Control file info: %v", files)
	return files, nil
}

func getRedoLogs(db *sql.DB) ([]RedoLogInfo, error) {
	var logs []RedoLogInfo
	// Aliased g.GROUP# to GROUP_NO and g.THREAD# to THREAD_NO for struct field matching.
	// SequenceNo is not in the original query, so it won't be populated.
	query := `
SELECT
    g.GROUP# AS GROUP_NO,
    g.THREAD# AS THREAD_NO,
    g.MEMBERS,
    l.MEMBER,
    round(g.BYTES / 1024 / 1024) AS SIZE_MB,
    g.STATUS,
    l.TYPE,
    g.ARCHIVED
FROM V$LOG g JOIN V$LOGFILE l ON g.GROUP# = l.GROUP#
ORDER BY g.GROUP#, l.MEMBER`

	err := ExecuteQueryAndScanToStructs(db, &logs, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get redo log info (generic scan): %w", err)
	}
	logger.Infof("Successfully fetched info for %d redo logs.", len(logs))
	logger.Debugf("Redo log info: %v", logs)
	return logs, nil
}

func getDataFiles(db *sql.DB) ([]DataFileInfo, error) {
	query := `
SELECT
    df.TABLESPACE_NAME AS TABLESPACE_NAME,
	df.file_id AS FILE_ID,
    df.FILE_NAME AS FILE_NAME,
    df.BYTES / 1024 / 1024 AS SIZE_MB,
    df.STATUS AS STATUS,
    df.AUTOEXTENSIBLE AS AUTOEXTENSIBLE
FROM DBA_DATA_FILES df
UNION ALL
SELECT
    tf.TABLESPACE_NAME AS TABLESPACE_NAME,
	tf.file_id AS FILE_ID,
    tf.FILE_NAME AS FILE_NAME,
    tf.BYTES / 1024 / 1024 AS SIZE_MB,
    tf.STATUS AS STATUS,
    tf.AUTOEXTENSIBLE AS AUTOEXTENSIBLE
FROM DBA_TEMP_FILES tf
ORDER BY TABLESPACE_NAME, FILE_ID`

	var files []DataFileInfo
	err := ExecuteQueryAndScanToStructs(db, &files, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get data file info (generic scan): %w", err)
	}
	logger.Infof("Successfully fetched info for %d data files.", len(files))
	logger.Debugf("Data file info: %v", files)
	return files, nil
}

func getTablespaceUsage(db *sql.DB) ([]TablespaceInfo, error) {
	// 查询普通表空间（永久、UNDO）
	// Aliased columns for clarity and direct mapping to struct fields
	queryPermanentUndo := `
SELECT
            d.status
                , d.tablespace_name
                , d.contents
                , d.extent_management
                , d.segment_space_management
                , NVL(b.allocatesize - NVL(f.freesize, 0), 0)   used_MB
                , b.allocatesize current_size_MB
                , to_char(NVL((b.allocatesize - NVL(f.freesize, 0)) / b.allocatesize * 100, 0),'990.99')||'%' pct_used
                , a.maxsize canextend_size_MB
                --, to_char(NVL((b.allocatesize - NVL(f.freesize, 0)) / a.maxsize * 100, 0),'990.99')||'%' tot_pct_used
        FROM dba_tablespaces d
                , (     SELECT tablespace_name,sum(maxsize) maxsize
                        FROM (  SELECT tablespace_name, decode(autoextensible,'YES',round(sum(maxbytes)/1024/1024),round(sum(bytes)/1024/1024)) maxsize
                                        FROM dba_data_files
                                        GROUP BY tablespace_name,autoextensible
                                ) GROUP BY tablespace_name
                  ) a
                , ( SELECT tablespace_name, sum(bytes)/1024/1024 allocatesize
              from dba_data_files
              group by tablespace_name
                  ) b
                , (     SELECT tablespace_name, sum(bytes)/1024/1024 freesize
                        FROM dba_free_space
                        GROUP BY tablespace_name
                  ) f
        WHERE d.tablespace_name = a.tablespace_name(+)
        AND d.tablespace_name = b.tablespace_name(+)
        AND d.tablespace_name = f.tablespace_name(+)
        AND d.contents in ('PERMANENT','UNDO')`

	// 查询临时表空间
	// Aliased columns for clarity and direct mapping to struct fields
	queryTemporary := `
SELECT
            d.status
                , d.tablespace_name
                , d.contents
                , d.extent_management
                , d.segment_space_management
                , NVL(b.allocatesize - NVL(f.usedsize, 0), 0)   used_MB
                , b.allocatesize current_size_MB
                , to_char(NVL(NVL(f.usedsize, 0) / b.allocatesize * 100, 0),'990.99')||'%' pct_used
                , a.maxsize canextend_size_MB
                --, to_char(NVL(f.usedsize,0) / a.maxsize * 100,'990.99')||'%' tot_pct_used
        FROM
            sys.dba_tablespaces d
                , (     SELECT tablespace_name,sum(maxsize) maxsize
                        FROM (  SELECT tablespace_name, decode(autoextensible,'YES',round(sum(maxbytes)/1024/1024),round(sum(bytes)/1024/1024)) maxsize
                                        FROM dba_temp_files
                                        GROUP BY tablespace_name,autoextensible
                                ) GROUP BY tablespace_name
                  ) a
          , ( select tablespace_name, sum(bytes)/1024/1024  allocatesize
              from dba_temp_files
              group by tablespace_name
            ) b
          , ( select tablespace_name, sum(bytes_cached)/1024/1024 usedsize
              from v$temp_extent_pool
              group by tablespace_name
            ) f
        WHERE d.tablespace_name = a.tablespace_name(+)
          AND d.tablespace_name = b.tablespace_name(+)
          AND d.tablespace_name = f.tablespace_name(+)
          AND d.extent_management like 'LOCAL'
          AND d.contents like 'TEMPORARY'`

	var tablespaces []TablespaceInfo
	var permUndoTablespaces []TablespaceInfo
	var tempTablespaces []TablespaceInfo
	var errPU, errTemp error

	// 执行永久和UNDO表空间查询
	errPU = ExecuteQueryAndScanToStructs(db, &permUndoTablespaces, queryPermanentUndo)
	if errPU != nil {
		logger.Errorf("Failed to query permanent/UNDO tablespace usage (generic scan): %v", errPU)
		// Do not return an error immediately, try to query the temporary tablespace
	} else {
		tablespaces = append(tablespaces, permUndoTablespaces...)
	}

	// 执行临时表空间查询
	errTemp = ExecuteQueryAndScanToStructs(db, &tempTablespaces, queryTemporary)
	if errTemp != nil {
		logger.Errorf("Failed to query temporary tablespace usage (generic scan): %v", errTemp)
	} else {
		tablespaces = append(tablespaces, tempTablespaces...)
	}

	if errPU != nil && errTemp != nil { // If both queries fail
		return nil, fmt.Errorf("failed to get any tablespace information. Permanent/UNDO error: %v; Temp error: %v", errPU, errTemp)
	}

	logger.Infof("Successfully fetched usage info for %d tablespaces.", len(tablespaces))
	logger.Debugf("Tablespace info: %v", tablespaces)
	return tablespaces, nil
}

func getArchivedLogSummary(db *sql.DB) ([]ArchivedLogSummary, error) {
	// Query the number and size of archived logs per day for the past 7 days.
	// COMPLETION_TIME is the time when archiving was completed.
	query := `
SELECT
    TO_CHAR(TRUNC(COMPLETION_TIME), 'YYYY-MM-DD') AS Day,
    COUNT(*) AS LogCount,
    SUM(BLOCKS * BLOCK_SIZE) / 1024 / 1024 AS TotalSizeMB
FROM V$ARCHIVED_LOG
WHERE COMPLETION_TIME >= TRUNC(SYSDATE) - 7 AND COMPLETION_TIME < TRUNC(SYSDATE) + 1
GROUP BY TRUNC(COMPLETION_TIME)
ORDER BY Day DESC`

	var summaries []ArchivedLogSummary
	err := ExecuteQueryAndScanToStructs(db, &summaries, query)
	if err != nil {
		// It could be a view permission issue, or the view is empty in non-archivelog mode but the query itself does not report an error
		return nil, fmt.Errorf("failed to get archived log summary (generic scan): %w. Please check permissions or confirm the database is in ARCHIVELOG mode.", err)
	}

	logger.Infof("Successfully fetched archived log summary for %d days.", len(summaries))
	logger.Debugf("Archived log summary: %v", summaries) // Changed Debug to Debugf
	return summaries, nil
}

func checkASMInstance(db *sql.DB) (bool, error) {
	var result string
	// Try to query a parameter or view that typically has a specific value only on an ASM instance.
	// For example, one could check instance_type, or directly try to query v$asm_diskgroup and catch the error
	// A simple method is used here: check if the V$ASM_DISKGROUP view exists (by attempting to query it).
	// Note: Even if it's not an ASM instance, querying a non-existent table/view will also result in an error, but the error type may differ.
	// A more reliable method is to query if GV$INSTANCE's INSTANCE_ROLE contains 'ASM STORAGE INSTANCE'.
	// 或者查询参数 CLUSTER_INTERCONNECTS (如果RAC+ASM)
	// Alternatively, try to query V$ASM_DISKGROUP directly. If it reports "ORA-00942: table or view does not exist", it is considered not to be ASM or to have no permissions
	// If it is another error, it may be a connection issue, etc.

	// Simplified: Try to query V$ASM_DISKGROUP, if successful (even if 0 rows are returned), it is considered an ASM environment or has access rights.
	// If ORA-00942 is reported, it is considered not to be ASM or to have no permissions.
	err := db.QueryRow("SELECT COUNT(*) FROM V$ASM_DISKGROUP").Scan(&result)
	if err != nil {
		if strings.Contains(err.Error(), "ORA-00942") { // ORA-00942: table or view does not exist
			logger.Info("V$ASM_DISKGROUP view does not exist or no access permission, assuming non-ASM environment or unable to query ASM information.")
			return false, nil // Not an error, just stating the situation
		}
		// Other errors, may be connection issues, etc.
		return false, fmt.Errorf("failed to access V$ASM_DISKGROUP: %w", err)
	}
	// If the query is successful, even if it returns 0 rows, the view is considered to exist, possibly an ASM environment
	logger.Info("Successfully accessed V$ASM_DISKGROUP, it might be an ASM environment.")
	return true, nil
}

func getASMDiskgroupInfo(db *sql.DB) ([]ASMDiskgroupInfo, error) {
	// 注意：查询V$ASM_DISKGROUP通常需要连接到ASM实例，或者通过DB link从数据库实例访问。
	// The implementation here assumes that the DB connection can already access the V$ASM_DISKGROUP view.
	// Aliased columns for direct mapping and clarity.
	query := `
SELECT
    name AS Name,
    total_mb AS TotalMB,
    free_mb AS FreeMB,
    ROUND((1 - COALESCE(free_mb,0) / DECODE(COALESCE(total_mb,0), 0, 1, COALESCE(total_mb,0))) * 100, 2) AS UsedPercent, -- Handle potential division by zero or NULLs
    state AS State,
    type AS RedundancyType
FROM V$ASM_DISKGROUP
ORDER BY name`

	var diskgroups []ASMDiskgroupInfo
	err := ExecuteQueryAndScanToStructs(db, &diskgroups, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get ASM diskgroup info (generic scan): %w. Please confirm the connected user has permission to access this view and the database environment is configured correctly.", err)
	}

	if len(diskgroups) == 0 {
		logger.Info("No ASM diskgroup information was obtained from V$ASM_DISKGROUP. ASM might not be in use, there might be no data, or permissions might be lacking.")
	}

	logger.Infof("Successfully fetched info for %d ASM diskgroups.", len(diskgroups))
	logger.Debugf("ASM diskgroup info: %v", diskgroups)
	return diskgroups, nil
}

// Helper to convert string to float64, returns 0 on error
func parseFloat(s string) float64 {
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return v
}

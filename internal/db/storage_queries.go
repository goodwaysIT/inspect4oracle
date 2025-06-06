package db

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/goodwaysIT/inspect4oracle/internal/logger"
)

// StorageInfo 包含所有存储相关的信息
type StorageInfo struct {
	ControlFiles        []ControlFileInfo
	RedoLogs            []RedoLogInfo
	DataFiles           []DataFileInfo
	Tablespaces         []TablespaceInfo
	ArchivedLogsSummary []ArchivedLogSummary // 过去7天每天的归档量
	ASMDiskgroups       []ASMDiskgroupInfo   // 如果使用ASM
}

// ControlFileInfo 存储控制文件的信息
type ControlFileInfo struct {
	Name   string  `db:"NAME"` // Assuming column name is NAME
	SizeMB float64 `db:"SIZE_MB"`
}

// RedoLogInfo 存储 Redo 日志组和成员的信息
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

// DataFileInfo 存储数据文件的信息
type DataFileInfo struct {
	FileID         int     `db:"FILE_ID"`
	FileName       string  `db:"FILE_NAME"`
	TablespaceName string  `db:"TABLESPACE_NAME"`
	SizeMB         float64 `db:"SIZE_MB"`
	Status         string  `db:"STATUS"`         // Assuming column name is STATUS
	Autoextensible string  `db:"AUTOEXTENSIBLE"` // Assuming column name is AUTOEXTENSIBLE
}

// TablespaceInfo 存储表空间的使用情况
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

// ArchivedLogSummary 存储每日归档日志的摘要信息
type ArchivedLogSummary struct {
	Day         string // YYYY-MM-DD
	LogCount    int
	TotalSizeMB float64
}

// ASMDiskgroupInfo 存储 ASM 磁盘组的信息
type ASMDiskgroupInfo struct {
	Name           string
	TotalMB        int64
	FreeMB         int64
	UsedPercent    float64
	State          string
	RedundancyType string
}

// GetStorageInfo 获取所有存储相关信息
// dbVersion 用于判断某些查询是否适用 (例如 ASM 相关视图在特定版本后才广泛使用)
// logMode 用于判断是否查询归档日志 (例如 NOARCHIVELOG 模式下归档日志不相关)
func GetStorageInfo(db *sql.DB) (*StorageInfo, error) {
	logger.Info("开始获取存储信息...")
	startTime := time.Now()
	storageInfo := &StorageInfo{}
	var err error

	// 1. 控制文件
	storageInfo.ControlFiles, err = getControlFiles(db)
	if err != nil {
		logger.Warnf("获取控制文件信息失败: %v. 继续处理其他存储项.", err)
		// 不中断，记录错误并继续
	}

	// 2. Redo 日志
	storageInfo.RedoLogs, err = getRedoLogs(db)
	if err != nil {
		logger.Warnf("获取Redo日志信息失败: %v. 继续处理其他存储项.", err)
	}

	// 3. 数据文件
	storageInfo.DataFiles, err = getDataFiles(db)
	if err != nil {
		logger.Warnf("获取数据文件信息失败: %v. 继续处理其他存储项.", err)
	}

	// 4. 表空间使用情况
	storageInfo.Tablespaces, err = getTablespaceUsage(db)
	if err != nil {
		logger.Warnf("获取表空间使用情况失败: %v. 继续处理其他存储项.", err)
	}

	// 5. 归档日志摘要 (仅在 ARCHIVELOG 模式下有意义)

	storageInfo.ArchivedLogsSummary, err = getArchivedLogSummary(db)
	if err != nil {
		logger.Warnf("获取归档日志摘要失败: %v. 继续处理其他存储项.", err)
	}

	// 6. ASM 磁盘组 (通常需要特定权限，且仅当使用ASM时)
	// 简单检查版本，实际可能需要更复杂的逻辑判断是否为ASM环境
	// 例如，检查 'asm instance' 参数或尝试查询 v$asm_diskgroup，如果失败则跳过
	isASM, errAsmCheck := checkASMInstance(db)
	if errAsmCheck != nil {
		logger.Warnf("检查ASM环境失败: %v. 跳过ASM磁盘组查询.", errAsmCheck)
	} else if isASM {
		storageInfo.ASMDiskgroups, err = getASMDiskgroupInfo(db)
		if err != nil {
			logger.Warnf("获取ASM磁盘组信息失败: %v. 继续处理其他存储项.", err)
		}
	} else {
		logger.Info("未检测到或无法确认ASM环境，跳过ASM磁盘组查询。")
	}

	logger.Infof("获取存储信息完成, 耗时: %s", time.Since(startTime))
	return storageInfo, nil // 返回收集到的信息，即使部分查询失败
}

func getControlFiles(db *sql.DB) ([]ControlFileInfo, error) {
	var files []ControlFileInfo
	query := "SELECT NAME, round(BLOCK_SIZE*FILE_SIZE_BLKS/1024/1024) AS SIZE_MB FROM V$CONTROLFILE"
	err := ExecuteQueryAndScanToStructs(db, &files, query)
	if err != nil {
		return nil, fmt.Errorf("获取控制文件信息失败 (generic scan): %w", err)
	}
	logger.Infof("成功获取 %d 个控制文件信息。", len(files))
	logger.Debugf("控制文件信息: %v", files)
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
		return nil, fmt.Errorf("获取Redo日志信息失败 (generic scan): %w", err)
	}
	logger.Infof("成功获取 %d 个Redo日志信息。", len(logs))
	logger.Debugf("Redo日志信息: %v", logs)
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
		return nil, fmt.Errorf("获取数据文件信息失败 (generic scan): %w", err)
	}
	logger.Infof("成功获取 %d 个数据文件信息。", len(files))
	logger.Debugf("数据文件信息: %v", files)
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
		logger.Errorf("查询永久/UNDO表空间使用情况失败 (generic scan): %v", errPU)
		// 不立即返回错误，尝试查询临时表空间
	} else {
		tablespaces = append(tablespaces, permUndoTablespaces...)
	}

	// 执行临时表空间查询
	errTemp = ExecuteQueryAndScanToStructs(db, &tempTablespaces, queryTemporary)
	if errTemp != nil {
		logger.Errorf("查询临时表空间使用情况失败 (generic scan): %v", errTemp)
	} else {
		tablespaces = append(tablespaces, tempTablespaces...)
	}

	if errPU != nil && errTemp != nil { // 如果两个查询都失败了
		return nil, fmt.Errorf("获取所有表空间信息均失败。永久/UNDO错误: %v; 临时错误: %v", errPU, errTemp)
	}

	logger.Infof("成功获取 %d 个表空间的使用情况信息。", len(tablespaces))
	logger.Debugf("表空间信息: %v", tablespaces)
	return tablespaces, nil
}

func getArchivedLogSummary(db *sql.DB) ([]ArchivedLogSummary, error) {
	// 查询过去7天每天的归档日志数量和大小
	// COMPLETION_TIME 是归档完成的时间。
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
		// 可能是视图权限问题，或者非归档模式下视图为空但查询本身不报错
		return nil, fmt.Errorf("获取归档日志摘要失败 (generic scan): %w。请检查权限或确认数据库是否处于ARCHIVELOG模式。", err)
	}

	logger.Infof("成功获取 %d 天的归档日志摘要。", len(summaries))
	logger.Debugf("归档日志摘要: %v", summaries) // Changed Debug to Debugf
	return summaries, nil
}

func checkASMInstance(db *sql.DB) (bool, error) {
	var result string
	// 尝试查询一个只在ASM实例上通常有特定值的参数或视图
	// 例如，可以检查 instance_type，或者直接尝试查询 v$asm_diskgroup 并捕获错误
	// 这里用一个简单的方法：检查是否存在 V$ASM_DISKGROUP 视图（通过尝试查询）
	// 注意：即使不是ASM实例，查询一个不存在的表/视图也会报错，但错误类型可能不同。
	// 一个更可靠的方法是查询 GV$INSTANCE 的 INSTANCE_ROLE 是否包含 'ASM STORAGE INSTANCE'
	// 或者查询参数 CLUSTER_INTERCONNECTS (如果RAC+ASM)
	// 或者直接尝试查询 V$ASM_DISKGROUP，如果报错 "ORA-00942: table or view does not exist"，则认为不是ASM或无权限
	// 如果是其他错误，则可能是连接问题等。

	// 简化：尝试查询 V$ASM_DISKGROUP，如果成功（即使返回0行），则认为是ASM环境或有权限访问。
	// 如果报错 ORA-00942，则认为不是ASM或无权限。
	err := db.QueryRow("SELECT COUNT(*) FROM V$ASM_DISKGROUP").Scan(&result)
	if err != nil {
		if strings.Contains(err.Error(), "ORA-00942") { // ORA-00942: table or view does not exist
			logger.Info("V$ASM_DISKGROUP 视图不存在或无权限访问，假定非ASM环境或无法查询ASM信息。")
			return false, nil // 不是错误，只是说明情况
		}
		// 其他错误，可能是连接问题等
		return false, fmt.Errorf("尝试访问 V$ASM_DISKGROUP 失败: %w", err)
	}
	// 如果查询成功，即使返回0行，也认为视图存在，可能是ASM环境
	logger.Info("成功访问 V$ASM_DISKGROUP，可能为ASM环境。")
	return true, nil
}

func getASMDiskgroupInfo(db *sql.DB) ([]ASMDiskgroupInfo, error) {
	// 注意：查询V$ASM_DISKGROUP通常需要连接到ASM实例，或者通过DB link从数据库实例访问。
	// 这里的实现假设DB连接已经能够访问到V$ASM_DISKGROUP视图。
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
		return nil, fmt.Errorf("获取ASM磁盘组信息失败 (generic scan): %w. 请确认连接用户有权限访问此视图，并且数据库环境配置正确。", err)
	}

	if len(diskgroups) == 0 {
		logger.Info("未从V$ASM_DISKGROUP获取到ASM磁盘组信息。可能未使用ASM，或无数据，或无权限。")
	}

	logger.Infof("成功获取 %d 个ASM磁盘组信息。", len(diskgroups))
	logger.Debugf("ASM磁盘组信息: %v", diskgroups)
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

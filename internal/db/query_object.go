package db

import (
	"database/sql"
	"fmt"

	"github.com/goodwaysIT/inspect4oracle/internal/logger"
)

// ObjectOverview contains the count of objects grouped by owner and object type.
// SQL: select owner,object_type,count(*) from dba_objects where owner not in ('SYS','SYSTEM') group by owner,object_type order by 1,2;
type ObjectOverview struct {
	Owner       string `json:"owner"`
	ObjectType  string `json:"object_type"`
	ObjectCount int    `json:"object_count"`
}

// InvalidObjectInfo contains information about invalid objects.
// SQL: SELECT owner, object_name, object_type, created, last_ddl_time FROM dba_objects WHERE status = 'INVALID' ORDER BY owner, object_type, object_name;
type InvalidObjectInfo struct {
	Owner       string `json:"owner"`
	ObjectName  string `json:"object_name"`
	ObjectType  string `json:"object_type"`
	Created     string `json:"created"`       // Assuming string representation for simplicity
	LastDDLTime string `json:"last_ddl_time"` // Assuming string representation
}

// TopSegment contains information about objects with the largest segment sizes.
// SQL: select * from (select owner,segment_type,segment_name,sum(bytes)/1024/1024 size_mb from dba_segments group by owner,segment_type,segment_name order by size_mb desc) where rownum < 11;
type TopSegment struct {
	Owner       string  `json:"owner"`
	SegmentType string  `json:"segment_type"`
	SegmentName string  `json:"segment_name"`
	SizeMB      float64 `json:"size_mb"`
}

// AllObjectInfo contains information for all object-related modules.
type AllObjectInfo struct {
	Overview       []ObjectOverview
	TopSegments    []TopSegment
	InvalidObjects []InvalidObjectInfo
}

// getObjectOverview gets object overview statistics.
func getObjectOverview(db *sql.DB) ([]ObjectOverview, error) {
	query := `
SELECT 
    owner AS Owner, 
    object_type AS ObjectType, 
    count(*) AS ObjectCount 
FROM dba_objects 
WHERE owner NOT IN ('SYS', 'SYSTEM', 'DBSNMP', 'OUTLN', 'DIP', 'TSMSYS', 'ORACLE_OCM', 'APPQOSSYS', 'GSMADMIN_INTERNAL', 'XDB', 'WMSYS', 'AUDSYS', 'CTXSYS', 'LBACSYS', 'ORDDATA', 'ORDSYS', 'SI_INFORMTN_SCHEMA', 'MDSYS', 'DVSYS', 'EXFSYS', 'OLAPSYS', 'GGSYS', 'ANONYMOUS', 'XS$NULL', 'OJVMSYS', 'DBSFWUSER', 'REMOTE_SCHEDULER_AGENT', 'SYS$UMF', 'SYSBACKUP', 'SYSDG', 'SYSKM', 'SYSRAC') 
  AND owner NOT LIKE 'APEX%' 
  AND owner NOT LIKE 'FLOWS_%' 
  AND owner NOT LIKE 'GG%' 
  AND owner NOT LIKE 'RDSADMIN%' 
GROUP BY owner, object_type 
ORDER BY owner, object_type`
	var overview []ObjectOverview
	err := ExecuteQueryAndScanToStructs(db, &overview, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get object overview: %w", err)
	}
	logger.Infof("Successfully fetched %d object overview entries.", len(overview))
	// logger.Debugf("Object overview info: %+v", overview) // Can be very verbose
	return overview, nil
}

// getInvalidObjects gets information for all invalid objects.
func getInvalidObjects(db *sql.DB) ([]InvalidObjectInfo, error) {
	query := `
SELECT 
    owner AS Owner, 
    object_name AS ObjectName, 
    object_type AS ObjectType,
    TO_CHAR(created, 'YYYY-MM-DD HH24:MI:SS') AS Created,
    TO_CHAR(last_ddl_time, 'YYYY-MM-DD HH24:MI:SS') AS LastDDLTime
FROM dba_objects 
WHERE status = 'INVALID'
  AND owner NOT IN ('SYS', 'SYSTEM', 'DBSNMP', 'OUTLN', 'DIP', 'TSMSYS', 'ORACLE_OCM', 'APPQOSSYS', 'GSMADMIN_INTERNAL', 'XDB', 'WMSYS', 'AUDSYS', 'CTXSYS', 'LBACSYS', 'ORDDATA', 'ORDSYS', 'SI_INFORMTN_SCHEMA', 'MDSYS', 'DVSYS', 'EXFSYS', 'OLAPSYS', 'GGSYS', 'ANONYMOUS', 'XS$NULL', 'OJVMSYS', 'DBSFWUSER', 'REMOTE_SCHEDULER_AGENT', 'SYS$UMF', 'SYSBACKUP', 'SYSDG', 'SYSKM', 'SYSRAC') 
  AND owner NOT LIKE 'APEX%' 
  AND owner NOT LIKE 'FLOWS_%' 
  AND owner NOT LIKE 'GG%' 
  AND owner NOT LIKE 'RDSADMIN%' 
ORDER BY owner, object_type, object_name`
	var invalidObjects []InvalidObjectInfo
	err := ExecuteQueryAndScanToStructs(db, &invalidObjects, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get invalid object info: %w", err)
	}
	logger.Infof("Successfully fetched info for %d invalid objects.", len(invalidObjects))
	return invalidObjects, nil
}

// getTopSegments gets information for the top ten largest segments by size.
func getTopSegments(db *sql.DB) ([]TopSegment, error) {
	// Oracle's ROWNUM is applied *before* ORDER BY in a subquery if not careful.
	// The subquery correctly orders by size_mb DESC, then the outer query limits to ROWNUM < 11.
	query := `
SELECT Owner, SegmentType, SegmentName, SizeMB
FROM (
    SELECT 
        owner AS Owner, 
        segment_type AS SegmentType, 
        segment_name AS SegmentName, 
        SUM(bytes) / 1024 / 1024 AS SizeMB
    FROM dba_segments 
    WHERE owner NOT IN ('SYS', 'SYSTEM', 'DBSNMP', 'OUTLN', 'DIP', 'TSMSYS', 'ORACLE_OCM', 'APPQOSSYS', 
	'GSMADMIN_INTERNAL', 'XDB', 'WMSYS', 'AUDSYS', 'CTXSYS', 'LBACSYS', 'ORDDATA', 'ORDSYS', 'SI_INFORMTN_SCHEMA',
	 'MDSYS', 'DVSYS', 'EXFSYS', 'OLAPSYS', 'GGSYS', 'ANONYMOUS', 'XS$NULL', 'OJVMSYS', 'DBSFWUSER', 'REMOTE_SCHEDULER_AGENT', 
	 'SYS$UMF', 'SYSBACKUP', 'SYSDG', 'SYSKM', 'SYSRAC','DVF', 'ORDPLUGINS') 
      AND owner NOT LIKE 'APEX%' 
      AND owner NOT LIKE 'FLOWS_%' 
      AND owner NOT LIKE 'GG%' 
      AND owner NOT LIKE 'RDSADMIN%' 
    GROUP BY owner, segment_type, segment_name 
    ORDER BY SizeMB DESC
) 
WHERE ROWNUM < 11`
	var segments []TopSegment
	err := ExecuteQueryAndScanToStructs(db, &segments, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get top segments info: %w", err)
	}
	logger.Infof("Successfully fetched info for %d top segments.", len(segments))
	// logger.Debugf("Top segments info: %+v", segments)
	return segments, nil
}

// GetObjectDetails gets all object-related information.
// It returns AllObjectInfo and the independent error status of each sub-query.
func GetObjectDetails(db *sql.DB) (allInfo *AllObjectInfo, overviewErr error, topSegmentsErr error, invalidObjectsErr error) {
	logger.Info("Starting to fetch object module information...")
	allInfo = &AllObjectInfo{}

	allInfo.Overview, overviewErr = getObjectOverview(db)
	if overviewErr != nil {
		logger.Warnf("Error getting object overview: %v", overviewErr)
	}

	allInfo.TopSegments, topSegmentsErr = getTopSegments(db)
	if topSegmentsErr != nil {
		logger.Warnf("Error getting top segments info: %v", topSegmentsErr)
	}

	allInfo.InvalidObjects, invalidObjectsErr = getInvalidObjects(db)
	if invalidObjectsErr != nil {
		logger.Warnf("Error getting invalid objects info: %v", invalidObjectsErr)
	}

	logger.Info("Object module information fetching complete.")
	return
}

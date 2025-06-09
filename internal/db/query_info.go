package db

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	// "github.com/goodwaysIT/inspect4oracle/internal/logger" // Logger might be unused now
	"github.com/goodwaysIT/inspect4oracle/internal/logger" // Keep logger for now, for warnings

	_ "github.com/sijms/go-ora/v2" // Oracle driver
)

// InstanceInfo holds information about a specific database instance.
type InstanceInfo struct {
	InstanceNumber int    `json:"instance_number" db:"INSTANCE_NUMBER"`
	InstanceName   string `json:"instance_name" db:"INSTANCE_NAME"`
	HostName       string `json:"host_name" db:"HOST_NAME"`
	Version        string `json:"version" db:"VERSION"`
	StartupTime    string `json:"startup_time" db:"STARTUP_TIME"`
	Status         string `json:"status" db:"STATUS"`
	DatabaseStatus string `json:"database_status" db:"DATABASE_STATUS"`
	InstanceRole   string `json:"instance_role" db:"INSTANCE_ROLE"`
	Archiver       string `json:"archiver" db:"ARCHIVER"`
}

// DatabaseDetail holds detailed information about the Oracle database.
type DatabaseDetail struct {
	DBID                 sql.NullInt64  `json:"dbid" db:"DBID"`
	Name                 sql.NullString `json:"name" db:"NAME"`
	Created              sql.NullString `json:"created" db:"CREATED"`
	LogMode              string         `json:"log_mode" db:"LOG_MODE"`
	OpenMode             string         `json:"open_mode" db:"OPEN_MODE"`
	CDB                  sql.NullString `json:"cdb" db:"CDB"` // YES/NO, applicable for 12c+
	DatabaseRole         string         `json:"database_role" db:"DATABASE_ROLE"`
	ProtectionMode       string         `json:"protection_mode" db:"PROTECTION_MODE"`
	ForceLogging         sql.NullString `json:"force_logging" db:"FORCE_LOGGING"`
	FlashbackOn          string         `json:"flashback_on" db:"FLASHBACK_ON"`
	PlatformName         string         `json:"platform_name" db:"PLATFORM_NAME"`
	DBUniqueName         sql.NullString `json:"db_unique_name" db:"DB_UNIQUE_NAME"`
	CharacterSet         sql.NullString `json:"character_set" db:"CHARACTER_SET"`                   // For 12c+
	NationalCharacterSet sql.NullString `json:"national_character_set" db:"NATIONAL_CHARACTER_SET"` // For 12c+
	OverallVersion       string         `json:"overall_version"`                                    // Version string from instance, used for logic
}

// FullDBInfo encapsulates all collected database and instance information.
type FullDBInfo struct {
	Instances []InstanceInfo `json:"instances"`
	Database  DatabaseDetail `json:"database"`
}

// GetDatabaseInfo retrieves comprehensive information about the Oracle database and its instances.
func GetDatabaseInfo(db *sql.DB) (*FullDBInfo, error) {
	var fullInfo FullDBInfo
	var firstInstanceVersion string

	// Query 1: Get instance information from gv$instance
	instanceQuery := `
SELECT instance_number, instance_name, host_name, version, 
       TO_CHAR(startup_time, 'YYYY-MM-DD HH24:MI:SS') as startup_time, 
       status, database_status, instance_role, archiver 
FROM gv$instance ORDER BY instance_number`

	err := ExecuteQueryAndScanToStructs(db, &fullInfo.Instances, instanceQuery)
	if err != nil {
		return nil, fmt.Errorf("error querying gv$instance using generic scan: %w", err)
	}

	if len(fullInfo.Instances) == 0 {
		return nil, fmt.Errorf("no instance information found from gv$instance")
	}
	firstInstanceVersion = fullInfo.Instances[0].Version
	fullInfo.Database.OverallVersion = firstInstanceVersion

	// Determine major version for conditional query
	majorVersion := 0
	if len(firstInstanceVersion) > 0 {
		parts := strings.Split(firstInstanceVersion, ".")
		if len(parts) > 0 {
			majorVersion, _ = strconv.Atoi(parts[0]) // Error ignored, default to 0 if parsing fails
		}
	}

	// Query 2: Get database details (v$database and NLS parameters)
	var dbDetailQuery string
	// Ensure column aliases in the query match struct field names (case-insensitively)
	// or are handled by struct tags if ExecuteQueryAndScanToStructs is enhanced to use them.
	if majorVersion >= 12 {
		dbDetailQuery = `
SELECT dbid, name, TO_CHAR(created, 'YYYY-MM-DD HH24:MI:SS') as created, log_mode, open_mode, cdb, 
       database_role, protection_mode, force_logging, flashback_on, platform_name, 
       db_unique_name, 
       (SELECT value FROM nls_database_parameters WHERE parameter = 'NLS_CHARACTERSET') AS character_set, 
       (SELECT value FROM nls_database_parameters WHERE parameter = 'NLS_NCHAR_CHARACTERSET') AS national_character_set 
FROM v$database`
	} else {
		// Fallback for versions < 12c
		dbDetailQuery = `
SELECT dbid, name, TO_CHAR(created, 'YYYY-MM-DD HH24:MI:SS') as created, log_mode, open_mode, 'NO' as cdb, database_role, protection_mode,
           force_logging, flashback_on, platform_name, db_unique_name,
           (SELECT value FROM nls_database_parameters WHERE parameter = 'NLS_CHARACTERSET') AS character_set, 
           (SELECT value FROM nls_database_parameters WHERE parameter = 'NLS_NCHAR_CHARACTERSET') AS national_character_set 
FROM v$database`
	}

	var dbDetails []DatabaseDetail
	err = ExecuteQueryAndScanToStructs(db, &dbDetails, dbDetailQuery)
	if err != nil {
		// Log the error but potentially return partial instance info if that's desired behavior
		logger.Warnf("error querying database details using generic scan: %v. Instance info might be available.", err)
		// Depending on requirements, you might return fullInfo here with an error, or just the error.
		// For now, let's assume if db details fail, we still want to try returning what we have with a clear error.
		return &fullInfo, fmt.Errorf("error querying database details: %w (instance info might be present but incomplete overall)", err)
	}

	if len(dbDetails) > 0 {
		fullInfo.Database = dbDetails[0]
		fullInfo.Database.OverallVersion = firstInstanceVersion // Ensure OverallVersion is set from instance info
	} else {
		// No rows returned for database details, this could be an issue or expected in some DB states.
		logger.Warnf("no database detail rows returned from v$database query. Instance info might be available.")
		// Return partial info. The DatabaseDetail struct will have zero values.
		// Consider if this state warrants an error or just a warning and partial data.
		// For consistency with original logic (which returned fullInfo and nil error on sql.ErrNoRows for this part):
		return &fullInfo, nil
	}

	return &fullInfo, nil
}

// 其他巡检项实现已拆分到 storage_queries.go、params_queries.go 等独立文件。请在对应文件查找实现。

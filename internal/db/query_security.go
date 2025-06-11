package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/goodwaysIT/inspect4oracle/internal/logger"
)

// NonSystemUserInfo contains information about non-system users queried from DBA_USERS
type NonSystemUserInfo struct {
	Username            string       `json:"username"`
	AccountStatus       string       `json:"account_status"`
	LockDate            sql.NullTime `json:"lock_date"`
	ExpiryDate          sql.NullTime `json:"expiry_date"`
	DefaultTablespace   string       `json:"default_tablespace"`
	TemporaryTablespace string       `json:"temporary_tablespace"`
	Profile             string       `json:"profile"`
	Created             time.Time    `json:"created"`
	LastLogin           sql.NullTime `json:"last_login"` // Note: DBA_USERS.LAST_LOGIN may not be populated in all versions or configurations
}

// GetNonSystemUsers gets information for all non-system users
func GetNonSystemUsers(db *sql.DB) ([]NonSystemUserInfo, error) {
	query := `
SELECT 
    USERNAME AS Username, 
    ACCOUNT_STATUS AS AccountStatus, 
    LOCK_DATE AS LockDate, 
    EXPIRY_DATE AS ExpiryDate, 
    DEFAULT_TABLESPACE AS DefaultTablespace, 
    TEMPORARY_TABLESPACE AS TemporaryTablespace, 
    PROFILE AS Profile, 
    CREATED AS Created,
    LAST_LOGIN AS LastLogin
FROM DBA_USERS
WHERE (ORACLE_MAINTAINED = 'N' OR ORACLE_MAINTAINED IS NULL) AND USERNAME NOT IN (
    -- Common known non-application accounts, can be adjusted according to the actual situation
    'ANONYMOUS', 'APEX_PUBLIC_USER', 'AUDSYS', 'BI', 'CTXSYS', 'DBSFWUSER', 
    'DBSNMP', 'DIP', 'DMSYS', 'DVF', 'DVSYS', 'EXFSYS', 'FLOWS_FILES', 
    'GGSYS', 'GSMADMIN_INTERNAL', 'GSMCATUSER', 'GSMUSER', 'HR', 'IX', 'LBACSYS', 
    'MDDATA', 'MDSYS', 'MGMT_VIEW', 'OE', 'OLAPSYS', 'ORACLE_OCM', 'ORDDATA', 
    'ORDPLUGINS', 'ORDSYS', 'OUTLN', 'PDBADMIN', 'PM', 'REMOTE_SCHEDULER_AGENT', 
    'SCOTT', 'SH', 'SI_INFORMTN_SCHEMA', 'SPATIAL_CSW_ADMIN_USR', 
    'SPATIAL_WFS_ADMIN_USR', 'SYS$UMF', 'SYSBACKUP', 'SYSDG', 'SYSKM', 'SYSRAC', 
    'SYSTEM', 'SYS', 'TSMSYS', 'WKPROXY', 'WMSYS', 'XDB', 'XS$NULL'
)
ORDER BY USERNAME`
	// Note: ORACLE_MAINTAINED = 'N' is a good way to distinguish whether a user is maintained by Oracle internally in 12c and later versions.
	// For earlier versions, a longer hard-coded exclusion list may be required.
	// (ORACLE_MAINTAINED IS NULL) is for compatibility with earlier versions where this column may not exist, or where the column value is NULL.

	var users []NonSystemUserInfo
	// Assume ExecuteQueryAndScanToStructs can handle sql.NullTime and time.Time
	err := ExecuteQueryAndScanToStructs(db, &users, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get non-system user info: %w", err)
	}
	logger.Infof("Successfully fetched info for %d non-system users.", len(users))
	return users, nil
}

// ProfileInfo contains profile configuration information, especially password-related parameters
type ProfileInfo struct {
	Profile      string `json:"profile"`
	ResourceName string `json:"resource_name"`
	Limit        string `json:"limit"`
}

// GetProfiles gets profile configuration information (focusing on password policies and the DEFAULT profile)
func GetProfiles(db *sql.DB) ([]ProfileInfo, error) {
	query := `
SELECT 
    PROFILE AS Profile, 
    RESOURCE_NAME AS ResourceName, 
    LIMIT AS Limit
FROM DBA_PROFILES
WHERE PROFILE != 'DEFAULT'  -- All settings for all user-defined Profiles
   OR (PROFILE = 'DEFAULT' AND RESOURCE_TYPE = 'PASSWORD') -- And password-related settings for the DEFAULT Profile
ORDER BY PROFILE, RESOURCE_NAME`

	var profiles []ProfileInfo
	err := ExecuteQueryAndScanToStructs(db, &profiles, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get profile configuration info: %w", err)
	}
	logger.Infof("Successfully fetched %d profile configuration entries.", len(profiles))
	return profiles, nil
}

// NonSystemRoleInfo contains information about non-system roles
type NonSystemRoleInfo struct {
	RoleName           string `json:"role_name"`
	AuthenticationType string `json:"authentication_type"` // NONE, PASSWORD, EXTERNAL, GLOBAL
}

// GetNonSystemRoles gets all roles not maintained by Oracle
func GetNonSystemRoles(db *sql.DB) ([]NonSystemRoleInfo, error) {
	query := `
SELECT ROLE AS RoleName, AUTHENTICATION_TYPE
FROM DBA_ROLES
WHERE (ORACLE_MAINTAINED = 'N' OR ORACLE_MAINTAINED IS NULL) -- IS NULL for older versions or if column doesn't exist
ORDER BY ROLE`

	var roles []NonSystemRoleInfo
	err := ExecuteQueryAndScanToStructs(db, &roles, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get non-system role list: %w", err)
	}
	logger.Infof("Successfully fetched %d non-system roles.", len(roles))
	return roles, nil
}

// UserPrivilegedRoleInfo contains information about privileged roles granted to users
type UserPrivilegedRoleInfo struct {
	Grantee     string `json:"grantee"`      // 用户名或角色名
	GrantedRole string `json:"granted_role"` // The role that was granted
	AdminOption string `json:"admin_option"` // YES/NO
	DefaultRole string `json:"default_role"` // YES/NO
}

// GetUsersWithPrivilegedRoles gets information about users who have been granted privileged roles (focusing on non-system users)
func GetUsersWithPrivilegedRoles(db *sql.DB) ([]UserPrivilegedRoleInfo, error) {
	query := `
SELECT drp.GRANTEE, drp.GRANTED_ROLE, drp.ADMIN_OPTION, drp.DEFAULT_ROLE
FROM DBA_ROLE_PRIVS drp
JOIN DBA_USERS u ON drp.GRANTEE = u.USERNAME
WHERE (u.ORACLE_MAINTAINED = 'N' OR u.ORACLE_MAINTAINED IS NULL)
  AND drp.GRANTED_ROLE IN ('DBA', 'SYSDBA', 'SYSOPER', 'AQ_ADMINISTRATOR_ROLE', 'AQ_USER_ROLE', 'SCHEDULER_ADMIN', 'HS_ADMIN_ROLE', 'IMP_FULL_DATABASE', 'EXP_FULL_DATABASE', 'DATAPUMP_IMP_FULL_DATABASE', 'DATAPUMP_EXP_FULL_DATABASE', 'GATHER_SYSTEM_STATISTICS', 'LOGSTDBY_ADMINISTRATOR', 'RECOVERY_CATALOG_OWNER') -- 常见高权限或敏感角色，可调整
ORDER BY drp.GRANTEE, drp.GRANTED_ROLE`

	var userRoles []UserPrivilegedRoleInfo
	err := ExecuteQueryAndScanToStructs(db, &userRoles, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get user privileged role info: %w", err)
	}
	logger.Infof("Successfully fetched %d user privileged role entries.", len(userRoles))
	return userRoles, nil
}

// UserSystemPrivilegeInfo contains information about system privileges granted to users
type UserSystemPrivilegeInfo struct {
	Grantee     string `json:"grantee"`      // 用户名
	Privilege   string `json:"privilege"`    // 系统权限
	AdminOption string `json:"admin_option"` // YES/NO
}

// GetUsersWithSystemPrivileges gets information about non-system users who have been granted system privileges
func GetUsersWithSystemPrivileges(db *sql.DB) ([]UserSystemPrivilegeInfo, error) {
	query := `
SELECT 
    dsp.GRANTEE AS Grantee, 
    dsp.PRIVILEGE AS Privilege, 
    dsp.ADMIN_OPTION AS AdminOption
FROM DBA_SYS_PRIVS dsp
JOIN DBA_USERS u ON dsp.GRANTEE = u.USERNAME
WHERE (u.ORACLE_MAINTAINED = 'N' OR u.ORACLE_MAINTAINED IS NULL)
ORDER BY dsp.GRANTEE, dsp.PRIVILEGE`

	var userSysPrivs []UserSystemPrivilegeInfo
	err := ExecuteQueryAndScanToStructs(db, &userSysPrivs, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get user system privilege info: %w", err)
	}
	logger.Infof("Successfully fetched %d user system privilege entries.", len(userSysPrivs))
	return userSysPrivs, nil
}

// RoleToRoleGrantInfo contains information about roles granted to other roles
type RoleToRoleGrantInfo struct {
	Role        string `json:"role"`         // The role to which privileges are granted
	GrantedRole string `json:"granted_role"` // The role that was granted
	AdminOption string `json:"admin_option"` // YES/NO
}

// GetRoleToRoleGrants gets information about roles granted to other roles (mainly focusing on cases where the grantor is a non-system role)
func GetRoleToRoleGrants(db *sql.DB) ([]RoleToRoleGrantInfo, error) {
	query := `
SELECT rrp.ROLE, rrp.GRANTED_ROLE, rrp.ADMIN_OPTION
FROM ROLE_ROLE_PRIVS rrp
LEFT JOIN DBA_ROLES dr_role ON rrp.ROLE = dr_role.ROLE -- 授予者角色
-- LEFT JOIN DBA_ROLES dr_granted_role ON rrp.GRANTED_ROLE = dr_granted_role.ROLE -- 被授予者角色 (可选，如果也想过滤被授予者)
WHERE (dr_role.ORACLE_MAINTAINED = 'N' OR dr_role.ORACLE_MAINTAINED IS NULL) -- 确保授予者是非系统角色
  AND rrp.ROLE != rrp.GRANTED_ROLE -- 排除角色授予自身
ORDER BY rrp.ROLE, rrp.GRANTED_ROLE`

	var roleGrants []RoleToRoleGrantInfo
	err := ExecuteQueryAndScanToStructs(db, &roleGrants, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get role-to-role grant info: %w", err)
	}
	logger.Infof("Successfully fetched %d role-to-role grant entries.", len(roleGrants))
	return roleGrants, nil
}

// TODO: Further security checks like audit settings etc.

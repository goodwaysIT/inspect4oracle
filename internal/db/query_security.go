package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/goodwaysIT/inspect4oracle/internal/logger"
)

// NonSystemUserInfo 包含从DBA_USERS查询的非系统用户信息
type NonSystemUserInfo struct {
	Username            string       `json:"username"`
	AccountStatus       string       `json:"account_status"`
	LockDate            sql.NullTime `json:"lock_date"`
	ExpiryDate          sql.NullTime `json:"expiry_date"`
	DefaultTablespace   string       `json:"default_tablespace"`
	TemporaryTablespace string       `json:"temporary_tablespace"`
	Profile             string       `json:"profile"`
	Created             time.Time    `json:"created"`
	LastLogin           sql.NullTime `json:"last_login"` // 注意：DBA_USERS.LAST_LOGIN 可能不被所有版本或配置填充
}

// GetNonSystemUsers 获取所有非系统用户的信息
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
    -- 常见的已知非应用账户，可以根据实际情况调整
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
	// 注意: ORACLE_MAINTAINED = 'N' 是12c及以后版本区分用户是否为Oracle内部维护的一个好方法。
	// 对于更早的版本，可能需要依赖一个更长的硬编码排除列表。
	// (ORACLE_MAINTAINED IS NULL) 是为了兼容可能不存在该列的更早版本，或者该列值为NULL的情况。

	var users []NonSystemUserInfo
	// 假设 ExecuteQueryAndScanToStructs 能够处理 sql.NullTime 和 time.Time
	err := ExecuteQueryAndScanToStructs(db, &users, query)
	if err != nil {
		return nil, fmt.Errorf("获取非系统用户信息失败: %w", err)
	}
	logger.Infof("成功获取 %d 个非系统用户信息。", len(users))
	return users, nil
}

// ProfileInfo 包含 Profile 的配置信息，特别是密码相关参数
type ProfileInfo struct {
	Profile      string `json:"profile"`
	ResourceName string `json:"resource_name"`
	Limit        string `json:"limit"`
}

// GetProfiles 获取 Profile 配置信息 (重点关注密码策略和 DEFAULT profile)
func GetProfiles(db *sql.DB) ([]ProfileInfo, error) {
	query := `
SELECT 
    PROFILE AS Profile, 
    RESOURCE_NAME AS ResourceName, 
    LIMIT AS Limit
FROM DBA_PROFILES
WHERE PROFILE != 'DEFAULT'  -- 所有用户自定义 Profile 的所有设置
   OR (PROFILE = 'DEFAULT' AND RESOURCE_TYPE = 'PASSWORD') -- 以及 DEFAULT Profile 的密码相关设置
ORDER BY PROFILE, RESOURCE_NAME`

	var profiles []ProfileInfo
	err := ExecuteQueryAndScanToStructs(db, &profiles, query)
	if err != nil {
		return nil, fmt.Errorf("获取 Profile 配置信息失败: %w", err)
	}
	logger.Infof("成功获取 %d 条 Profile 配置信息。", len(profiles))
	return profiles, nil
}

// NonSystemRoleInfo 包含非系统角色的信息
type NonSystemRoleInfo struct {
	RoleName           string `json:"role_name"`
	AuthenticationType string `json:"authentication_type"` // NONE, PASSWORD, EXTERNAL, GLOBAL
}

// GetNonSystemRoles 获取所有非 Oracle 维护的角色
func GetNonSystemRoles(db *sql.DB) ([]NonSystemRoleInfo, error) {
	query := `
SELECT ROLE AS RoleName, AUTHENTICATION_TYPE
FROM DBA_ROLES
WHERE (ORACLE_MAINTAINED = 'N' OR ORACLE_MAINTAINED IS NULL) -- IS NULL for older versions or if column doesn't exist
ORDER BY ROLE`

	var roles []NonSystemRoleInfo
	err := ExecuteQueryAndScanToStructs(db, &roles, query)
	if err != nil {
		return nil, fmt.Errorf("获取非系统角色列表失败: %w", err)
	}
	logger.Infof("成功获取 %d 个非系统角色。", len(roles))
	return roles, nil
}

// UserPrivilegedRoleInfo 包含用户被授予的特权角色信息
type UserPrivilegedRoleInfo struct {
	Grantee     string `json:"grantee"`      // 用户名或角色名
	GrantedRole string `json:"granted_role"` // 被授予的角色
	AdminOption string `json:"admin_option"` // YES/NO
	DefaultRole string `json:"default_role"` // YES/NO
}

// GetUsersWithPrivilegedRoles 获取被授予了特权角色的用户信息 (重点关注非系统用户)
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
		return nil, fmt.Errorf("获取用户特权角色信息失败: %w", err)
	}
	logger.Infof("成功获取 %d 条用户特权角色信息。", len(userRoles))
	return userRoles, nil
}

// UserSystemPrivilegeInfo 包含用户被授予的系统权限信息
type UserSystemPrivilegeInfo struct {
	Grantee     string `json:"grantee"`      // 用户名
	Privilege   string `json:"privilege"`    // 系统权限
	AdminOption string `json:"admin_option"` // YES/NO
}

// GetUsersWithSystemPrivileges 获取被授予了系统权限的非系统用户信息
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
		return nil, fmt.Errorf("获取用户系统权限信息失败: %w", err)
	}
	logger.Infof("成功获取 %d 条用户系统权限信息。", len(userSysPrivs))
	return userSysPrivs, nil
}

// RoleToRoleGrantInfo 包含角色授予其他角色的信息
type RoleToRoleGrantInfo struct {
	Role        string `json:"role"`         // 授予权限的角色
	GrantedRole string `json:"granted_role"` // 被授予的角色
	AdminOption string `json:"admin_option"` // YES/NO
}

// GetRoleToRoleGrants 获取角色授予其他角色的信息 (主要关注授予者为非系统角色的情况)
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
		return nil, fmt.Errorf("获取角色授予角色信息失败: %w", err)
	}
	logger.Infof("成功获取 %d 条角色授予角色信息。", len(roleGrants))
	return roleGrants, nil
}

// TODO: Further security checks like audit settings etc.

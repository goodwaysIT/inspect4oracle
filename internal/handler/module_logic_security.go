package handler

import (
	"database/sql"
	"fmt"

	// Required for DB version check
	// Required for strings.Split
	"github.com/goodwaysIT/inspect4oracle/internal/db"
	"github.com/goodwaysIT/inspect4oracle/internal/logger"
)

// processSecurityModule handles the "security" inspection item.
func processSecurityModule(dbConn *sql.DB, lang string, fullDBInfo *db.FullDBInfo) (cards []ReportCard, tables []*ReportTable, charts []ReportChart, overallErr error) {
	logger.Infof("开始处理安全模块... 语言: %s", lang)

	// 1. 获取非系统用户信息
	nonSystemUsers, err := db.GetNonSystemUsers(dbConn)
	if err != nil {
		logger.Errorf("处理安全模块 - 获取非系统用户信息失败: %v", err)
		cards = append(cards, ReportCard{
			Title: langText("非系统用户信息错误", "Non-System User Info Error", lang),
			Value: fmt.Sprintf(langText("获取非系统用户信息失败: %v", "Failed to get non-system user info: %v", lang), err),
		})
		if overallErr == nil {
			overallErr = err
		}
	} else {
		if len(nonSystemUsers) > 0 {
			usersTable := &ReportTable{
				Name: langText("非系统用户账户", "Non-System User Accounts", lang),
				Headers: []string{
					langText("用户名", "Username", lang),
					langText("账户状态", "Account Status", lang),
					langText("锁定日期", "Lock Date", lang),
					langText("过期日期", "Expiry Date", lang),
					langText("默认表空间", "Default Tablespace", lang),
					langText("临时表空间", "Temp Tablespace", lang),
					langText("Profile", "Profile", lang),
					langText("创建日期", "Created", lang),
					langText("上次登录", "Last Login", lang),
				},
				Rows: [][]string{},
			}
			for _, u := range nonSystemUsers {
				lockDateStr := "N/A"
				if u.LockDate.Valid {
					lockDateStr = u.LockDate.Time.Format("2006-01-02 15:04:05")
				}
				expiryDateStr := "N/A"
				if u.ExpiryDate.Valid {
					expiryDateStr = u.ExpiryDate.Time.Format("2006-01-02 15:04:05")
				}
				createdStr := u.Created.Format("2006-01-02 15:04:05")
				lastLoginStr := "N/A"
				if u.LastLogin.Valid {
					lastLoginStr = u.LastLogin.Time.Format("2006-01-02 15:04:05")
				}
				row := []string{
					u.Username,
					u.AccountStatus,
					lockDateStr,
					expiryDateStr,
					u.DefaultTablespace,
					u.TemporaryTablespace,
					u.Profile,
					createdStr,
					lastLoginStr,
				}
				usersTable.Rows = append(usersTable.Rows, row)
			}
			tables = append(tables, usersTable)
		} else {
			cards = append(cards, ReportCard{
				Title: langText("非系统用户账户", "Non-System User Accounts", lang),
				Value: langText("未发现符合条件的非系统用户账户。", "No non-system user accounts found matching criteria.", lang),
			})
		}
	}

	// 2. 获取 Profile 配置信息
	profiles, err := db.GetProfiles(dbConn)
	if err != nil {
		logger.Errorf("处理安全模块 - 获取 Profile 配置信息失败: %v", err)
		cards = append(cards, ReportCard{
			Title: langText("Profile 配置错误", "Profile Configuration Error", lang),
			Value: fmt.Sprintf(langText("获取 Profile 配置信息失败: %v", "Failed to get Profile configuration: %v", lang), err),
		})
		if overallErr == nil {
			overallErr = err
		}
	} else {
		if len(profiles) > 0 {
			profilesTable := &ReportTable{
				Name: langText("Profile 配置 (密码策略与DEFAULT)", "Profile Configuration (Password Policies & DEFAULT)", lang),
				Headers: []string{
					langText("Profile 名称", "Profile Name", lang),
					langText("资源名称", "Resource Name", lang),
					langText("限制值", "Limit", lang),
				},
				Rows: [][]string{},
			}
			for _, p := range profiles {
				row := []string{
					p.Profile,
					p.ResourceName,
					p.Limit,
				}
				profilesTable.Rows = append(profilesTable.Rows, row)
			}
			tables = append(tables, profilesTable)
		} else {
			cards = append(cards, ReportCard{
				Title: langText("Profile 配置", "Profile Configuration", lang),
				Value: langText("未发现相关的 Profile 配置信息。", "No relevant Profile configurations found.", lang),
			})
		}
	}

	// 3. 获取非系统角色列表
	nonSystemRoles, err := db.GetNonSystemRoles(dbConn)
	if err != nil {
		logger.Errorf("处理安全模块 - 获取非系统角色列表失败: %v", err)
		cards = append(cards, ReportCard{
			Title: langText("非系统角色错误", "Non-System Roles Error", lang),
			Value: fmt.Sprintf(langText("获取非系统角色列表失败: %v", "Failed to get non-system roles: %v", lang), err),
		})
		if overallErr == nil {
			overallErr = err
		}
	} else {
		if len(nonSystemRoles) > 0 {
			rolesTable := &ReportTable{
				Name: langText("非系统角色", "Non-System Roles", lang),
				Headers: []string{
					langText("角色名称", "Role Name", lang),
					langText("认证类型", "Authentication Type", lang),
				},
				Rows: [][]string{},
			}
			for _, r := range nonSystemRoles {
				row := []string{r.RoleName, r.AuthenticationType}
				rolesTable.Rows = append(rolesTable.Rows, row)
			}
			tables = append(tables, rolesTable)
		} else {
			cards = append(cards, ReportCard{
				Title: langText("非系统角色", "Non-System Roles", lang),
				Value: langText("未发现非系统角色。", "No non-system roles found.", lang),
			})
		}
	}

	// 4. 获取拥有特权角色的用户列表
	usersWithPrivRoles, err := db.GetUsersWithPrivilegedRoles(dbConn)
	if err != nil {
		logger.Errorf("处理安全模块 - 获取用户特权角色信息失败: %v", err)
		cards = append(cards, ReportCard{
			Title: langText("用户特权角色错误", "User Privileged Roles Error", lang),
			Value: fmt.Sprintf(langText("获取用户特权角色信息失败: %v", "Failed to get user privileged roles: %v", lang), err),
		})
		if overallErr == nil {
			overallErr = err
		}
	} else {
		if len(usersWithPrivRoles) > 0 {
			userPrivRolesTable := &ReportTable{
				Name: langText("拥有特权角色的用户", "Users with Privileged Roles", lang),
				Headers: []string{
					langText("用户名", "Username", lang),
					langText("授予的角色", "Granted Role", lang),
					langText("Admin Option", "Admin Option", lang),
				},
				Rows: [][]string{},
			}
			for _, upr := range usersWithPrivRoles {
				row := []string{upr.Grantee, upr.GrantedRole, upr.AdminOption}
				userPrivRolesTable.Rows = append(userPrivRolesTable.Rows, row)
			}
			tables = append(tables, userPrivRolesTable)
		} else {
			cards = append(cards, ReportCard{
				Title: langText("用户特权角色", "User Privileged Roles", lang),
				Value: langText("未发现拥有特权角色的用户。", "No users found with privileged roles.", lang),
			})
		}
	}

	return cards, tables, nil, overallErr
}

package handler

import (
	"database/sql"
	"fmt"

	// Required for DB version check
	// Required for strings.Split
	"github.com/goodwaysIT/inspect4oracle/internal/db"
	"github.com/goodwaysIT/inspect4oracle/internal/logger"
)

// generateNonSystemUsersTable fetches non-system user information and prepares a ReportTable or ReportCard.
func generateNonSystemUsersTable(dbConn *sql.DB, lang string) (*ReportTable, *ReportCard, error) {
	nonSystemUsers, err := db.GetNonSystemUsers(dbConn)
	if err != nil {
		msg := fmt.Sprintf(langText("获取非系统用户信息失败: %v", "Failed to get non-system user info: %v", "非システムユーザー情報の取得に失敗しました: %v", lang), err)
		card := &ReportCard{
			Title: langText("非系统用户信息错误", "Non-System User Info Error", "非システムユーザー情報エラー", lang),
			Value: msg,
		}
		return nil, card, fmt.Errorf("failed to get non-system users: %w", err)
	}

	if len(nonSystemUsers) > 0 {
		usersTable := &ReportTable{
			Name: langText("非系统用户账户", "Non-System User Accounts", "非システムユーザーアカウント", lang),
			Headers: []string{
				langText("用户名", "Username", "ユーザー名", lang),
				langText("状态", "Status", "ステータス", lang),
				langText("锁定时间", "Lock Time", "ロック時間", lang),
				langText("过期时间", "Expiry Time", "有効期限", lang),
				langText("默认表空间", "Default Tablespace", "デフォルトテーブルスペース", lang),
				langText("临时表空间", "Temp Tablespace", "一時テーブルスペース", lang),
				langText("配置文件", "Profile", "プロファイル", lang),
				langText("创建时间", "Created Time", "作成時間", lang),
				langText("最后登录时间", "Last Login Time", "最終ログイン時間", lang),
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
		return usersTable, nil, nil
	} else {
		card := &ReportCard{
			Title: langText("非系统用户账户", "Non-System User Accounts", "非システムユーザーアカウント", lang),
			Value: langText("未发现符合条件的非系统用户账户。", "No non-system user accounts found matching criteria.", "基準に一致する非システムユーザーアカウントが見つかりませんでした。", lang),
		}
		return nil, card, nil
	}
}

// generateProfilesTable fetches Profile configuration information and prepares a ReportTable or ReportCard.
func generateProfilesTable(dbConn *sql.DB, lang string) (*ReportTable, *ReportCard, error) {
	profiles, err := db.GetProfiles(dbConn)
	if err != nil {
		msg := fmt.Sprintf(langText("获取配置文件失败: %v", "Failed to get Profile configuration: %v", "プロファイル構成の取得に失敗しました: %v", lang), err)
		card := &ReportCard{
			Title: langText("配置文件错误", "Profile Configuration Error", "プロファイル構成エラー", lang),
			Value: msg,
		}
		return nil, card, fmt.Errorf("failed to get profiles: %w", err)
	}

	if len(profiles) > 0 {
		profilesTable := &ReportTable{
			Name: langText("配置文件 (密码策略 & 默认)", "Profile Configuration (Password Policies & DEFAULT)", "プロファイル構成 (パスワードポリシー & デフォルト)", lang),
			Headers: []string{
				langText("配置文件名称", "Profile Name", "プロファイル名", lang),
				langText("资源名称", "Resource Name", "リソース名", lang),
				langText("限制值", "Limit", "制限値", lang),
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
		return profilesTable, nil, nil
	} else {
		card := &ReportCard{
			Title: langText("配置文件", "Profile Configuration", "プロファイル構成", lang),
			Value: langText("未找到相关的配置文件。", "No relevant Profile configurations found.", "関連するプロファイル構成が見つかりませんでした。", lang),
		}
		return nil, card, nil
	}
}

// generateNonSystemRolesTable fetches non-system roles information and prepares a ReportTable or ReportCard.
func generateNonSystemRolesTable(dbConn *sql.DB, lang string) (*ReportTable, *ReportCard, error) {
	nonSystemRoles, err := db.GetNonSystemRoles(dbConn)
	if err != nil {
		msg := fmt.Sprintf(langText("获取非系统角色失败: %v", "Failed to get non-system roles: %v", "非システムロールの取得に失敗しました: %v", lang), err)
		card := &ReportCard{
			Title: langText("非系统角色错误", "Non-System Roles Error", "非システムロールエラー", lang),
			Value: msg,
		}
		return nil, card, fmt.Errorf("failed to get non-system roles: %w", err)
	}

	if len(nonSystemRoles) > 0 {
		rolesTable := &ReportTable{
			Name: langText("非系统角色", "Non-System Roles", "非システムロール", lang),
			Headers: []string{
				langText("角色名称", "Role Name", "ロール名", lang),
				langText("认证类型", "Authentication Type", "認証タイプ", lang),
			},
			Rows: [][]string{},
		}
		for _, r := range nonSystemRoles {
			row := []string{r.RoleName, r.AuthenticationType}
			rolesTable.Rows = append(rolesTable.Rows, row)
		}
		return rolesTable, nil, nil
	} else {
		card := &ReportCard{
			Title: langText("非系统角色", "Non-System Roles", "非システムロール", lang),
			Value: langText("未发现非系统角色。", "No non-system roles found.", "非システムロールが見つかりませんでした。", lang),
		}
		return nil, card, nil
	}
}

// generateUsersWithPrivilegedRolesTable fetches users with privileged roles and prepares a ReportTable or ReportCard.
func generateUsersWithPrivilegedRolesTable(dbConn *sql.DB, lang string) (*ReportTable, *ReportCard, error) {
	usersWithPrivRoles, err := db.GetUsersWithPrivilegedRoles(dbConn)
	if err != nil {
		msg := fmt.Sprintf(langText("获取用户特权角色失败: %v", "Failed to get user privileged roles: %v", "ユーザーの特権ロールの取得に失敗しました: %v", lang), err)
		card := &ReportCard{
			Title: langText("用户特权角色错误", "User Privileged Roles Error", "ユーザー特権ロールエラー", lang),
			Value: msg,
		}
		return nil, card, fmt.Errorf("failed to get users with privileged roles: %w", err)
	}

	if len(usersWithPrivRoles) > 0 {
		userPrivRolesTable := &ReportTable{
			Name: langText("拥有特权角色的用户", "Users with Privileged Roles", "特権ロールを持つユーザー", lang),
			Headers: []string{
				langText("用户名", "Username", "ユーザー名", lang),
				langText("授予的角色", "Granted Role", "付与されたロール", lang),
				langText("Admin Option", "Admin Option", "管理者オプション", lang),
			},
			Rows: [][]string{},
		}
		for _, upr := range usersWithPrivRoles {
			row := []string{upr.Grantee, upr.GrantedRole, upr.AdminOption}
			userPrivRolesTable.Rows = append(userPrivRolesTable.Rows, row)
		}
		return userPrivRolesTable, nil, nil
	} else {
		card := &ReportCard{
			Title: langText("用户特权角色", "User Privileged Roles", "ユーザー特権ロール", lang),
			Value: langText("未发现拥有特权角色的用户。", "No users found with privileged roles.", "特権ロールを持つユーザーが見つかりませんでした。", lang),
		}
		return nil, card, nil
	}
}

// processSecurityModule handles the "security" inspection item.
func processSecurityModule(dbConn *sql.DB, lang string, fullDBInfo *db.FullDBInfo) (cards []ReportCard, tables []*ReportTable, charts []ReportChart, overallErr error) {
	logger.Infof("Starting to process security module... Language: %s", lang)

	// 1. Get non-system user information
	userTable, userCard, userErr := generateNonSystemUsersTable(dbConn, lang)
	if userErr != nil {
		logger.Errorf("Error processing security module - fetching non-system users: %v", userErr)
		if userCard != nil { // Helper provided a specific error card
			cards = append(cards, *userCard)
		} else { // Generic error card if helper didn't provide one (should not happen with current helper design)
			cards = append(cards, cardFromError("非系统用户信息错误", "Non-System User Info Error", "非システムユーザー情報エラー", userErr, lang))
		}
		if overallErr == nil {
			overallErr = userErr
		}
	} else if userTable != nil {
		tables = append(tables, userTable)
	} else if userCard != nil { // No data card from helper
		cards = append(cards, *userCard)
	}

	// 2. Get Profile configuration information
	profileTable, profileCard, profileErr := generateProfilesTable(dbConn, lang)
	if profileErr != nil {
		logger.Errorf("Error processing security module - fetching profile configurations: %v", profileErr)
		if profileCard != nil { // Helper provided a specific error card
			cards = append(cards, *profileCard)
		} else { // Generic error card
			cards = append(cards, cardFromError("配置文件错误", "Profile Configuration Error", "プロファイル構成エラー", profileErr, lang))
		}
		if overallErr == nil {
			overallErr = profileErr
		}
	} else if profileTable != nil {
		tables = append(tables, profileTable)
	} else if profileCard != nil { // No data card from helper
		cards = append(cards, *profileCard)
	}

	// 3. Get non-system role list
	rolesTable, rolesCard, rolesErr := generateNonSystemRolesTable(dbConn, lang)
	if rolesErr != nil {
		logger.Errorf("Error processing security module - fetching non-system roles: %v", rolesErr)
		if rolesCard != nil { // Helper provided a specific error card
			cards = append(cards, *rolesCard)
		} else { // Generic error card
			cards = append(cards, cardFromError("非系统角色错误", "Non-System Roles Error", "非システムロールエラー", rolesErr, lang))
		}
		if overallErr == nil {
			overallErr = rolesErr
		}
	} else if rolesTable != nil {
		tables = append(tables, rolesTable)
	} else if rolesCard != nil { // No data card from helper
		cards = append(cards, *rolesCard)
	}

	// 4. Get list of users with privileged roles
	userPrivRolesTable, userPrivRolesCard, userPrivRolesErr := generateUsersWithPrivilegedRolesTable(dbConn, lang)
	if userPrivRolesErr != nil {
		logger.Errorf("Error processing security module - fetching user privileged roles: %v", userPrivRolesErr)
		if userPrivRolesCard != nil { // Helper provided a specific error card
			cards = append(cards, *userPrivRolesCard)
		} else { // Generic error card
			cards = append(cards, cardFromError("用户特权角色错误", "User Privileged Roles Error", "ユーザー特権ロールエラー", userPrivRolesErr, lang))
		}
		if overallErr == nil {
			overallErr = userPrivRolesErr
		}
	} else if userPrivRolesTable != nil {
		tables = append(tables, userPrivRolesTable)
	} else if userPrivRolesCard != nil { // No data card from helper
		cards = append(cards, *userPrivRolesCard)
	}

	return cards, tables, nil, overallErr
}

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
		msg := fmt.Sprintf(langText("Failed to get non-system user info: %v", "Failed to get non-system user info: %v", "Failed to get non-system user info: %v", lang), err)
		card := &ReportCard{
			Title: langText("Non-System User Info Error", "Non-System User Info Error", "Non-System User Info Error", lang),
			Value: msg,
		}
		return nil, card, fmt.Errorf("failed to get non-system users: %w", err)
	}

	if len(nonSystemUsers) > 0 {
		usersTable := &ReportTable{
			Name: langText("非系统用户账户", "Non-System User Accounts", "非システムユーザーアカウント", lang),
			Headers: []string{
				langText("Username", "Username", "Username", lang),
				langText("账户状态", "Account Status", "アカウントステータス", lang),
				langText("锁定日期", "Lock Date", "ロック日", lang),
				langText("过期日期", "Expiry Date", "有効期限", lang),
				langText("默认表空间", "Default Tablespace", "デフォルトテーブルスペース", lang),
				langText("临时表空间", "Temp Tablespace", "一時テーブルスペース", lang),
				langText("Profile", "Profile", "プロファイル", lang),
				langText("创建日期", "Created", "作成日", lang),
				langText("上次登录", "Last Login", "最終ログイン", lang),
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
		msg := fmt.Sprintf(langText("Failed to get Profile configuration: %v", "Failed to get Profile configuration: %v", "Failed to get Profile configuration: %v", lang), err)
		card := &ReportCard{
			Title: langText("Profile Configuration Error", "Profile Configuration Error", "Profile Configuration Error", lang),
			Value: msg,
		}
		return nil, card, fmt.Errorf("failed to get profiles: %w", err)
	}

	if len(profiles) > 0 {
		profilesTable := &ReportTable{
			Name: langText("Profile Configuration (Password Policies & DEFAULT)", "Profile Configuration (Password Policies & DEFAULT)", "Profile Configuration (Password Policies & DEFAULT)", lang),
			Headers: []string{
				langText("Profile Name", "Profile Name", "Profile Name", lang),
				langText("Resource Name", "Resource Name", "Resource Name", lang),
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
			Title: langText("Profile Configuration", "Profile Configuration", "Profile Configuration", lang),
			Value: langText("No relevant Profile configurations found.", "No relevant Profile configurations found.", "No relevant Profile configurations found.", lang),
		}
		return nil, card, nil
	}
}

// generateNonSystemRolesTable fetches non-system roles information and prepares a ReportTable or ReportCard.
func generateNonSystemRolesTable(dbConn *sql.DB, lang string) (*ReportTable, *ReportCard, error) {
	nonSystemRoles, err := db.GetNonSystemRoles(dbConn)
	if err != nil {
		msg := fmt.Sprintf(langText("Failed to get non-system roles: %v", "Failed to get non-system roles: %v", "Failed to get non-system roles: %v", lang), err)
		card := &ReportCard{
			Title: langText("Non-System Roles Error", "Non-System Roles Error", "Non-System Roles Error", lang),
			Value: msg,
		}
		return nil, card, fmt.Errorf("failed to get non-system roles: %w", err)
	}

	if len(nonSystemRoles) > 0 {
		rolesTable := &ReportTable{
			Name: langText("非系统角色", "Non-System Roles", "非システムロール", lang),
			Headers: []string{
				langText("Role Name", "Role Name", "Role Name", lang),
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
		msg := fmt.Sprintf(langText("Failed to get user privileged roles: %v", "Failed to get user privileged roles: %v", "Failed to get user privileged roles: %v", lang), err)
		card := &ReportCard{
			Title: langText("User Privileged Roles Error", "User Privileged Roles Error", "User Privileged Roles Error", lang),
			Value: msg,
		}
		return nil, card, fmt.Errorf("failed to get users with privileged roles: %w", err)
	}

	if len(usersWithPrivRoles) > 0 {
		userPrivRolesTable := &ReportTable{
			Name: langText("拥有特权角色的用户", "Users with Privileged Roles", "特権ロールを持つユーザー", lang),
			Headers: []string{
				langText("Username", "Username", "Username", lang),
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
			cards = append(cards, cardFromError("Non-System User Info Error", "Non-System User Info Error", "Non-System User Info Error", userErr, lang))
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
			cards = append(cards, cardFromError("Profile Configuration Error", "Profile Configuration Error", "Profile Configuration Error", profileErr, lang))
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
			cards = append(cards, cardFromError("Non-System Roles Error", "Non-System Roles Error", "Non-System Roles Error", rolesErr, lang))
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
			cards = append(cards, cardFromError("User Privileged Roles Error", "User Privileged Roles Error", "User Privileged Roles Error", userPrivRolesErr, lang))
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

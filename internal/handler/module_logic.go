package handler

import (
	"database/sql"
	"fmt"
	"strconv"

	"github.com/goodwaysIT/inspect4oracle/internal/db"
	"github.com/goodwaysIT/inspect4oracle/internal/logger"
)

// processParametersModule handles the "parameters" inspection item.
func processParametersModule(dbConn *sql.DB, lang string) (cards []ReportCard, tables []*ReportTable, charts []ReportChart, err error) {
	params, dbErr := db.GetParameterList(dbConn)
	if dbErr != nil {
		cards = append(cards, ReportCard{Title: langText("错误", "Error", "エラー", lang), Value: fmt.Sprintf(langText("获取参数失败: %v", "Failed to get parameters: %v", "パラメータの取得に失敗しました: %v", lang), dbErr)})
		return cards, nil, nil, dbErr
	}
	logger.Debugf("Fetched parameters: %v", params)
	if len(params) > 0 {
		paramTable := &ReportTable{
			Name:    langText("参数列表", "Parameter List", "パラメータリスト", lang),
			Headers: []string{langText("参数名", "Parameter Name", "パラメータ名", lang), langText("值", "Value", "値", lang)},
			Rows:    [][]string{},
		}
		for _, p := range params {
			row := []string{p.Name, p.Value.String}
			paramTable.Rows = append(paramTable.Rows, row)
		}
		tables = append(tables, paramTable)
	}
	return cards, tables, nil, nil
}

// processDbinfoModule handles the "dbinfo" inspection item.
func processDbinfoModule(dbConn *sql.DB, lang string, preFetchedInfo *db.FullDBInfo) (cards []ReportCard, tables []*ReportTable, charts []ReportChart, err error) {
	dbInfoToProcess := preFetchedInfo
	var fetchErr error

	if dbInfoToProcess == nil {
		dbInfoToProcess, fetchErr = db.GetDatabaseInfo(dbConn)
		if fetchErr != nil {
			cards = append(cards, ReportCard{Title: langText("错误", "Error", "エラー", lang), Value: fmt.Sprintf(langText("获取数据库信息失败: %v", "Failed to get database info: %v", "データベース情報の取得に失敗しました: %v", lang), fetchErr)})
			return cards, nil, nil, fetchErr
		}
	}

	dbCards := []ReportCard{
		{Title: langText("数据库名", "DB Name", "データベース名", lang), Value: dbInfoToProcess.Database.Name.String},
		{Title: langText("数据库ID", "DBID", "データベースID", lang), Value: formatNullInt64(dbInfoToProcess.Database.DBID)},
		{Title: langText("创建时间", "Created", "作成日時", lang), Value: dbInfoToProcess.Database.Created.String},
		{Title: langText("数据库版本", "Database Version", "データベースバージョン", lang), Value: dbInfoToProcess.Database.OverallVersion},
		{Title: langText("日志模式", "Log Mode", "ログモード", lang), Value: dbInfoToProcess.Database.LogMode},
		{Title: langText("打开模式", "Open Mode", "オープンモード", lang), Value: dbInfoToProcess.Database.OpenMode},
		{Title: langText("是否CDB", "CDB", "CDB", lang), Value: dbInfoToProcess.Database.CDB.String},
		{Title: langText("保护模式", "Protection Mode", "保護モード", lang), Value: dbInfoToProcess.Database.ProtectionMode},
		{Title: langText("闪回", "Flashback", "フラッシュバック", lang), Value: dbInfoToProcess.Database.FlashbackOn},
		{Title: langText("数据库角色", "DB Role", "データベースロール", lang), Value: dbInfoToProcess.Database.DatabaseRole},
		{Title: langText("平台名称", "Platform Name", "プラットフォーム名", lang), Value: dbInfoToProcess.Database.PlatformName},
		{Title: langText("数据库唯一名", "DB Unique Name", "DBユニーク名", lang), Value: dbInfoToProcess.Database.DBUniqueName.String},
		{Title: langText("字符集", "Character Set", "文字セット", lang), Value: dbInfoToProcess.Database.CharacterSet.String},
		{Title: langText("国家字符集", "National Character Set", "各国語文字セット", lang), Value: dbInfoToProcess.Database.NationalCharacterSet.String},
	}
	cards = append(cards, dbCards...)

	if len(dbInfoToProcess.Instances) > 0 {
		instanceTable := &ReportTable{
			Name:    langText("实例信息", "Instance Information", "インスタンス情報", lang),
			Headers: []string{langText("实例ID", "Inst ID", "インスタンスID", lang), langText("实例名", "Instance Name", "インスタンス名", lang), langText("主机名", "Host Name", "ホスト名", lang), langText("版本", "Version", "バージョン", lang), langText("启动时间", "Startup Time", "起動時間", lang), langText("状态", "Status", "ステータス", lang)},
			Rows:    [][]string{},
		}
		for _, inst := range dbInfoToProcess.Instances {
			row := []string{
				strconv.Itoa(inst.InstanceNumber),
				inst.InstanceName,
				inst.HostName,
				inst.Version,
				inst.StartupTime,
				inst.Status,
			}
			instanceTable.Rows = append(instanceTable.Rows, row)
		}
		tables = append(tables, instanceTable)
	}
	return cards, tables, nil, nil
}

// processStorageModule handles the "storage" inspection item.
func processStorageModule(dbConn *sql.DB, lang string) (cards []ReportCard, tables []*ReportTable, charts []ReportChart, err error) {
	storageData, dbErr := db.GetStorageInfo(dbConn)
	if dbErr != nil {
		cards = append(cards, ReportCard{Title: langText("错误", "Error", "エラー", lang), Value: fmt.Sprintf(langText("获取存储信息失败: %v", "Failed to get storage info: %v", "ストレージ情報の取得に失敗しました: %v", lang), dbErr)})
		return cards, nil, nil, dbErr
	}

	// Control Files
	if len(storageData.ControlFiles) > 0 {
		cfTable := &ReportTable{
			Name:    langText("控制文件", "Control Files", "制御ファイル", lang),
			Headers: []string{langText("文件路径", "File Path", "ファイルパス", lang), langText("大小(MB)", "Size(MB)", "サイズ(MB)", lang)},
			Rows:    [][]string{},
		}
		for _, cf := range storageData.ControlFiles {
			row := []string{cf.Name, fmt.Sprintf("%.2f", cf.SizeMB)}
			cfTable.Rows = append(cfTable.Rows, row)
		}
		tables = append(tables, cfTable)
	} else {
		cards = append(cards, ReportCard{Title: langText("控制文件", "Control Files", "制御ファイル", lang), Value: langText("未找到", "Not Found", "見つかりません", lang)})
	}

	// Redo Log Groups
	if len(storageData.RedoLogs) > 0 {
		redoTable := &ReportTable{
			Name:    langText("重做日志组", "Redo Log Groups", "REDOログ・グループ", lang),
			Headers: []string{langText("组号", "Group#", "グループ番号", lang), langText("线程号", "Thread#", "スレッド番号", lang), langText("成员数", "Members", "メンバー数", lang), langText("大小(MB)", "Size(MB)", "サイズ(MB)", lang), langText("成员文件", "Member Files", "メンバーファイル", lang), langText("状态", "Status", "ステータス", lang), langText("已归档", "Archived", "アーカイブ済み", lang), langText("类型", "Type", "タイプ", lang)},
			Rows:    [][]string{},
		}
		for _, rl := range storageData.RedoLogs {
			row := []string{
				strconv.Itoa(rl.GroupNo),
				strconv.Itoa(rl.ThreadNo),
				strconv.Itoa(rl.Members),
				fmt.Sprintf("%.2f", rl.SizeMB),
				rl.Member,
				rl.Status,
				rl.Archived,
				rl.Type,
			}
			redoTable.Rows = append(redoTable.Rows, row)
		}
		tables = append(tables, redoTable)
	} else {
		cards = append(cards, ReportCard{Title: langText("重做日志组", "Redo Log Groups", "REDOログ・グループ", lang), Value: langText("未找到", "Not Found", "見つかりません", lang)})
	}

	// Tablespace Usage
	if len(storageData.Tablespaces) > 0 {
		tsTable := &ReportTable{
			Name:    langText("表空间使用情况", "Tablespace Usage", "表領域使用率", lang),
			Headers: []string{langText("状态", "Status", "ステータス", lang), langText("表空间名", "Tablespace Name", "表領域名", lang), langText("类型", "Type", "タイプ", lang), langText("区管理", "Extent Management", "エクステント管理", lang), langText("段管理", "Segment Management", "セグメント管理", lang), langText("已用(MB)", "Used(MB)", "使用済み(MB)", lang), langText("总计(MB)", "Total(MB)", "合計(MB)", lang), langText("使用率 %", "Used %", "使用率 %", lang), langText("自动扩展大小(MB)", "Autoextend Size(MB)", "自動拡張サイズ(MB)", lang)},
			Rows:    [][]string{},
		}
		for _, ts := range storageData.Tablespaces {
			row := []string{
				ts.Status,
				ts.Name,
				ts.Type,
				ts.ExtentManagement,
				ts.SegmentSpaceManagement,
				fmt.Sprintf("%.2f", ts.UsedMB),
				fmt.Sprintf("%.2f", ts.TotalMB),
				ts.UsedPercent,
				fmt.Sprintf("%.2f", ts.CanExtendMB),
			}
			tsTable.Rows = append(tsTable.Rows, row)
		}
		tables = append(tables, tsTable)
	} else {
		cards = append(cards, ReportCard{Title: langText("表空间使用情况", "Tablespace Usage", "表領域使用率", lang), Value: langText("未找到", "Not Found", "見つかりません", lang)})
	}

	// Data Files
	if len(storageData.DataFiles) > 0 {
		dfTable := &ReportTable{
			Name:    langText("数据文件", "Data Files", "データファイル", lang),
			Headers: []string{langText("文件ID", "File ID", "ファイルID", lang), langText("文件名", "File Name", "ファイル名", lang), langText("表空间", "Tablespace", "表領域", lang), langText("大小(MB)", "Size(MB)", "サイズ(MB)", lang), langText("状态", "Status", "ステータス", lang), langText("自动扩展", "Autoextend", "自動拡張", lang)},
			Rows:    [][]string{},
		}
		for _, df := range storageData.DataFiles {
			row := []string{
				fmt.Sprintf("%d", df.FileID),
				df.FileName,
				df.TablespaceName,
				fmt.Sprintf("%.2f", df.SizeMB),
				df.Status,
				df.Autoextensible,
			}
			dfTable.Rows = append(dfTable.Rows, row)
		}
		tables = append(tables, dfTable)
	} else {
		cards = append(cards, ReportCard{Title: langText("数据文件", "Data Files", "データファイル", lang), Value: langText("未找到", "Not Found", "見つかりません", lang)})
	}

	// Archived Log Summary
	if len(storageData.ArchivedLogsSummary) > 0 {
		archTable := &ReportTable{
			Name:    langText("归档日志摘要 (最近7天)", "Archived Log Summary (Last 7 Days)", "アーカイブREDOログのサマリー (過去7日間)", lang),
			Headers: []string{langText("日期", "Day", "日付", lang), langText("日志计数", "Log Count", "ログ数", lang), langText("总大小(MB)", "Total Size(MB)", "合計サイズ(MB)", lang)},
			Rows:    [][]string{},
		}
		for _, al := range storageData.ArchivedLogsSummary {
			row := []string{
				al.Day,
				strconv.Itoa(al.LogCount),
				fmt.Sprintf("%.2f", al.TotalSizeMB),
			}
			archTable.Rows = append(archTable.Rows, row)
		}
		tables = append(tables, archTable)
	} else {
		cards = append(cards, ReportCard{Title: langText("归档日志摘要 (最近7天)", "Archived Log Summary (Last 7 Days)", "アーカイブREDOログのサマリー (過去7日間)", lang), Value: langText("未找到", "Not Found", "見つかりません", lang)})
	}

	// ASM Disk Groups
	if len(storageData.ASMDiskgroups) > 0 {
		asmTable := &ReportTable{
			Name:    langText("ASM磁盘组", "ASM Disk Groups", "ASMディスク・グループ", lang),
			Headers: []string{langText("磁盘组名", "Diskgroup Name", "ディスクグループ名", lang), langText("总大小(MB)", "Total Size(MB)", "合計サイズ(MB)", lang), langText("空闲(MB)", "Free(MB)", "空き(MB)", lang), langText("使用率 %", "Used %", "使用率 %", lang), langText("状态", "State", "状態", lang), langText("冗余", "Redundancy", "冗長性", lang)},
			Rows:    [][]string{},
		}
		for _, adg := range storageData.ASMDiskgroups {
			row := []string{
				adg.Name,
				strconv.FormatInt(adg.TotalMB, 10),
				strconv.FormatInt(adg.FreeMB, 10),
				fmt.Sprintf("%.2f", adg.UsedPercent),
				adg.State,
				adg.RedundancyType,
			}
			asmTable.Rows = append(asmTable.Rows, row)
		}
		tables = append(tables, asmTable)
	} else {
		cards = append(cards, ReportCard{Title: langText("ASM磁盘组", "ASM Disk Groups", "ASMディスク・グループ", lang), Value: langText("未找到", "Not Found", "見つかりません", lang)})
	}

	// TODO: Add logic for TablespaceGrowth chart if storageData.TablespaceGrowth is not empty
	// TODO: Add logic for DatafileIOStats chart if storageData.DatafileIOStats is not empty

	return cards, tables, charts, nil
}

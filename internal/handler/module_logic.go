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
		cards = append(cards, ReportCard{Title: langText("错误", "Error", lang), Value: fmt.Sprintf(langText("获取参数失败: %v", "Failed to get parameters: %v", lang), dbErr)})
		return cards, nil, nil, dbErr
	}
	logger.Debugf("获取到参数: %v", params)
	if len(params) > 0 {
		paramTable := &ReportTable{
			Name:    langText("参数列表", "Parameter List", lang),
			Headers: []string{langText("参数名", "Parameter Name", lang), langText("值", "Value", lang)},
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
			cards = append(cards, ReportCard{Title: langText("错误", "Error", lang), Value: fmt.Sprintf(langText("获取数据库信息失败: %v", "Failed to get database info: %v", lang), fetchErr)})
			return cards, nil, nil, fetchErr
		}
	}

	dbCards := []ReportCard{
		{Title: langText("数据库名称", "DB Name", lang), Value: dbInfoToProcess.Database.Name.String},
		{Title: langText("DBID", "DBID", lang), Value: formatNullInt64(dbInfoToProcess.Database.DBID)},
		{Title: langText("创建时间", "Created", lang), Value: dbInfoToProcess.Database.Created.String},
		{Title: langText("版本", "Overall Version", lang), Value: dbInfoToProcess.Database.OverallVersion},
		{Title: langText("日志模式", "Log Mode", lang), Value: dbInfoToProcess.Database.LogMode},
		{Title: langText("打开模式", "Open Mode", lang), Value: dbInfoToProcess.Database.OpenMode},
		{Title: langText("容器数据库", "CDB", lang), Value: dbInfoToProcess.Database.CDB.String},
		{Title: langText("保护模式", "Protection Mode", lang), Value: dbInfoToProcess.Database.ProtectionMode},
		{Title: langText("闪回", "Flashback", lang), Value: dbInfoToProcess.Database.FlashbackOn},
		{Title: langText("强制日志", "Force Logging", lang), Value: dbInfoToProcess.Database.ForceLogging.String},
		{Title: langText("角色", "DB Role", lang), Value: dbInfoToProcess.Database.DatabaseRole},
		{Title: langText("平台名称", "Platform Name", lang), Value: dbInfoToProcess.Database.PlatformName},
		{Title: langText("数据库唯一名称", "DB Unique Name", lang), Value: dbInfoToProcess.Database.DBUniqueName.String},
		{Title: langText("字符集", "Character Set", lang), Value: dbInfoToProcess.Database.CharacterSet.String},
		{Title: langText("全国字符集", "National Character Set", lang), Value: dbInfoToProcess.Database.NationalCharacterSet.String},
	}
	cards = append(cards, dbCards...)

	if len(dbInfoToProcess.Instances) > 0 {
		instanceTable := &ReportTable{
			Name:    langText("实例详情", "Instance Details", lang),
			Headers: []string{langText("实例号", "Inst ID", lang), langText("实例名", "Instance Name", lang), langText("主机名", "Host Name", lang), langText("版本", "Version", lang), langText("启动时间", "Startup Time", lang), langText("状态", "Status", lang)},
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
		cards = append(cards, ReportCard{Title: langText("错误", "Error", lang), Value: fmt.Sprintf(langText("获取存储信息失败: %v", "Failed to get storage info: %v", lang), dbErr)})
		return cards, nil, nil, dbErr
	}

	// Control Files
	if len(storageData.ControlFiles) > 0 {
		cfTable := &ReportTable{
			Name:    langText("控制文件", "Control Files", lang),
			Headers: []string{langText("文件路径", "File Path", lang), langText("大小(MB)", "Size(MB)", lang)},
			Rows:    [][]string{},
		}
		for _, cf := range storageData.ControlFiles {
			row := []string{cf.Name, fmt.Sprintf("%.2f", cf.SizeMB)}
			cfTable.Rows = append(cfTable.Rows, row)
		}
		tables = append(tables, cfTable)
	} else {
		cards = append(cards, ReportCard{Title: langText("控制文件", "Control Files", lang), Value: langText("未找到", "Not Found", lang)})
	}

	// Redo Log Groups
	if len(storageData.RedoLogs) > 0 {
		redoTable := &ReportTable{
			Name:    langText("Redo 日志组", "Redo Log Groups", lang),
			Headers: []string{langText("组号", "Group#", lang), langText("线程号", "Thread#", lang), langText("成员数", "Members", lang), langText("大小(MB)", "Size(MB)", lang), langText("文件", "Members", lang), langText("状态", "Status", lang), langText("归档", "Archived", lang), langText("类型", "Type", lang)},
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
		cards = append(cards, ReportCard{Title: langText("Redo 日志组", "Redo Log Groups", lang), Value: langText("未找到", "Not Found", lang)})
	}

	// Tablespace Usage
	if len(storageData.Tablespaces) > 0 {
		tsTable := &ReportTable{
			Name:    langText("表空间使用情况", "Tablespace Usage", lang),
			Headers: []string{langText("状态", "Status", lang), langText("表空间名称", "Tablespace", lang), langText("类型", "Type", lang), langText("extent管理", "Extent Management", lang), langText("segment管理", "Segment Management", lang), langText("已用(MB)", "Used(MB)", lang), langText("总量(MB)", "Total(MB)", lang), langText("使用率(%)", "Used %", lang), langText("可扩展大小(MB)", "Autoextend Size(MB)", lang)},
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
		cards = append(cards, ReportCard{Title: langText("表空间使用情况", "Tablespace Usage", lang), Value: langText("未找到", "Not Found", lang)})
	}

	// Data Files
	if len(storageData.DataFiles) > 0 {
		dfTable := &ReportTable{
			Name:    langText("数据文件", "Data Files", lang),
			Headers: []string{langText("文件编号", "File ID", lang), langText("文件名", "File Name", lang), langText("表空间", "Tablespace", lang), langText("大小(MB)", "Size(MB)", lang), langText("状态", "Status", lang), langText("自动扩展", "Autoextend", lang)},
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
		cards = append(cards, ReportCard{Title: langText("数据文件", "Data Files", lang), Value: langText("未找到", "Not Found", lang)})
	}

	// Archived Log Summary
	if len(storageData.ArchivedLogsSummary) > 0 {
		archTable := &ReportTable{
			Name:    langText("归档日志摘要 (近7日)", "Archived Log Summary (Last 7 Days)", lang),
			Headers: []string{langText("日期", "Day", lang), langText("日志数量", "Log Count", lang), langText("总大小(MB)", "Total Size(MB)", lang)},
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
		cards = append(cards, ReportCard{Title: langText("归档日志摘要 (近7日)", "Archived Log Summary (Last 7 Days)", lang), Value: langText("未找到", "Not Found", lang)})
	}

	// ASM Disk Groups
	if len(storageData.ASMDiskgroups) > 0 {
		asmTable := &ReportTable{
			Name:    langText("ASM 磁盘组", "ASM Diskgroups", lang),
			Headers: []string{langText("磁盘组名称", "Diskgroup", lang), langText("总大小(MB)", "Total(MB)", lang), langText("空闲(MB)", "Free(MB)", lang), langText("使用率(%)", "Used %", lang), langText("状态", "State", lang), langText("冗余类型", "Redundancy", lang)},
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
		cards = append(cards, ReportCard{Title: langText("ASM 磁盘组", "ASM Diskgroups", lang), Value: langText("未找到", "Not Found", lang)})
	}

	// TODO: Add logic for TablespaceGrowth chart if storageData.TablespaceGrowth is not empty
	// TODO: Add logic for DatafileIOStats chart if storageData.DatafileIOStats is not empty

	return cards, tables, charts, nil
}

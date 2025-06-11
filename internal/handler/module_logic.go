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
		cards = append(cards, ReportCard{Title: langText("Error", "Error", "Error", lang), Value: fmt.Sprintf(langText("Failed to get parameters: %v", "Failed to get parameters: %v", "Failed to get parameters: %v", lang), dbErr)})
		return cards, nil, nil, dbErr
	}
	logger.Debugf("Fetched parameters: %v", params)
	if len(params) > 0 {
		paramTable := &ReportTable{
			Name:    langText("Parameter List", "Parameter List", "Parameter List", lang),
			Headers: []string{langText("Parameter Name", "Parameter Name", "Parameter Name", lang), langText("Value", "Value", "Value", lang)},
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
			cards = append(cards, ReportCard{Title: langText("Error", "Error", "Error", lang), Value: fmt.Sprintf(langText("Failed to get database info: %v", "Failed to get database info: %v", "Failed to get database info: %v", lang), fetchErr)})
			return cards, nil, nil, fetchErr
		}
	}

	dbCards := []ReportCard{
		{Title: langText("DB Name", "DB Name", "DB Name", lang), Value: dbInfoToProcess.Database.Name.String},
		{Title: langText("DBID", "DBID", "DBID", lang), Value: formatNullInt64(dbInfoToProcess.Database.DBID)},
		{Title: langText("Created", "Created", "Created", lang), Value: dbInfoToProcess.Database.Created.String},
		{Title: langText("Database Version", "Database Version", "Database Version", lang), Value: dbInfoToProcess.Database.OverallVersion},
		{Title: langText("Log Mode", "Log Mode", "Log Mode", lang), Value: dbInfoToProcess.Database.LogMode},
		{Title: langText("Open Mode", "Open Mode", "Open Mode", lang), Value: dbInfoToProcess.Database.OpenMode},
		{Title: langText("CDB", "CDB", "CDB", lang), Value: dbInfoToProcess.Database.CDB.String},
		{Title: langText("Protection Mode", "Protection Mode", "Protection Mode", lang), Value: dbInfoToProcess.Database.ProtectionMode},
		{Title: langText("Flashback", "Flashback", "Flashback", lang), Value: dbInfoToProcess.Database.FlashbackOn},
		{Title: langText("DB Role", "DB Role", "DB Role", lang), Value: dbInfoToProcess.Database.DatabaseRole},
		{Title: langText("Platform Name", "Platform Name", "Platform Name", lang), Value: dbInfoToProcess.Database.PlatformName},
		{Title: langText("DB Unique Name", "DB Unique Name", "DB Unique Name", lang), Value: dbInfoToProcess.Database.DBUniqueName.String},
		{Title: langText("Character Set", "Character Set", "Character Set", lang), Value: dbInfoToProcess.Database.CharacterSet.String},
		{Title: langText("National Character Set", "National Character Set", "National Character Set", lang), Value: dbInfoToProcess.Database.NationalCharacterSet.String},
	}
	cards = append(cards, dbCards...)

	if len(dbInfoToProcess.Instances) > 0 {
		instanceTable := &ReportTable{
			Name:    langText("Instance Information", "Instance Information", "Instance Information", lang),
			Headers: []string{langText("Inst ID", "Inst ID", "Inst ID", lang), langText("Instance Name", "Instance Name", "Instance Name", lang), langText("Host Name", "Host Name", "Host Name", lang), langText("Version", "Version", "Version", lang), langText("Startup Time", "Startup Time", "Startup Time", lang), langText("Status", "Status", "Status", lang)},
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
		cards = append(cards, ReportCard{Title: langText("Error", "Error", "Error", lang), Value: fmt.Sprintf(langText("Failed to get storage info: %v", "Failed to get storage info: %v", "Failed to get storage info: %v", lang), dbErr)})
		return cards, nil, nil, dbErr
	}

	// Control Files
	if len(storageData.ControlFiles) > 0 {
		cfTable := &ReportTable{
			Name:    langText("Control Files", "Control Files", "Control Files", lang),
			Headers: []string{langText("File Path", "File Path", "File Path", lang), langText("Size(MB)", "Size(MB)", "Size(MB)", lang)},
			Rows:    [][]string{},
		}
		for _, cf := range storageData.ControlFiles {
			row := []string{cf.Name, fmt.Sprintf("%.2f", cf.SizeMB)}
			cfTable.Rows = append(cfTable.Rows, row)
		}
		tables = append(tables, cfTable)
	} else {
		cards = append(cards, ReportCard{Title: langText("Control Files", "Control Files", "Control Files", lang), Value: langText("Not Found", "Not Found", "Not Found", lang)})
	}

	// Redo Log Groups
	if len(storageData.RedoLogs) > 0 {
		redoTable := &ReportTable{
			Name:    langText("Redo Log Groups", "Redo Log Groups", "Redo Log Groups", lang),
			Headers: []string{langText("Group#", "Group#", "Group#", lang), langText("Thread#", "Thread#", "Thread#", lang), langText("Members", "Members", "Members", lang), langText("Size(MB)", "Size(MB)", "Size(MB)", lang), langText("Member Files", "Member Files", "Member Files", lang), langText("Status", "Status", "Status", lang), langText("Archived", "Archived", "Archived", lang), langText("Type", "Type", "Type", lang)},
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
		cards = append(cards, ReportCard{Title: langText("Redo Log Groups", "Redo Log Groups", "Redo Log Groups", lang), Value: langText("Not Found", "Not Found", "Not Found", lang)})
	}

	// Tablespace Usage
	if len(storageData.Tablespaces) > 0 {
		tsTable := &ReportTable{
			Name:    langText("Tablespace Usage", "Tablespace Usage", "Tablespace Usage", lang),
			Headers: []string{langText("Status", "Status", "Status", lang), langText("Tablespace Name", "Tablespace Name", "Tablespace Name", lang), langText("Type", "Type", "Type", lang), langText("Extent Management", "Extent Management", "Extent Management", lang), langText("Segment Management", "Segment Management", "Segment Management", lang), langText("Used(MB)", "Used(MB)", "Used(MB)", lang), langText("Total(MB)", "Total(MB)", "Total(MB)", lang), langText("Used %", "Used %", "Used %", lang), langText("Autoextend Size(MB)", "Autoextend Size(MB)", "Autoextend Size(MB)", lang)},
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
		cards = append(cards, ReportCard{Title: langText("Tablespace Usage", "Tablespace Usage", "Tablespace Usage", lang), Value: langText("Not Found", "Not Found", "Not Found", lang)})
	}

	// Data Files
	if len(storageData.DataFiles) > 0 {
		dfTable := &ReportTable{
			Name:    langText("Data Files", "Data Files", "Data Files", lang),
			Headers: []string{langText("File ID", "File ID", "File ID", lang), langText("File Name", "File Name", "File Name", lang), langText("Tablespace", "Tablespace", "Tablespace", lang), langText("Size(MB)", "Size(MB)", "Size(MB)", lang), langText("Status", "Status", "Status", lang), langText("Autoextend", "Autoextend", "Autoextend", lang)},
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
		cards = append(cards, ReportCard{Title: langText("Data Files", "Data Files", "Data Files", lang), Value: langText("Not Found", "Not Found", "Not Found", lang)})
	}

	// Archived Log Summary
	if len(storageData.ArchivedLogsSummary) > 0 {
		archTable := &ReportTable{
			Name:    langText("Archived Log Summary (Last 7 Days)", "Archived Log Summary (Last 7 Days)", "Archived Log Summary (Last 7 Days)", lang),
			Headers: []string{langText("Day", "Day", "Day", lang), langText("Log Count", "Log Count", "Log Count", lang), langText("Total Size(MB)", "Total Size(MB)", "Total Size(MB)", lang)},
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
		cards = append(cards, ReportCard{Title: langText("Archived Log Summary (Last 7 Days)", "Archived Log Summary (Last 7 Days)", "Archived Log Summary (Last 7 Days)", lang), Value: langText("Not Found", "Not Found", "Not Found", lang)})
	}

	// ASM Disk Groups
	if len(storageData.ASMDiskgroups) > 0 {
		asmTable := &ReportTable{
			Name:    langText("ASM Disk Groups", "ASM Disk Groups", "ASM Disk Groups", lang),
			Headers: []string{langText("Diskgroup Name", "Diskgroup Name", "Diskgroup Name", lang), langText("Total Size(MB)", "Total Size(MB)", "Total Size(MB)", lang), langText("Free(MB)", "Free(MB)", "Free(MB)", lang), langText("Used %", "Used %", "Used %", lang), langText("State", "State", "State", lang), langText("Redundancy", "Redundancy", "Redundancy", lang)},
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
		cards = append(cards, ReportCard{Title: langText("ASM Disk Groups", "ASM Disk Groups", "ASM Disk Groups", lang), Value: langText("Not Found", "Not Found", "Not Found", lang)})
	}

	// TODO: Add logic for TablespaceGrowth chart if storageData.TablespaceGrowth is not empty
	// TODO: Add logic for DatafileIOStats chart if storageData.DatafileIOStats is not empty

	return cards, tables, charts, nil
}

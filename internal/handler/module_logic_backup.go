package handler

import (
	"database/sql"
	"fmt"

	"github.com/goodwaysIT/inspect4oracle/internal/db"
	"github.com/goodwaysIT/inspect4oracle/internal/logger"
)

// generateArchivelogModeCard generates a report card for the archivelog mode.
func generateArchivelogModeCard(backupData *db.AllBackupInfo, lang string) (card ReportCard, err error) {
	if backupData.ArchivelogModeError != nil {
		logger.Errorf("获取归档模式失败: %v", backupData.ArchivelogModeError)
		return cardFromError("归档模式错误", "Archivelog Mode Error", backupData.ArchivelogModeError, lang), backupData.ArchivelogModeError
	}
	return ReportCard{
		Title: langText("数据库日志模式", "Database Log Mode", lang),
		Value: backupData.ArchivelogMode.LogMode,
	}, nil
}

// generateFlashbackStatusCard generates a report card for the flashback database status.
func generateFlashbackStatusCard(backupData *db.AllBackupInfo, lang string) (card ReportCard, err error) {
	if backupData.FlashbackStatusError != nil {
		logger.Errorf("获取闪回状态失败: %v", backupData.FlashbackStatusError)
		return cardFromError("闪回状态错误", "Flashback Status Error", backupData.FlashbackStatusError, lang), backupData.FlashbackStatusError
	}

	flashbackCardValue := fmt.Sprintf("%s", backupData.FlashbackStatus.FlashbackOn)
	if backupData.FlashbackStatus.FlashbackOn == "YES" {
		if backupData.FlashbackStatus.OldestFlashbackTime.Valid {
			flashbackCardValue += fmt.Sprintf(langText(" (最早可至: %s)", " (Oldest: %s)", lang), backupData.FlashbackStatus.OldestFlashbackTime.Time.Format("2006-01-02 15:04:05"))
		}
		if backupData.FlashbackStatus.RetentionTarget.Valid {
			flashbackCardValue += fmt.Sprintf(langText(", 保留目标: %d 分钟", ", Retention: %d mins", lang), backupData.FlashbackStatus.RetentionTarget.Int64)
		}
	}
	return ReportCard{
		Title: langText("闪回数据库状态", "Flashback Database Status", lang),
		Value: flashbackCardValue,
	}, nil
}

// generateRMANJobsTable generates a report table for RMAN backup jobs or a card if no data/error.
func generateRMANJobsTable(backupData *db.AllBackupInfo, lang string) (card *ReportCard, table *ReportTable, err error) {
	if backupData.RMANJobsError != nil {
		logger.Errorf("获取RMAN作业失败: %v", backupData.RMANJobsError)
		noDataCard := cardFromError("RMAN备份作业错误", "RMAN Backup Jobs Error", backupData.RMANJobsError, lang)
		return &noDataCard, nil, backupData.RMANJobsError
	}

	if len(backupData.RMANJobs) > 0 {
		rmanTable := &ReportTable{
			Name: langText("最近RMAN备份作业 (过去7天)", "Recent RMAN Backup Jobs (Last 7 Days)", lang),
			Headers: []string{
				langText("会话键", "Session Key", lang), langText("开始时间", "Start Time", lang), langText("结束时间", "End Time", lang),
				langText("状态", "Status", lang), langText("输入", "Input", lang), langText("输出", "Output", lang), langText("耗时", "Duration", lang),
				langText("优化?", "Optimized?", lang), langText("压缩率", "Compression Ratio", lang),
			},
			Rows: [][]string{},
		}
		for _, job := range backupData.RMANJobs {
			startTime := "N/A"
			if job.StartTime.Valid {
				startTime = job.StartTime.Time.Format("2006-01-02 15:04:05")
			}
			endTime := "N/A"
			if job.EndTime.Valid {
				endTime = job.EndTime.Time.Format("2006-01-02 15:04:05")
			}
			row := []string{
				fmt.Sprintf("%d", job.SessionKey),
				startTime,
				endTime,
				job.Status,
				job.InputBytesDisplay,
				job.OutputBytesDisplay,
				job.TimeTakenDisplay,
				job.Optimized,
				fmt.Sprintf("%.2f", job.CompressionRatio),
			}
			rmanTable.Rows = append(rmanTable.Rows, row)
		}
		table = rmanTable
	} else {
		noDataCard := ReportCard{
			Title: langText("RMAN备份作业", "RMAN Backup Jobs", lang),
			Value: langText("过去7天内未发现RMAN备份作业记录。", "No RMAN backup jobs found in the last 7 days.", lang),
		}
		card = &noDataCard
	}
	return card, table, nil
}

// generateRecycleBinTable generates a report table for recycle bin objects or a card if no data/error.
func generateRecycleBinTable(backupData *db.AllBackupInfo, lang string) (card *ReportCard, table *ReportTable, err error) {
	if backupData.RecycleBinError != nil {
		logger.Errorf("获取回收站对象失败: %v", backupData.RecycleBinError)
		noDataCard := cardFromError("回收站错误", "Recycle Bin Error", backupData.RecycleBinError, lang)
		return &noDataCard, nil, backupData.RecycleBinError
	}

	if len(backupData.RecycleBinItems) > 0 {
		rbTable := &ReportTable{
			Name: langText("回收站对象 (可恢复)", "Recycle Bin Objects (Restorable)", lang),
			Headers: []string{
				langText("所有者", "Owner", lang), langText("对象名", "Object Name", lang), langText("原始名", "Original Name", lang),
				langText("类型", "Type", lang), langText("删除时间", "Drop Time", lang), langText("空间(块)", "Space (Blocks)", lang), langText("可恢复?", "Can Undrop?", lang),
			},
			Rows: [][]string{},
		}
		for _, item := range backupData.RecycleBinItems {
			row := []string{
				item.Owner,
				item.ObjectName,
				item.OriginalName,
				item.Type,
				item.Droptime.String, // Assuming Droptime is sql.NullString or similar
				fmt.Sprintf("%d", item.Space),
				item.CanUndrop,
			}
			rbTable.Rows = append(rbTable.Rows, row)
		}
		table = rbTable
	} else {
		noDataCard := ReportCard{
			Title: langText("回收站", "Recycle Bin", lang),
			Value: langText("回收站中未发现可恢复的对象。", "No restorable objects found in the recycle bin.", lang),
		}
		card = &noDataCard
	}
	return card, table, nil
}

// generateDataPumpJobsTable generates a report table for Data Pump jobs or a card if no data/error.
func generateDataPumpJobsTable(backupData *db.AllBackupInfo, lang string) (card *ReportCard, table *ReportTable, err error) {
	if backupData.DataPumpJobsError != nil {
		logger.Errorf("获取Data Pump作业失败: %v", backupData.DataPumpJobsError)
		noDataCard := cardFromError("Data Pump作业错误", "Data Pump Jobs Error", backupData.DataPumpJobsError, lang)
		return &noDataCard, nil, backupData.DataPumpJobsError
	}

	if len(backupData.DataPumpJobs) > 0 {
		dpTable := &ReportTable{
			Name: langText("Data Pump 作业", "Data Pump Jobs", lang),
			Headers: []string{
				langText("作业名", "Job Name", lang), langText("所有者", "Owner", lang), langText("操作", "Operation", lang),
				langText("模式", "Mode", lang), langText("状态", "State", lang), langText("附加会话", "Attached Sessions", lang),
			},
			Rows: [][]string{},
		}
		for _, job := range backupData.DataPumpJobs {
			attachedSessions := "N/A"
			if job.AttachedSessions.Valid {
				attachedSessions = fmt.Sprintf("%d", job.AttachedSessions.Int64)
			}
			row := []string{
				job.JobName,
				job.OwnerName,
				job.Operation,
				job.JobMode,
				job.State,
				attachedSessions,
			}
			dpTable.Rows = append(dpTable.Rows, row)
		}
		table = dpTable
	} else {
		noDataCard := ReportCard{
			Title: langText("Data Pump 作业", "Data Pump Jobs", lang),
			Value: langText("未发现活动的或最近的Data Pump作业。", "No active or recent Data Pump jobs found.", lang),
		}
		card = &noDataCard
	}
	return card, table, nil
}

// processBackupModule 处理 "backup" 巡检项
func processBackupModule(dbConn *sql.DB, lang string) (allCards []ReportCard, allTables []*ReportTable, charts []ReportChart, overallErr error) {
	logger.Infof("开始处理备份模块... 语言: %s", lang)

	backupData := db.GetAllBackupDetails(dbConn) // backupData is of type db.AllBackupInfo

	// If there's an error getting ArchivelogMode, it might indicate a broader issue with DB access for backup info.
	if backupData.ArchivelogModeError != nil {
		logger.Errorf("处理备份模块 - 获取基础备份信息(如归档模式)失败: %v", backupData.ArchivelogModeError)
		allCards = append(allCards, cardFromError("备份信息错误", "Backup Information Error", backupData.ArchivelogModeError, lang))
		return allCards, allTables, nil, backupData.ArchivelogModeError // No charts for backup module
	}

	// Helper to manage overall error, ensuring we capture the first non-nil error.
	// And to correctly wrap multiple errors if they occur.
	appendErr := func(currentOverallErr error, newErr error) error {
		if newErr == nil {
			return currentOverallErr
		}
		if currentOverallErr == nil {
			return newErr
		}
		return fmt.Errorf("%v; %w", currentOverallErr, newErr)
	}

	// 1. Archivelog Mode
	archivelogCard, err := generateArchivelogModeCard(&backupData, lang) // Pass address of backupData
	allCards = append(allCards, archivelogCard)
	overallErr = appendErr(overallErr, err)

	// 2. Flashback Status
	flashbackCard, err := generateFlashbackStatusCard(&backupData, lang) // Pass address of backupData
	allCards = append(allCards, flashbackCard)
	overallErr = appendErr(overallErr, err)

	// 3. RMAN Jobs
	rmanCard, rmanTable, err := generateRMANJobsTable(&backupData, lang) // Pass address of backupData
	if rmanCard != nil {
		allCards = append(allCards, *rmanCard)
	}
	if rmanTable != nil {
		allTables = append(allTables, rmanTable)
	}
	overallErr = appendErr(overallErr, err)

	// 4. Recycle Bin
	rbCard, rbTable, err := generateRecycleBinTable(&backupData, lang) // Pass address of backupData
	if rbCard != nil {
		allCards = append(allCards, *rbCard)
	}
	if rbTable != nil {
		allTables = append(allTables, rbTable)
	}
	overallErr = appendErr(overallErr, err)

	// 5. Data Pump Jobs
	dpCard, dpTable, err := generateDataPumpJobsTable(&backupData, lang) // Pass address of backupData
	if dpCard != nil {
		allCards = append(allCards, *dpCard)
	}
	if dpTable != nil {
		allTables = append(allTables, dpTable)
	}
	overallErr = appendErr(overallErr, err)

	return allCards, allTables, nil, overallErr // No charts for backup module
}

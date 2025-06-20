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
		logger.Errorf("Failed to get archivelog mode: %v", backupData.ArchivelogModeError)
		return cardFromError("归档日志模式错误", "Archivelog Mode Error", "アーカイブログモードエラー", backupData.ArchivelogModeError, lang), backupData.ArchivelogModeError
	}
	return ReportCard{
		Title: langText("数据库日志模式", "Database Log Mode", "データベースログモード", lang),
		Value: backupData.ArchivelogMode.LogMode,
	}, nil
}

// generateFlashbackStatusCard generates a report card for the flashback database status.
func generateFlashbackStatusCard(backupData *db.AllBackupInfo, lang string) (card ReportCard, err error) {
	if backupData.FlashbackStatusError != nil {
		logger.Errorf("Failed to get flashback status: %v", backupData.FlashbackStatusError)
		return cardFromError("闪回状态错误", "Flashback Status Error", "フラッシュバックステータスエラー", backupData.FlashbackStatusError, lang), backupData.FlashbackStatusError
	}

	flashbackCardValue := fmt.Sprintf("%s", backupData.FlashbackStatus.FlashbackOn)
	if backupData.FlashbackStatus.FlashbackOn == "YES" {
		if backupData.FlashbackStatus.OldestFlashbackTime.Valid {
			flashbackCardValue += fmt.Sprintf(langText(" (最早可至: %s)", " (Oldest: %s)", " (最古の: %s)", lang), backupData.FlashbackStatus.OldestFlashbackTime.Time.Format("2006-01-02 15:04:05"))
		}
		if backupData.FlashbackStatus.RetentionTarget.Valid {
			flashbackCardValue += fmt.Sprintf(langText(", 保留目标: %d 分钟", ", Retention: %d mins", ", 保持期間: %d 分", lang), backupData.FlashbackStatus.RetentionTarget.Int64)
		}
	}
	return ReportCard{
		Title: langText("闪回数据库状态", "Flashback Database Status", "フラッシュバックデータベースステータス", lang),
		Value: flashbackCardValue,
	}, nil
}

// generateRMANJobsTable generates a report table for RMAN backup jobs or a card if no data/error.
func generateRMANJobsTable(backupData *db.AllBackupInfo, lang string) (card *ReportCard, table *ReportTable, err error) {
	if backupData.RMANJobsError != nil {
		logger.Errorf("Failed to get RMAN jobs: %v", backupData.RMANJobsError)
		noDataCard := cardFromError("RMAN备份作业错误", "RMAN Backup Jobs Error", "RMANバックアップジョブエラー", backupData.RMANJobsError, lang)
		return &noDataCard, nil, backupData.RMANJobsError
	}

	if len(backupData.RMANJobs) > 0 {
		rmanTable := &ReportTable{
			Name: langText("最近RMAN备份作业 (过去7天)", "Recent RMAN Backup Jobs (Last 7 Days)", "最近のRMANバックアップジョブ (過去7日間)", lang),
			Headers: []string{
				langText("会话键", "Session Key", "セッションキー", lang), langText("开始时间", "Start Time", "開始時間", lang), langText("结束时间", "End Time", "終了時間", lang),
				langText("状态", "Status", "ステータス", lang), langText("输入", "Input", "入力", lang), langText("输出", "Output", "出力", lang), langText("耗时", "Duration", "所要時間", lang),
				langText("优化?", "Optimized?", "最適化?", lang), langText("压缩率", "Compression Ratio", "圧縮率", lang),
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
			Title: langText("RMAN备份作业", "RMAN Backup Jobs", "RMANバックアップジョブ", lang),
			Value: langText("在过去7天内未找到RMAN备份作业。注意：RMAN备份作业数据来自过去7天。如果V$RMAN_BACKUP_JOB_DETAILS为空，则会尝试从V$BACKUP_SET获取数据，其中可能不包含所有作业详细信息。", "No RMAN backup jobs found in the last 7 days. Note: RMAN backup job data is from the last 7 days. If V$RMAN_BACKUP_JOB_DETAILS is empty, data is attempted from V$BACKUP_SET, which may not include all job details.", "過去7日間でRMANバックアップジョブが見つかりませんでした。注意：RMANバックアップジョブデータは過去7日間のものです。V$RMAN_BACKUP_JOB_DETAILSが空の場合、データはV$BACKUP_SETから試行されますが、これにはすべてのジョブ詳細が含まれていない場合があります。", lang),
		}
		card = &noDataCard
	}
	return card, table, nil
}

// generateRecycleBinTable generates a report table for recycle bin objects or a card if no data/error.
func generateRecycleBinTable(backupData *db.AllBackupInfo, lang string) (card *ReportCard, table *ReportTable, err error) {
	if backupData.RecycleBinError != nil {
		logger.Errorf("Failed to get recycle bin objects: %v", backupData.RecycleBinError)
		noDataCard := cardFromError("回收站错误", "Recycle Bin Error", "リサイクルビンエラー", backupData.RecycleBinError, lang)
		return &noDataCard, nil, backupData.RecycleBinError
	}

	if len(backupData.RecycleBinItems) > 0 {
		rbTable := &ReportTable{
			Name: langText("回收站对象 (可恢复)", "Recycle Bin Objects (Restorable)", "リサイクルビンオブジェクト (復元可能)", lang),
			Headers: []string{
				langText("所有者", "Owner", "所有者", lang), langText("对象名称", "Object Name", "オブジェクト名", lang), langText("原始名称", "Original Name", "元の名前", lang),
				langText("类型", "Type", "タイプ", lang), langText("删除时间", "Drop Time", "削除時間", lang), langText("空间(块)", "Space (Blocks)", "スペース(ブロック)", lang), langText("可恢复?", "Can Undrop?", "復元可能?", lang),
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
			Title: langText("回收站", "Recycle Bin", "リサイクルビン", lang),
			Value: langText("回收站中未发现可恢复的对象。", "No restorable objects found in the recycle bin.", "リサイクルビンに復元可能なオブジェクトが見つかりませんでした。", lang),
		}
		card = &noDataCard
	}
	return card, table, nil
}

// generateDataPumpJobsTable generates a report table for Data Pump jobs or a card if no data/error.
func generateDataPumpJobsTable(backupData *db.AllBackupInfo, lang string) (card *ReportCard, table *ReportTable, err error) {
	if backupData.DataPumpJobsError != nil {
		logger.Errorf("Failed to get Data Pump jobs: %v", backupData.DataPumpJobsError)
		noDataCard := cardFromError("数据泵作业错误", "Data Pump Jobs Error", "データポンプジョブエラー", backupData.DataPumpJobsError, lang)
		return &noDataCard, nil, backupData.DataPumpJobsError
	}

	if len(backupData.DataPumpJobs) > 0 {
		dpTable := &ReportTable{
			Name: langText("Data Pump 作业", "Data Pump Jobs", "データポンプジョブ", lang),
			Headers: []string{
				langText("作业名称", "Job Name", "ジョブ名", lang), langText("所有者", "Owner", "所有者", lang), langText("操作", "Operation", "操作", lang),
				langText("模式", "Mode", "モード", lang), langText("状态", "State", "ステータス", lang), langText("附加会话", "Attached Sessions", "アタッチセッション", lang),
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
			Title: langText("Data Pump 作业", "Data Pump Jobs", "データポンプジョブ", lang),
			Value: langText("未发现活动的或最近的Data Pump作业。", "No active or recent Data Pump jobs found.", "アクティブまたは最近のデータポンプジョブが見つかりませんでした。", lang),
		}
		card = &noDataCard
	}
	return card, table, nil
}

// processBackupModule handles the "backup" inspection item.
func processBackupModule(dbConn *sql.DB, lang string) (allCards []ReportCard, allTables []*ReportTable, charts []ReportChart, overallErr error) {
	logger.Infof("Starting to process backup module... Language: %s", lang)

	backupData := db.GetAllBackupDetails(dbConn) // backupData is of type db.AllBackupInfo

	// If there's an error getting ArchivelogMode, it might indicate a broader issue with DB access for backup info.
	if backupData.ArchivelogModeError != nil {
		logger.Errorf("Error processing backup module - failed to get basic backup information (e.g., archivelog mode): %v", backupData.ArchivelogModeError)
		allCards = append(allCards, cardFromError("备份信息错误", "Backup Information Error", "バックアップ情報エラー", backupData.ArchivelogModeError, lang))
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

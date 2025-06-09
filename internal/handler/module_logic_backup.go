package handler

import (
	"database/sql"
	"fmt"

	"github.com/goodwaysIT/inspect4oracle/internal/db"
	"github.com/goodwaysIT/inspect4oracle/internal/logger"
)

// processBackupModule 处理 "backup" 巡检项
func processBackupModule(dbConn *sql.DB, lang string) (cards []ReportCard, tables []*ReportTable, charts []ReportChart, overallErr error) {
	logger.Infof("开始处理备份模块... 语言: %s", lang)

	backupData := db.GetAllBackupDetails(dbConn)
	if backupData.ArchivelogModeError != nil {
		logger.Errorf("处理备份模块 - 获取所有备份详情失败: %v", backupData.ArchivelogModeError)
		cards = append(cards, cardFromError("备份信息错误", "Backup Information Error", backupData.ArchivelogModeError, lang))
		return cards, tables, charts, backupData.ArchivelogModeError
	}

	// 1. 数据库日志模式 (归档模式)
	if backupData.ArchivelogModeError != nil {
		logger.Errorf("处理备份模块 - 获取归档模式失败: %v", backupData.ArchivelogModeError)
		cards = append(cards, cardFromError("归档模式错误", "Archivelog Mode Error", backupData.ArchivelogModeError, lang))
		if overallErr == nil {
			overallErr = backupData.ArchivelogModeError
		}
	} else {
		cards = append(cards, ReportCard{
			Title: langText("数据库日志模式", "Database Log Mode", lang),
			Value: backupData.ArchivelogMode.LogMode,
		})
	}

	// 2. 闪回数据库状态
	if backupData.FlashbackStatusError != nil {
		logger.Errorf("处理备份模块 - 获取闪回状态失败: %v", backupData.FlashbackStatusError)
		cards = append(cards, cardFromError("闪回状态错误", "Flashback Status Error", backupData.FlashbackStatusError, lang))
		if overallErr == nil {
			overallErr = backupData.FlashbackStatusError
		} else {
			overallErr = fmt.Errorf("%v; %w", overallErr, backupData.FlashbackStatusError)
		}
	} else {
		flashbackCardValue := fmt.Sprintf("%s", backupData.FlashbackStatus.FlashbackOn)
		if backupData.FlashbackStatus.FlashbackOn == "YES" {
			if backupData.FlashbackStatus.OldestFlashbackTime.Valid {
				flashbackCardValue += fmt.Sprintf(langText(" (最早可至: %s)", " (Oldest: %s)", lang), backupData.FlashbackStatus.OldestFlashbackTime.Time.Format("2006-01-02 15:04:05"))
			}
			if backupData.FlashbackStatus.RetentionTarget.Valid {
				flashbackCardValue += fmt.Sprintf(langText(", 保留目标: %d 分钟", ", Retention: %d mins", lang), backupData.FlashbackStatus.RetentionTarget.Int64)
			}
		}
		cards = append(cards, ReportCard{
			Title: langText("闪回数据库状态", "Flashback Database Status", lang),
			Value: flashbackCardValue,
		})
	}

	// 3. RMAN 备份作业 (过去7天)
	if backupData.RMANJobsError != nil {
		logger.Errorf("处理备份模块 - 获取RMAN作业失败: %v", backupData.RMANJobsError)
		cards = append(cards, cardFromError("RMAN备份作业错误", "RMAN Backup Jobs Error", backupData.RMANJobsError, lang))
		if overallErr == nil {
			overallErr = backupData.RMANJobsError
		} else {
			overallErr = fmt.Errorf("%v; %w", overallErr, backupData.RMANJobsError)
		}
	} else {
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
			tables = append(tables, rmanTable)
		} else {
			cards = append(cards, ReportCard{
				Title: langText("RMAN备份作业", "RMAN Backup Jobs", lang),
				Value: langText("过去7天内未发现RMAN备份作业记录。", "No RMAN backup jobs found in the last 7 days.", lang),
			})
		}
	}

	// 4. 回收站对象
	if backupData.RecycleBinError != nil {
		logger.Errorf("处理备份模块 - 获取回收站对象失败: %v", backupData.RecycleBinError)
		cards = append(cards, cardFromError("回收站错误", "Recycle Bin Error", backupData.RecycleBinError, lang))
		if overallErr == nil {
			overallErr = backupData.RecycleBinError
		} else {
			overallErr = fmt.Errorf("%v; %w", overallErr, backupData.RecycleBinError)
		}
	} else {
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
			tables = append(tables, rbTable)
		} else {
			cards = append(cards, ReportCard{
				Title: langText("回收站", "Recycle Bin", lang),
				Value: langText("回收站中未发现可恢复的对象。", "No restorable objects found in the recycle bin.", lang),
			})
		}
	}

	// 5. Data Pump 作业
	if backupData.DataPumpJobsError != nil {
		logger.Errorf("处理备份模块 - 获取Data Pump作业失败: %v", backupData.DataPumpJobsError)
		cards = append(cards, cardFromError("Data Pump作业错误", "Data Pump Jobs Error", backupData.DataPumpJobsError, lang))
		if overallErr == nil {
			overallErr = backupData.DataPumpJobsError
		} else {
			overallErr = fmt.Errorf("%v; %w", overallErr, backupData.DataPumpJobsError)
		}
	} else {
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
			tables = append(tables, dpTable)
		} else {
			cards = append(cards, ReportCard{
				Title: langText("Data Pump 作业", "Data Pump Jobs", lang),
				Value: langText("未发现活动的或最近的Data Pump作业。", "No active or recent Data Pump jobs found.", lang),
			})
		}
	}

	charts = nil // Backup module does not have charts
	return cards, tables, charts, overallErr
}

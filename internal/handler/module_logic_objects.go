package handler

import (
	"database/sql"
	"fmt"

	"github.com/goodwaysIT/inspect4oracle/internal/db"
	"github.com/goodwaysIT/inspect4oracle/internal/logger"
)

// processObjectsModule handles the "objects" inspection item.
func appendError(existingErr, newErr error) error {
	if newErr == nil {
		return existingErr
	}
	if existingErr == nil {
		return newErr
	}
	return fmt.Errorf("%v; %w", existingErr, newErr)
}

func processObjectsModule(dbConn *sql.DB, lang string) (cards []ReportCard, tables []*ReportTable, charts []ReportChart, overallErr error) {
	logger.Infof("开始处理对象模块... 语言: %s", lang)

	allDbObjectInfo, overviewErr, topSegmentsErr, invalidObjectsErr := db.GetObjectDetails(dbConn)

	// 1. 处理对象类型统计
	if overviewErr != nil {
		logger.Errorf("处理对象模块 - 获取对象类型统计失败: %v", overviewErr)
		cards = append(cards, ReportCard{
			Title: langText("对象类型统计错误", "Object Type Count Error", lang),
			Value: fmt.Sprintf(langText("获取对象类型统计失败: %v", "Failed to get object type counts: %v", lang), overviewErr),
		})
		overallErr = appendError(overallErr, overviewErr)
	} else if allDbObjectInfo != nil && len(allDbObjectInfo.Overview) > 0 {
		objCountTable := &ReportTable{
			Name:    langText("对象类型统计", "Object Type Counts", lang),
			Headers: []string{langText("所有者", "Owner", lang), langText("对象类型", "Object Type", lang), langText("数量", "Count", lang)}, // Added Owner header
			Rows:    [][]string{},
		}
		for _, oc := range allDbObjectInfo.Overview {
			row := []string{oc.Owner, oc.ObjectType, fmt.Sprintf("%d", oc.ObjectCount)} // Added oc.Owner, changed oc.Count to oc.ObjectCount
			objCountTable.Rows = append(objCountTable.Rows, row)
		}
		tables = append(tables, objCountTable)
	} else {
		cards = append(cards, ReportCard{
			Title: langText("对象类型统计", "Object Type Counts", lang),
			Value: langText("未查询到对象类型统计数据。", "No object type count data found.", lang),
		})
	}

	// 2. 处理 Top 段信息
	if topSegmentsErr != nil {
		logger.Errorf("处理对象模块 - 获取 Top 段信息失败: %v", topSegmentsErr)
		cards = append(cards, ReportCard{
			Title: langText("Top段错误", "Top Segments Error", lang),
			Value: fmt.Sprintf(langText("获取 Top 段信息失败: %v", "Failed to get top segments: %v", lang), topSegmentsErr),
		})
		overallErr = appendError(overallErr, topSegmentsErr)
	} else if allDbObjectInfo != nil && len(allDbObjectInfo.TopSegments) > 0 {
		topSegmentsTable := &ReportTable{
			Name:    langText("Top 段 (按大小)", "Top Segments (by Size)", lang),
			// TablespaceName is not available in db.TopSegment struct, so it's removed from headers.
			Headers: []string{langText("所有者", "Owner", lang), langText("段名", "Segment Name", lang), langText("段类型", "Segment Type", lang), langText("大小 (GB)", "Size (GB)", lang)},
			Rows:    [][]string{},
		}
		for _, ts := range allDbObjectInfo.TopSegments {
			sizeGB := ts.SizeMB / 1024 // Convert MB to GB
			// ts.TablespaceName is not available
			row := []string{ts.Owner, ts.SegmentName, ts.SegmentType, fmt.Sprintf("%.2f", sizeGB)}
			topSegmentsTable.Rows = append(topSegmentsTable.Rows, row)
		}
		tables = append(tables, topSegmentsTable)
	} else {
		cards = append(cards, ReportCard{
			Title: langText("Top段", "Top Segments", lang),
			Value: langText("未查询到Top段数据。", "No top segments data found.", lang),
		})
	}

	// 3. 处理无效对象列表
	if invalidObjectsErr != nil {
		logger.Errorf("处理对象模块 - 获取无效对象列表失败: %v", invalidObjectsErr)
		cards = append(cards, ReportCard{
			Title: langText("无效对象列表错误", "Invalid Objects List Error", lang),
			Value: fmt.Sprintf(langText("获取无效对象列表失败: %v", "Failed to get invalid objects list: %v", lang), invalidObjectsErr),
		})
		overallErr = appendError(overallErr, invalidObjectsErr)
	} else if allDbObjectInfo != nil && len(allDbObjectInfo.InvalidObjects) > 0 {
		invalidObjectsTable := &ReportTable{
			Name:    langText("无效对象列表", "Invalid Objects List", lang),
			Headers: []string{langText("所有者", "Owner", lang), langText("对象名", "Object Name", lang), langText("对象类型", "Object Type", lang), langText("创建时间", "Created", lang), langText("最后DDL时间", "Last DDL Time", lang)},
			Rows:    [][]string{},
		}
		for _, obj := range allDbObjectInfo.InvalidObjects {
			row := []string{obj.Owner, obj.ObjectName, obj.ObjectType, obj.Created, obj.LastDDLTime}
			invalidObjectsTable.Rows = append(invalidObjectsTable.Rows, row)
		}
		tables = append(tables, invalidObjectsTable)
	} else {
		cards = append(cards, ReportCard{
			Title: langText("无效对象检查", "Invalid Objects Check", lang),
			Value: langText("未发现无效对象。", "No invalid objects found.", lang),
		})
	}

	charts = nil // Objects module does not have charts
	return cards, tables, charts, overallErr
}

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
	logger.Infof("Starting to process objects module... Language: %s", lang)

	allDbObjectInfo, overviewErr, topSegmentsErr, invalidObjectsErr := db.GetObjectDetails(dbConn)

	// 1. Process object type statistics
	if overviewErr != nil {
		logger.Errorf("Error processing objects module - failed to get object type statistics: %v", overviewErr)
		cards = append(cards, ReportCard{
			Title: langText("对象类型计数错误", "Object Type Count Error", "オブジェクトタイプカウントエラー", lang),
			Value: fmt.Sprintf(langText("获取对象类型计数失败: %v", "Failed to get object type counts: %v", "オブジェクトタイプ数の取得に失敗しました: %v", lang), overviewErr),
		})
		overallErr = appendError(overallErr, overviewErr)
	} else if allDbObjectInfo != nil && len(allDbObjectInfo.Overview) > 0 {
		objCountTable := &ReportTable{
			Name:    langText("对象类型统计", "Object Type Counts", "オブジェクトタイプ統計", lang),
			Headers: []string{langText("所有者", "Owner", "所有者", lang), langText("对象类型", "Object Type", "オブジェクトタイプ", lang), langText("数量", "Count", "数量", lang)}, // Added Owner header
			Rows:    [][]string{},
		}
		for _, oc := range allDbObjectInfo.Overview {
			row := []string{oc.Owner, oc.ObjectType, fmt.Sprintf("%d", oc.ObjectCount)} 
			objCountTable.Rows = append(objCountTable.Rows, row)
		}
		tables = append(tables, objCountTable)
	} else {
		cards = append(cards, ReportCard{
			Title: langText("对象类型统计", "Object Type Statistics", "オブジェクトタイプ統計", lang),
			Value: langText("没有可用的对象类型统计数据。", "No object type statistics data available.", "利用可能なオブジェクトタイプ統計データがありません。", lang),
		})
	}

	// 2. Process Top Segments information
	if topSegmentsErr != nil {
		logger.Errorf("Error processing objects module - failed to get Top Segments information: %v", topSegmentsErr)
		cards = append(cards, ReportCard{
			Title: langText("按大小排列的热点段错误", "Top Segments by Size Error", "サイズ別トップセグメントエラー", lang),
			Value: fmt.Sprintf(langText("按大小获取热点段失败: %v", "Failed to get top segments by size: %v", "サイズ順のトップセグメントの取得に失敗しました: %v", lang), topSegmentsErr),
		})
		overallErr = appendError(overallErr, topSegmentsErr)
	} else if allDbObjectInfo != nil && len(allDbObjectInfo.TopSegments) > 0 {
		topSegmentsTable := &ReportTable{
			Name:    langText("按大小排列的热点段", "Top Segments by Size", "サイズ別トップセグメント", lang),
			Headers: []string{langText("所有者", "Owner", "所有者", lang), langText("段名", "Segment Name", "セグメント名", lang), langText("段类型", "Segment Type", "セグメントタイプ", lang), langText("大小(GB)", "Size (GB)", "サイズ(GB)", lang)},
			Rows:    [][]string{},
		}
		for _, ts := range allDbObjectInfo.TopSegments {
			sizeGB := ts.SizeMB / 1024 
			row := []string{ts.Owner, ts.SegmentName, ts.SegmentType, fmt.Sprintf("%.2f", sizeGB)}
			topSegmentsTable.Rows = append(topSegmentsTable.Rows, row)
		}
		tables = append(tables, topSegmentsTable)
	} else {
		cards = append(cards, ReportCard{
			Title: langText("按大小排列的热点段", "Top Segments by Size", "サイズ別トップセグメント", lang),
			Value: langText("无段数据可用。", "No segment data available.", "セグメントデータがありません。", lang),
		})
	}

	// 3. Process Invalid Objects list
	if invalidObjectsErr != nil {
		logger.Errorf("Error processing objects module - failed to get Invalid Objects list: %v", invalidObjectsErr)
		cards = append(cards, ReportCard{
			Title: langText("无效对象错误", "Invalid Objects Error", "無効なオブジェクトエラー", lang),
			Value: fmt.Sprintf(langText("获取无效对象列表失败: %v", "Failed to get invalid objects list: %v", "無効なオブジェクトリストの取得に失敗しました: %v", lang), invalidObjectsErr),
		})
		overallErr = appendError(overallErr, invalidObjectsErr)
	} else if allDbObjectInfo != nil && len(allDbObjectInfo.InvalidObjects) > 0 {
		invalidObjectsTable := &ReportTable{
			Name:    langText("无效对象", "Invalid Objects", "無効なオブジェクト", lang),
			Headers: []string{langText("所有者", "Owner", "所有者", lang), langText("对象名", "Object Name", "オブジェクト名", lang), langText("对象类型", "Object Type", "オブジェクトタイプ", lang), langText("创建时间", "Created", "作成日時", lang), langText("最后DDL时间", "Last DDL Time", "最終DDL時間", lang)},
			Rows:    [][]string{},
		}
		for _, obj := range allDbObjectInfo.InvalidObjects {
			row := []string{obj.Owner, obj.ObjectName, obj.ObjectType, obj.Created, obj.LastDDLTime}
			invalidObjectsTable.Rows = append(invalidObjectsTable.Rows, row)
		}
		tables = append(tables, invalidObjectsTable)
	} else {
		cards = append(cards, ReportCard{
			Title: langText("无效对象", "Invalid Objects", "無効なオブジェクト", lang),
			Value: langText("无无效对象数据可用。", "No invalid object data available.", "無効なオブジェクトデータがありません。", lang),
		})
	}

	charts = nil 
	return cards, tables, charts, overallErr
}

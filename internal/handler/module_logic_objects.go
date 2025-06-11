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
			Title: langText("Object Type Count Error", "Object Type Count Error", "Object Type Count Error", lang),
			Value: fmt.Sprintf(langText("Failed to get object type counts: %v", "Failed to get object type counts: %v", "Failed to get object type counts: %v", lang), overviewErr),
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
			Title: langText("Object Type Statistics", "Object Type Statistics", "Object Type Statistics", lang),
			Value: langText("No object type statistics data available.", "No object type statistics data available.", "No object type statistics data available.", lang),
		})
	}

	// 2. Process Top Segments information
	if topSegmentsErr != nil {
		logger.Errorf("Error processing objects module - failed to get Top Segments information: %v", topSegmentsErr)
		cards = append(cards, ReportCard{
			Title: langText("Top Segments by Size Error", "Top Segments by Size Error", "Top Segments by Size Error", lang),
			Value: fmt.Sprintf(langText("Failed to get top segments by size: %v", "Failed to get top segments by size: %v", "Failed to get top segments by size: %v", lang), topSegmentsErr),
		})
		overallErr = appendError(overallErr, topSegmentsErr)
	} else if allDbObjectInfo != nil && len(allDbObjectInfo.TopSegments) > 0 {
		topSegmentsTable := &ReportTable{
			Name:    langText("Top Segments by Size", "Top Segments by Size", "Top Segments by Size", lang),
			Headers: []string{langText("Owner", "Owner", "Owner", lang), langText("Segment Name", "Segment Name", "Segment Name", lang), langText("Segment Type", "Segment Type", "Segment Type", lang), langText("Size (GB)", "Size (GB)", "Size (GB)", lang)},
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
			Title: langText("Top Segments by Size", "Top Segments by Size", "Top Segments by Size", lang),
			Value: langText("No segment data available.", "No segment data available.", "No segment data available.", lang),
		})
	}

	// 3. Process Invalid Objects list
	if invalidObjectsErr != nil {
		logger.Errorf("Error processing objects module - failed to get Invalid Objects list: %v", invalidObjectsErr)
		cards = append(cards, ReportCard{
			Title: langText("Invalid Objects Error", "Invalid Objects Error", "Invalid Objects Error", lang),
			Value: fmt.Sprintf(langText("Failed to get invalid objects list: %v", "Failed to get invalid objects list: %v", "Failed to get invalid objects list: %v", lang), invalidObjectsErr),
		})
		overallErr = appendError(overallErr, invalidObjectsErr)
	} else if allDbObjectInfo != nil && len(allDbObjectInfo.InvalidObjects) > 0 {
		invalidObjectsTable := &ReportTable{
			Name:    langText("Invalid Objects", "Invalid Objects", "Invalid Objects", lang),
			Headers: []string{langText("Owner", "Owner", "Owner", lang), langText("Object Name", "Object Name", "Object Name", lang), langText("Object Type", "Object Type", "Object Type", lang), langText("Created", "Created", "Created", lang), langText("Last DDL Time", "Last DDL Time", "Last DDL Time", lang)},
			Rows:    [][]string{},
		}
		for _, obj := range allDbObjectInfo.InvalidObjects {
			row := []string{obj.Owner, obj.ObjectName, obj.ObjectType, obj.Created, obj.LastDDLTime}
			invalidObjectsTable.Rows = append(invalidObjectsTable.Rows, row)
		}
		tables = append(tables, invalidObjectsTable)
	} else {
		cards = append(cards, ReportCard{
			Title: langText("Invalid Objects", "Invalid Objects", "Invalid Objects", lang),
			Value: langText("No invalid object data available.", "No invalid object data available.", "No invalid object data available.", lang),
		})
	}

	charts = nil 
	return cards, tables, charts, overallErr
}

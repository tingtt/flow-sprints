package sprint

import (
	"flow-sprints/mysql"
	"time"
)

func GetList(userId uint64, projectId *uint64) (sprints []Sprint, err error) {
	// TODO: Embed child sprints

	// Generate query
	queryStr := "SELECT id, name, description, start, end, parent_id, project_id FROM sprints WHERE user_id = ?"
	queryParams := []interface{}{userId}
	if projectId != nil {
		queryStr += " AND project_id = ?"
		queryParams = append(queryParams, projectId)
	}
	queryStr += " ORDER BY start, end"

	db, err := mysql.Open()
	if err != nil {
		return
	}
	defer db.Close()

	stmtOut, err := db.Prepare(queryStr)
	if err != nil {
		return
	}
	defer stmtOut.Close()

	rows, err := stmtOut.Query(queryParams...)
	if err != nil {
		return
	}

	for rows.Next() {
		s := Sprint{}
		err = rows.Scan(&s.Id, &s.Name, &s.Description, &s.Start, &s.End, &s.ParentId, &s.ProjectId)
		if err != nil {
			return
		}
		sprints = append(sprints, s)
	}

	return
}

func GetListDate(userId uint64, dateStr string, dateRange *uint, projectId *uint64) (sprints []Sprint, invalidDateStr bool, invalidRange bool, err error) {
	// Validate params
	date, err := time.Parse("20060102", dateStr)
	if err != nil {
		date, err = time.Parse("2006-1-2", dateStr)
		if err != nil {
			err = nil
			invalidDateStr = true
			return
		}
	}
	if dateRange != nil && *dateRange <= 1 {
		invalidRange = true
		return
	}

	// TODO: Embed child sprints

	// Generate query
	queryStr := ""
	queryParams := []interface{}{userId}
	if dateRange == nil {
		queryStr = "SELECT id, name, description, start, end, parent_id, project_id FROM sprints WHERE user_id = ? AND ? BETWEEN start AND end"
		queryParams = append(queryParams, dateStr)
	} else {
		queryStr = "SELECT id, name, description, start, end, parent_id, project_id FROM sprints WHERE user_id = ? AND (? BETWEEN start AND end OR ? BETWEEN start AND end OR start BETWEEN ? AND ?)"
		dateEndStr := date.AddDate(0, 0, int(*dateRange)-1).Format("2006-1-2")
		queryParams = append(queryParams, dateStr, dateEndStr, dateStr, dateEndStr)
	}
	if projectId != nil {
		queryStr += " AND project_id = ?"
		queryParams = append(queryParams, projectId)
	}
	queryStr += " ORDER BY start, end"

	db, err := mysql.Open()
	if err != nil {
		return
	}
	defer db.Close()

	stmtOut, err := db.Prepare(queryStr)
	if err != nil {
		return
	}
	defer stmtOut.Close()

	rows, err := stmtOut.Query(queryParams...)
	if err != nil {
		return
	}

	for rows.Next() {
		s := Sprint{}
		err = rows.Scan(&s.Id, &s.Name, &s.Description, &s.Start, &s.End, &s.ParentId, &s.ProjectId)
		if err != nil {
			return
		}
		sprints = append(sprints, s)
	}

	return
}

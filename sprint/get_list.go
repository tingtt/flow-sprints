package sprint

import (
	"database/sql"
	"flow-sprints/mysql"
	"time"
)

func GetList(userId uint64, projectId *uint64) (sprints []Sprint, err error) {
	// TODO: Embed child sprints

	// Generate query
	queryStr := "SELECT id, name, description, start, end, parent_id, project_id FROM sprints WHERE user_id = ?"
	if projectId != nil {
		queryStr += " AND project_id = ?"
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

	var rows *sql.Rows
	if projectId == nil {
		rows, err = stmtOut.Query(userId)
	} else {
		rows, err = stmtOut.Query(userId, *projectId)
	}
	if err != nil {
		return
	}

	for rows.Next() {
		// TODO: uint64に対応
		var (
			id          uint64
			name        string
			description sql.NullString
			start       string
			end         string
			parentId    sql.NullInt64
			projectId   sql.NullInt64
		)
		err = rows.Scan(&id, &name, &description, &start, &end, &parentId, &projectId)
		if err != nil {
			return
		}

		t := Sprint{Id: id, Name: name, Start: start, End: end}
		if description.Valid {
			t.Description = &description.String
		}
		if parentId.Valid {
			sprintIdTmp := uint64(parentId.Int64)
			t.ParentId = &sprintIdTmp
		}
		if projectId.Valid {
			projectIdTmp := uint64(projectId.Int64)
			t.ProjectId = &projectIdTmp
		}

		sprints = append(sprints, t)
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
	if dateRange == nil {
		queryStr = "SELECT id, name, description, start, end, parent_id, project_id FROM sprints WHERE user_id = ? AND ? BETWEEN start AND end"
	} else {
		queryStr = "SELECT id, name, description, start, end, parent_id, project_id FROM sprints WHERE user_id = ? AND (? BETWEEN start AND end OR ? BETWEEN start AND end OR start BETWEEN ? AND ?)"
	}
	if projectId != nil {
		queryStr += " AND project_id = ?"
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

	var rows *sql.Rows
	if dateRange == nil {
		if projectId == nil {
			rows, err = stmtOut.Query(userId, dateStr)
		} else {
			rows, err = stmtOut.Query(userId, dateStr, *projectId)
		}
	} else {
		dateEnd := date.AddDate(0, 0, int(*dateRange)-1)
		if projectId == nil {
			rows, err = stmtOut.Query(userId, dateStr, dateEnd.Format("2006-1-2"), dateStr, dateEnd.Format("2006-1-2"))
		} else {
			rows, err = stmtOut.Query(userId, dateStr, dateEnd.Format("2006-1-2"), dateStr, dateEnd.Format("2006-1-2"), *projectId)
		}
	}
	if err != nil {
		return
	}

	for rows.Next() {
		// TODO: uint64に対応
		var (
			id          uint64
			name        string
			description sql.NullString
			start       string
			end         string
			parentId    sql.NullInt64
			projectId   sql.NullInt64
		)
		err = rows.Scan(&id, &name, &description, &start, &end, &parentId, &projectId)
		if err != nil {
			return
		}

		t := Sprint{Id: id, Name: name, Start: start, End: end}
		if description.Valid {
			t.Description = &description.String
		}
		if parentId.Valid {
			sprintIdTmp := uint64(parentId.Int64)
			t.ParentId = &sprintIdTmp
		}
		if projectId.Valid {
			projectIdTmp := uint64(projectId.Int64)
			t.ProjectId = &projectIdTmp
		}

		sprints = append(sprints, t)
	}

	return
}

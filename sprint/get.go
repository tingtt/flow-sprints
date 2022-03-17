package sprint

import (
	"database/sql"
	"flow-sprints/mysql"
)

func Get(userId uint64, id uint64) (t Sprint, notFound bool, err error) {
	db, err := mysql.Open()
	if err != nil {
		return Sprint{}, false, err
	}
	defer db.Close()

	stmtOut, err := db.Prepare("SELECT name, description, start, end, parent_id, project_id FROM sprints WHERE user_id = ? AND id = ?")
	if err != nil {
		return Sprint{}, false, err
	}
	defer stmtOut.Close()

	rows, err := stmtOut.Query(userId, id)
	if err != nil {
		return Sprint{}, false, err
	}

	// TODO: uint64に対応
	var (
		name        string
		description sql.NullString
		start       sql.NullString
		end         sql.NullString
		parentId    sql.NullInt64
		projectId   sql.NullInt64
	)
	if !rows.Next() {
		// Not found
		return Sprint{}, true, nil
	}
	err = rows.Scan(&name, &description, &start, &end, &parentId, &projectId)
	if err != nil {
		return Sprint{}, false, err
	}

	t.Id = id
	t.Name = name
	if description.Valid {
		t.Description = &description.String
	}
	if start.Valid {
		t.Start = start.String
	}
	if end.Valid {
		t.End = end.String
	}
	if parentId.Valid {
		sprintIdTmp := uint64(parentId.Int64)
		t.ParentId = &sprintIdTmp
	}
	if projectId.Valid {
		projectIdTmp := uint64(projectId.Int64)
		t.ProjectId = &projectIdTmp
	}

	return
}

package sprint

import (
	"database/sql"
	"flow-sprints/mysql"
	"time"

	"github.com/go-playground/validator"
)

type Sprint struct {
	Id          uint64  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Start       string  `json:"start"`
	End         string  `json:"end"`
	ParentId    *uint64 `json:"parent_id,omitempty"`
	ProjectId   *uint64 `json:"project_id,omitempty"`
}

type Post struct {
	Name        string  `json:"name" validate:"required"`
	Description *string `json:"description" validate:"omitempty"`
	Start       string  `json:"start,omitempty" validate:"required,Y-M-D"`
	End         string  `json:"end,omitempty" validate:"required,Y-M-D"`
	ParentId    *uint64 `json:"parent_id" validate:"omitempty,gte=1"`
	ProjectId   *uint64 `json:"project_id" validate:"omitempty,gte=1"`
}

type Patch struct {
	Name        *string `json:"name" validate:"omitempty"`
	Description *string `json:"description" validate:"omitempty"`
	Start       *string `json:"start,omitempty" validate:"omitempty,Y-M-D"`
	End         *string `json:"end,omitempty" validate:"omitempty,Y-M-D"`
	ParentId    *uint64 `json:"parent_id" validate:"omitempty,gte=1"`
	ProjectId   *uint64 `json:"project_id" validate:"omitempty,gte=1"`
}

func DateStrValidation(fl validator.FieldLevel) bool {
	// `yyyy-mm-dd`
	_, err := time.Parse("2006-1-2", fl.Field().String())
	return err == nil
}

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

func Insert(userId uint64, post Post) (p Sprint, startAfterEnd bool, invalidParentId bool, invalidChildDate bool, err error) {
	// Check start/end
	start, err := time.Parse("2006-1-2", post.Start)
	if err != nil {
		return
	}
	end, err := time.Parse("2006-1-2", post.End)
	if err != nil {
		return
	}
	if start.After(end) {
		startAfterEnd = true
		return
	}

	// Check parent id/start/end
	//TODO: Check parent has no parent
	if post.ParentId != nil {
		var parent Sprint
		parent, invalidParentId, err = Get(userId, *post.ParentId)
		if err != nil {
			return
		}
		if invalidParentId {
			invalidParentId = true
			return
		}
		pStart, _ := time.Parse("2006-1-2", parent.Start)
		pEnd, _ := time.Parse("2006-1-2", parent.End)
		if start.Before(pStart) || end.After(pEnd) {
			// child start/end not between parent start/end
			invalidChildDate = true
			return
		}
	}

	// Insert DB
	db, err := mysql.Open()
	if err != nil {
		return
	}
	defer db.Close()
	stmtIns, err := db.Prepare("INSERT INTO sprints (user_id, name, description, start, end, parent_id, project_id) VALUES (?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return
	}
	defer stmtIns.Close()
	result, err := stmtIns.Exec(userId, post.Name, post.Description, post.Start, post.End, post.ParentId, post.ProjectId)
	if err != nil {
		return
	}
	id, err := result.LastInsertId()
	if err != nil {
		return
	}

	p.Id = uint64(id)
	p.Name = post.Name
	p.Start = post.Start
	p.End = post.End
	if post.Description != nil {
		p.Description = post.Description
	}
	if post.ParentId != nil {
		p.ParentId = post.ParentId
	}
	if post.ProjectId != nil {
		p.ProjectId = post.ProjectId
	}

	return
}

func Update(userId uint64, id uint64, new Patch) (t Sprint, notFound bool, startAfterEnd bool, parentNotFound bool, loopParent bool, invalidChildDate bool, err error) {
	// Get old
	old, notFound, err := Get(userId, id)
	if err != nil {
		return
	}
	if notFound {
		return
	}

	// Set no update values
	if new.Name == nil {
		new.Name = &old.Name
	}
	if new.Description == nil {
		new.Description = old.Description
	}
	if new.Start == nil {
		new.Start = &old.Start
	}
	if new.End == nil {
		new.End = &old.End
	}
	if new.ParentId == nil {
		new.ParentId = old.ParentId
	}
	if new.ProjectId == nil {
		new.ProjectId = old.ProjectId
	}

	// Check start/end
	start, err := time.Parse("2006-1-2", *new.Start)
	if err != nil {
		return
	}
	end, err := time.Parse("2006-1-2", *new.End)
	if err != nil {
		return
	}
	if start.After(end) {
		startAfterEnd = true
		return
	}

	// Check parent id/start/end
	if new.ParentId != nil {
		var parent Sprint
		parent, parentNotFound, err = Get(userId, *new.ParentId)
		if err != nil {
			return
		}
		if parentNotFound {
			parentNotFound = true
			return
		}
		if parent.Id == id {
			loopParent = true
			return
		}
		pStart, _ := time.Parse("2006-1-2", parent.Start)
		pEnd, _ := time.Parse("2006-1-2", parent.End)
		if start.Before(pStart) || end.After(pEnd) {
			// child start/end not between parent start/end
			invalidChildDate = true
			return
		}
	}

	// TODO: Update child start/end

	// Update row
	db, err := mysql.Open()
	if err != nil {
		return
	}
	defer db.Close()
	stmtIns, err := db.Prepare("UPDATE sprints SET name = ?, description = ?, start = ?, end = ?, parent_id = ?, project_id = ? WHERE user_id = ? AND id = ?")
	if err != nil {
		return
	}
	defer stmtIns.Close()
	_, err = stmtIns.Exec(new.Name, new.Description, new.Start, new.End, new.ParentId, new.ProjectId, userId, id)
	if err != nil {
		return
	}

	t = Sprint{id, *new.Name, new.Description, *new.Start, *new.End, new.ParentId, new.ProjectId}
	return
}

func Delete(userId uint64, id uint64) (notFound bool, err error) {
	db, err := mysql.Open()
	if err != nil {
		return false, err
	}
	defer db.Close()
	stmtIns, err := db.Prepare("DELETE FROM sprints WHERE user_id = ? AND id = ?")
	if err != nil {
		return false, err
	}
	defer stmtIns.Close()
	result, err := stmtIns.Exec(userId, id)
	if err != nil {
		return false, err
	}
	affectedRowCount, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	if affectedRowCount == 0 {
		// Not found
		return true, nil
	}

	return false, nil
}

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

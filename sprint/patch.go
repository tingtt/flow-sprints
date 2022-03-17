package sprint

import (
	"flow-sprints/mysql"
	"strings"
	"time"
)

type PatchBody struct {
	Name        *string `json:"name" validate:"omitempty"`
	Description *string `json:"description" validate:"omitempty"`
	Start       *string `json:"start,omitempty" validate:"omitempty,Y-M-D"`
	End         *string `json:"end,omitempty" validate:"omitempty,Y-M-D"`
	ProjectId   *uint64 `json:"project_id" validate:"omitempty,gte=1"`
}

func Patch(userId uint64, id uint64, new PatchBody) (s Sprint, notFound bool, startAfterEnd bool, err error) {
	// Get old
	s, notFound, err = Get(userId, id)
	if err != nil {
		return
	}
	if notFound {
		return
	}

	// Generate query
	queryStr := "UPDATE schemes SET"
	var queryParams []interface{}
	if new.Name != nil {
		queryStr += " name = ?,"
		queryParams = append(queryParams, new.Name)
		s.Name = *new.Name
	}
	if new.Description != nil {
		queryStr += " description = ?,"
		queryParams = append(queryParams, new.Description)
		s.Description = new.Description
	}
	if new.Start != nil {
		queryStr += " start = ?,"
		queryParams = append(queryParams, new.Start)
		s.Start = *new.Start
	}
	if new.End != nil {
		queryStr += " end = ?,"
		queryParams = append(queryParams, new.End)
		s.End = *new.End
	}
	if new.ProjectId != nil {
		queryStr += " project_id = ?,"
		queryParams = append(queryParams, new.ProjectId)
		s.ProjectId = new.ProjectId
	}
	queryStr = strings.TrimRight(queryStr, ",")
	queryStr += " WHERE user_id = ? AND id = ?"
	queryParams = append(queryParams, userId, id)

	// Check start/end
	start, err := time.Parse("2006-1-2", s.Start)
	if err != nil {
		return
	}
	end, err := time.Parse("2006-1-2", s.End)
	if err != nil {
		return
	}
	if start.After(end) {
		startAfterEnd = true
		return
	}

	// Update row
	db, err := mysql.Open()
	if err != nil {
		return
	}
	defer db.Close()
	stmtIns, err := db.Prepare(queryStr)
	if err != nil {
		return
	}
	defer stmtIns.Close()
	_, err = stmtIns.Exec(queryParams...)
	if err != nil {
		return
	}

	return
}
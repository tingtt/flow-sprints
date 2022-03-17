package sprint

import (
	"flow-sprints/mysql"
	"time"

	"github.com/go-playground/validator"
)

type PostBody struct {
	Name        string  `json:"name" validate:"required"`
	Description *string `json:"description" validate:"omitempty"`
	Start       string  `json:"start,omitempty" validate:"required,Y-M-D"`
	End         string  `json:"end,omitempty" validate:"required,Y-M-D"`
	ParentId    *uint64 `json:"parent_id" validate:"omitempty,gte=1"`
	ProjectId   *uint64 `json:"project_id" validate:"omitempty,gte=1"`
}

func DateStrValidation(fl validator.FieldLevel) bool {
	// `yyyy-mm-dd`
	_, err := time.Parse("2006-1-2", fl.Field().String())
	return err == nil
}

func Post(userId uint64, post PostBody) (p Sprint, startAfterEnd bool, invalidParentId bool, invalidChildDate bool, err error) {
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

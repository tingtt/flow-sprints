package sprint

import (
	"flow-sprints/mysql"
	"time"
)

type PatchBody struct {
	Name        *string `json:"name" validate:"omitempty"`
	Description *string `json:"description" validate:"omitempty"`
	Start       *string `json:"start,omitempty" validate:"omitempty,Y-M-D"`
	End         *string `json:"end,omitempty" validate:"omitempty,Y-M-D"`
	ParentId    *uint64 `json:"parent_id" validate:"omitempty,gte=1"`
	ProjectId   *uint64 `json:"project_id" validate:"omitempty,gte=1"`
}

func Patch(userId uint64, id uint64, new PatchBody) (t Sprint, notFound bool, startAfterEnd bool, parentNotFound bool, loopParent bool, invalidChildDate bool, err error) {

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

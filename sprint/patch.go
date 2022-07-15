package sprint

import (
	"encoding/json"
	"flow-sprints/mysql"
	"strings"
	"time"
)

type PatchBody struct {
	Name        *string             `json:"name" validate:"omitempty"`
	Description PatchNullJSONString `json:"description" validate:"omitempty"`
	Start       *string             `json:"start,omitempty" validate:"omitempty,Y-M-D"`
	End         *string             `json:"end,omitempty" validate:"omitempty,Y-M-D"`
	ProjectId   PatchNullJSONUint64 `json:"project_id" validate:"dive"`
}

type PatchNullJSONString struct {
	String **string `validate:"omitempty,gte=1"`
}

type PatchNullJSONUint64 struct {
	UInt64 **uint64 `validate:"omitempty,gte=1"`
}

func (p *PatchNullJSONString) UnmarshalJSON(data []byte) error {
	// If this method was called, the value was set.
	var valueP *string = nil
	if string(data) == "null" {
		// key exists and value is null
		p.String = &valueP
		return nil
	}

	var tmp string
	tmpP := &tmp
	if err := json.Unmarshal(data, &tmp); err != nil {
		// invalid value type
		return err
	}
	// valid value
	p.String = &tmpP
	return nil
}

func (p *PatchNullJSONUint64) UnmarshalJSON(data []byte) error {
	// If this method was called, the value was set.
	var valueP *uint64 = nil
	if string(data) == "null" {
		// key exists and value is null
		p.UInt64 = &valueP
		return nil
	}

	var tmp uint64
	tmpP := &tmp
	if err := json.Unmarshal(data, &tmp); err != nil {
		// invalid value type
		return err
	}
	// valid value
	p.UInt64 = &tmpP
	return nil
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
	queryStr := "UPDATE sprints SET"
	var queryParams []interface{}
	if new.Name != nil {
		queryStr += " name = ?,"
		queryParams = append(queryParams, new.Name)
		s.Name = *new.Name
	}
	if new.Description.String != nil {
		if *new.Description.String != nil {
			queryStr += " description = ?,"
			queryParams = append(queryParams, **new.Description.String)
			s.Description = *new.Description.String
		} else {
			queryStr += " description = ?,"
			queryParams = append(queryParams, nil)
			s.Description = nil
		}
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
	if new.ProjectId.UInt64 != nil {
		if *new.ProjectId.UInt64 != nil {
			queryStr += " project_id = ?"
			queryParams = append(queryParams, **new.ProjectId.UInt64)
			s.ProjectId = *new.ProjectId.UInt64
		} else {
			queryStr += " project_id = ?"
			queryParams = append(queryParams, nil)
			s.ProjectId = nil
		}
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

package sprint

import (
	"flow-sprints/mysql"
)

type GetListQuery struct {
	Start     *string `query:"start" validate:"omitempty,Y-M-D"`
	End       *string `query:"end" validate:"omitempty,Y-M-D"`
	ProjectId *uint64 `query:"project_id" validate:"omitempty,gte=1"`
}

func GetList(userId uint64, q GetListQuery) (sprints []Sprint, err error) {
	// Generate query
	queryStr := "SELECT id, name, description, start, end, project_id FROM sprints WHERE user_id = ?"
	queryParams := []interface{}{userId}
	if q.Start != nil {
		queryStr += " AND end >= ?"
		queryParams = append(queryParams, q.Start)
	}
	if q.End != nil {
		queryStr += " AND start <= ?"
		queryParams = append(queryParams, q.End)
	}
	if q.ProjectId != nil {
		queryStr += " AND project_id = ?"
		queryParams = append(queryParams, q.ProjectId)
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
		err = rows.Scan(&s.Id, &s.Name, &s.Description, &s.Start, &s.End, &s.ProjectId)
		if err != nil {
			return
		}
		sprints = append(sprints, s)
	}

	return
}

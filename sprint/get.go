package sprint

import (
	"flow-sprints/mysql"
)

func Get(userId uint64, id uint64) (s Sprint, notFound bool, err error) {
	db, err := mysql.Open()
	if err != nil {
		return Sprint{}, false, err
	}
	defer db.Close()

	stmtOut, err := db.Prepare("SELECT name, description, start, end, project_id FROM sprints WHERE user_id = ? AND id = ?")
	if err != nil {
		return Sprint{}, false, err
	}
	defer stmtOut.Close()

	rows, err := stmtOut.Query(userId, id)
	if err != nil {
		return Sprint{}, false, err
	}

	if !rows.Next() {
		// Not found
		return Sprint{}, true, nil
	}
	err = rows.Scan(&s.Name, &s.Description, &s.Start, &s.End, &s.ProjectId)
	if err != nil {
		return Sprint{}, false, err
	}

	s.Id = id
	return
}

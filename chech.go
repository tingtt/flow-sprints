package main

import (
	"fmt"
	"net/http"
)

func checkProjectId(token string, id uint64) (valid bool, err error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/%d", *serviceUrlProjects, id), nil)
	if err != nil {
		return
	}
	req.Header.Set("Authorization", "Bearer "+token)

	client := new(http.Client)
	res, err := client.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()

	valid = res.StatusCode == 200
	return
}

package utils

import "net/http"

func HttpGet(url string, bearer *string) (status int, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	if bearer != nil {
		req.Header.Set("Authorization", "Bearer "+*bearer)
	}

	client := new(http.Client)
	res, err := client.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()

	status = res.StatusCode
	return
}

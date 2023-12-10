package httprequest

import (
	"net/http"
)

func Request(method, url string) (*http.Response, error) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	const contentType = "application/json"
	req.Header.Add("Accept", contentType)
	req.Header.Add("Content-Type", contentType)

	return client.Do(req)
}

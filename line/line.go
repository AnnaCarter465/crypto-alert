package line

import (
	"log"
	"net/http"
	"net/url"
	"os"
)

func NotiToLine(message string) {
	param := url.Values{}
	param.Add("message", message)
	url := "https://notify-api.line.me/api/notify?" + param.Encode()
	token := os.Getenv("LINE_TOKEN")

	client := http.Client{}

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		panic(err)
	}

	req.Header.Add("Content-type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", "Bearer "+token)

	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	log.Println("line res:", res)

	res.Body.Close()
}

package main

import (
	"bufio"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/time/rate"
)

type httpClient struct {
	client  *http.Client
	limiter *rate.Limiter
}

func NewHTTPClient(ops int) *httpClient {
	return &httpClient{
		client:  &http.Client{},
		limiter: rate.NewLimiter(rate.Every(time.Second), ops),
	}
}

func (c *httpClient) Do(req *http.Request) (*http.Response, error) {

	return c.client.Do(req)
}

func main() {
	client := http.DefaultClient
	// здесь, например, API биржевых котировок
	URL := "https://iss.moex.com/iss/statistics/engines/futures/markets/indicativerates/securities.xml"
	req, _ := http.NewRequest("GET", URL, nil)
	for {
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		// логика обработки результата запроса
		defer resp.Body.Close()
		scanner := bufio.NewScanner(resp.Body)
		for i := 0; scanner.Scan() && i < 20; i++ {
			fmt.Println(scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			panic(err)
		}
	}
}

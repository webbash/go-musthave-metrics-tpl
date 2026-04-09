package main

import (
	"flag"
	"net/http"
	"time"

	"github.com/webbash/go-musthave-metrics-tpl.git/internal/agent"
)

func main() {
	basicURL := flag.String("a", "localhost:8080", "agent endpoint url")
	pollInterval := flag.Int("p", 2, "poll interval in seconds")
	reportInterval := flag.Int("r", 10, "report interval in seconds")
	agentClient := agent.NewAgent(*basicURL, time.Duration(*pollInterval)*time.Second, time.Duration(*reportInterval)*time.Second, &http.Client{})

	flag.Parse()

	go func() {
		for {
			time.Sleep(agentClient.PollInterval)
			agentClient.ReadMetrics()
		}
	}()

	for {
		time.Sleep(agentClient.ReportInterval)
		agentClient.SendMetrics()
	}
}

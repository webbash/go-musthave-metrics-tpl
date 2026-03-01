package main

import (
	"net/http"
	"time"

	"github.com/webbash/go-musthave-metrics-tpl.git/internal/agent"
)

func main() {
	agentClient := agent.NewAgent("http://localhost:8080", 2*time.Second, 10*time.Second, &http.Client{})

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

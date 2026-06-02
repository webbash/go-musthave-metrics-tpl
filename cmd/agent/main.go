package main

import (
	"flag"
	"log/slog"
	"net/http"
	"time"

	"github.com/webbash/go-musthave-metrics-tpl.git/internal/agent"
)

func main() {
	basicURL := flag.String("a", "localhost:8080", "agent endpoint url")
	pollInterval := flag.Int("p", 2, "poll interval in seconds")
	reportInterval := flag.Int("r", 10, "report interval in seconds")

	flag.Parse()

	agentClient := agent.NewAgent(*basicURL, time.Duration(*pollInterval)*time.Second, time.Duration(*reportInterval)*time.Second, &http.Client{})

	go func() {
		pollTicker := time.NewTicker(agentClient.PollInterval)
		defer pollTicker.Stop()

		for {
			select {
			case <-pollTicker.C:
				agentClient.ReadMetrics()
			}
		}
	}()

	reportTicker := time.NewTicker(agentClient.ReportInterval)
	defer reportTicker.Stop()

	for {
		select {
		case <-reportTicker.C:
			if err := agentClient.SendMetrics(); err != nil {
				slog.Error("failed to send metrics", "error", err)
			}
		}
	}
}

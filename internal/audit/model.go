package audit

// Event describes metrics received by the server in a single update request.
type Event struct {
	// TS is the Unix timestamp when the event occurred.
	TS int64 `json:"ts"`
	// Metrics contains the names of the updated metrics.
	Metrics []string `json:"metrics"`
	// IPAddress is the network address of the client that sent the update.
	IPAddress string `json:"ip_address"`
}

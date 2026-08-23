package audit

// Observer handles metric update events produced by the server.
type Observer interface {
	// Observe records or forwards an audit event.
	Observe(Event) error
}

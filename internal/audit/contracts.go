package audit

type Observer interface {
	Observe(Event) error
}

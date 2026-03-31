package get_value_list

type storage interface {
	GetAllGauges() map[string]float64
	GetAllCounters() map[string]int64
}

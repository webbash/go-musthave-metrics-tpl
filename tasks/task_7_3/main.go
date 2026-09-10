package main

type Slice[T any] []T

func (s *Slice[T]) Filter(allow func(e T) bool) *Slice[T] {
	var result Slice[T]

	for _, e := range *s {
		if allow(e) {
			result = append(result, e)
		}
	}

	*s = result
	return s
}

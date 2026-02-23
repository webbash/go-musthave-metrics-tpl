package update

import "net/http"

type Handler struct {
}

func (h Handler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	data := []byte("Привет!")
	res.Write(data)
}

package get_value_list

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/go-chi/chi/v5"

	"github.com/webbash/go-musthave-metrics-tpl.git/internal/repository"
)

func ExampleHandler() {
	storage := repository.NewMemStorage()
	handler := NewHandler(storage)

	router := chi.NewRouter()
	router.Get("/", handler.ServeHTTP)

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	fmt.Println(recorder.Code)
	fmt.Println(recorder.Header().Get("Content-Type"))
	// Output:
	// 200
	// text/html
}

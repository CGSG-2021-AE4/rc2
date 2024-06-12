package api

import (
	"encoding/json"
	"net/http"
)

func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(v)
}

type apiHandleFunc func(http.ResponseWriter, *http.Request) error

type ApiError struct {
	Error string
}

func makeHandleFunc(f apiHandleFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
	}
}

// Error implementation

func (e rcError) Error() string {
	return e.err
}

func NewError(msg string) rcError {
	return rcError{
		err: msg,
	}
}

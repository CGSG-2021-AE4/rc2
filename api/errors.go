package api

// Simple error type
type Error string

func (err Error) Error() string {
	return string(err)
}

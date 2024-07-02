package api

// Error implementation - found in Effective Go
type rcError string

func (err rcError) Error() string {
	return string(err)
}

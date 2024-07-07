package api

import "log"

// Simple error type
type Error string

func (err Error) Error() string {
	return string(err)
}

// Function that runs a function and log returned error
func RunAndLog(f func() error, name string) {
	log.Println("Start", name)
	if err := f(); err != nil {
		log.Println("Finish", name, "with error:", err.Error())
	} else {
		log.Println("Finish", name)
	}
}

package api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type sendRequest struct {
	Login    string
	Password string
	Msg      string
}

func (s *APIServer) handleSend(w http.ResponseWriter, r *http.Request) error {
	fmt.Println("CGSG forever!!!")

	var body sendRequest

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return err
	}
	fmt.Printf("Body:%+v\n", body)

	return nil
}

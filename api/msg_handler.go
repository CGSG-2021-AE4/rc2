package api

import (
	"encoding/json"
	"net/http"

	cw "github.com/CGSG-2021-AE4/go_utils/conn_wrapper"
	"github.com/gin-gonic/gin"
)

func NewMsgHandlerService(s *APIServer) *MsgHandlerService {
	return &MsgHandlerService{
		s: s,
	}
}

func (mh *MsgHandlerService) HandleHTTP(c *gin.Context) {
	var body sendRequestMsg
	if err := json.NewDecoder(c.Request.Body).Decode(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "Invalid json"})
		return
	}

	conn := mh.s.clientService.conns[body.Login]
	if c == nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "Client is not connected"})
		return
	}

	response, err := conn.WriteMsg(body.Msg)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "Response type: " + cw.FormatError(response.mt) + ", Msg: " + string(response.buf)})
}

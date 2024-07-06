package msg_service

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"

	cw "github.com/CGSG-2021-AE4/go_utils/conn_wrapper"

	"github.com/CGSG-2021-AE4/rc2/api"
	"github.com/CGSG-2021-AE4/rc2/api/server/client_service"
)

// Message handler service
type Service struct {
	cs *client_service.Service
}

func New(cs *client_service.Service) *Service {
	return &Service{
		cs: cs,
	}
}

func (mh *Service) HandleHTTP(c *gin.Context) {
	var body api.SendRequestMsg
	if err := json.NewDecoder(c.Request.Body).Decode(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "Invalid json"})
		return
	}

	conn := mh.cs.Conns[body.Login]
	if c == nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "Client is not connected"})
		return
	}

	response, err := conn.WriteMsg(body.Msg)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "Response type: " + cw.FormatError(response.Mt) + ", Msg: " + string(response.Buf)})
}

package api

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"time"
)

func (ss *StatService) load() error {
	if _, err := os.Stat(ss.server.env.StatisticsFile); errors.Is(err, os.ErrNotExist) {
		return nil // Just do not load anything
	}

	data, err := os.ReadFile(ss.server.env.StatisticsFile)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, &ss.currentStat); err != nil {
		return err
	}
	return nil
}

func (ss *StatService) save() error {
	log.Println("Save")
	// Update some values
	ss.currentStat.Started = ss.startTime
	ss.currentStat.WriteTime = time.Now()
	ss.currentStat.WorkDur += ss.currentStat.Started.Sub(ss.currentStat.Started)

	// Write
	buf, err := json.Marshal(ss.currentStat)
	if err != nil {
		return err
	}
	if err := os.WriteFile(ss.server.env.StatisticsFile, buf, 0770); err != nil {
		return err
	}
	return nil
}

func (ss *StatService) OnConnect(c *ClientConn) {
	ss.Mutex.Lock()
	defer ss.Mutex.Unlock()

	ss.connectedUsers[c.login] = connectedUserStat{
		addr:          c.conn.NetConn.RemoteAddr().String(),
		connStartTime: time.Now(),
	}
}

func (ss *StatService) OnDisconnect(c *ClientConn) {
	ss.Mutex.Lock()
	defer ss.Mutex.Unlock()

	if _, err := ss.connectedUsers[c.login]; err == false {
		return // Cannot find user but it is not fatal
	}

	ui := -1 // Index of this user in stat
	for i := range len(ss.currentStat.Users) {
		if ss.currentStat.Users[i].Login == c.login {
			ui = i
			break
		}
	}
	if ui == -1 {
		// There is no such user so far
		// Create
		ss.currentStat.Users = append(ss.currentStat.Users, statPerUser{Login: c.login})
		ui = len(ss.currentStat.Users)
	}
	ss.currentStat.Users[ui].LastAddr = ss.connectedUsers[c.login].addr
	ss.currentStat.Users[ui].LastSeen = time.Now()
	ss.currentStat.Users[ui].WorkDur += ss.currentStat.Users[ui].LastSeen.Sub(ss.connectedUsers[c.login].connStartTime)
}

// Update statistics
func (ss *StatService) updateStat() error {
	log.Println("Stat update")
	// Old piece of code
	// Maybe I should save stat here
	return nil
}

func (ss *StatService) run() (err error) {
	defer func() {
		// Run cannot return an error so I have to handle it here
		if err != nil {
			log.Println("Stat service run finished with error:", err.Error())
		} else {
			log.Println("Stat service run finished")
		}
	}()

	// load
	if err := ss.load(); err != nil {
		return err
	}

	<-ss.server.DoneChan
	return ss.save()
	// update cycle
	// for {
	// 	log.Println("Stat cycle")
	// 	select {
	// 	case <-ss.server.DoneChan:
	// 		return ss.save()
	// 	case <-time.After(time.Duration(ss.server.env.StatTimeout * time.Millisecond)):
	// 		if err := ss.updateStat(); err != nil {
	// 			log.Println("Update state error:", err.Error())
	// 		}
	// 		break
	// 	}
	// }
}

func NewStatService(server *APIServer) *StatService {
	ss := StatService{
		server:         server,
		startTime:      time.Now(),
		connectedUsers: make(map[string]connectedUserStat),
	}
	go ss.run()

	return &ss
}

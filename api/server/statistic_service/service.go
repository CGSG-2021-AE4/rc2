package statistic_service

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"sync"
	"time"

	"github.com/CGSG-2021-AE4/rc2/api"
	"github.com/CGSG-2021-AE4/rc2/api/server/client_service"
)

/////////////// Statistics

type statPerUser struct {
	Login    string        `json:"login"`
	LastAddr string        `json:"lastIp"`
	WorkDur  time.Duration `json:"workDuration"`
	LastSeen time.Time     `json:"lastSeen"`
}

type stat struct {
	Host      string        `json:"host"`
	WorkDur   time.Duration `json:"workDuration"`
	Started   time.Time     `json:"started"`
	WriteTime time.Time     `json:"writeTime"`
	Users     []statPerUser `json:"users"`
}

type connectedUserStat struct {
	connStartTime time.Time
	addr          string
}

///////////////////// Statics service

type Service struct {
	filename       string
	currentStat    stat
	startTime      time.Time
	Mutex          sync.Mutex // Locks all local data
	connectedUsers map[string]connectedUserStat
}

func New(filename string) *Service {
	ss := Service{
		filename:       filename,
		startTime:      time.Now(),
		connectedUsers: make(map[string]connectedUserStat),
	}
	return &ss
}

func (ss *Service) Run() {
	go api.RunAndLog(ss.RunSync, "Statistics service")
}

func (ss *Service) OnConnect(c *client_service.Conn) {
	ss.Mutex.Lock()
	defer ss.Mutex.Unlock()

	ss.connectedUsers[c.Login] = connectedUserStat{
		// addr:          c.Conn.NetConn.RemoteAddr().String(),
		connStartTime: time.Now(),
	}
}

func (ss *Service) OnDisconnect(c *client_service.Conn) {
	ss.Mutex.Lock()
	defer ss.Mutex.Unlock()

	if _, err := ss.connectedUsers[c.Login]; err == false {
		return // Cannot find user but it is not fatal
	}

	ui := -1 // Index of this user in stat
	for i := range len(ss.currentStat.Users) {
		if ss.currentStat.Users[i].Login == c.Login {
			ui = i
			break
		}
	}
	if ui == -1 {
		// There is no such user so far
		// Create
		ss.currentStat.Users = append(ss.currentStat.Users, statPerUser{Login: c.Login})
		ui = len(ss.currentStat.Users)
	}
	ss.currentStat.Users[ui].LastAddr = ss.connectedUsers[c.Login].addr
	ss.currentStat.Users[ui].LastSeen = time.Now()
	ss.currentStat.Users[ui].WorkDur += ss.currentStat.Users[ui].LastSeen.Sub(ss.connectedUsers[c.Login].connStartTime)
}

func (ss *Service) RunSync() error {
	// load
	if err := ss.load(); err != nil {
		return err
	}

	// <-ss.server.DoneChan // TODO
	return ss.save()
}

func (ss *Service) load() error {
	if _, err := os.Stat(ss.filename); errors.Is(err, os.ErrNotExist) {
		return nil // Just do not load anything
	}

	data, err := os.ReadFile(ss.filename)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, &ss.currentStat); err != nil {
		return err
	}
	return nil
}

func (ss *Service) save() error {
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
	if err := os.WriteFile(ss.filename, buf, 0770); err != nil {
		return err
	}
	return nil
}

// Update statistics
func (ss *Service) updateStat() error {
	log.Println("Stat update")
	// Old piece of code
	// Maybe I should save stat here
	return nil
}

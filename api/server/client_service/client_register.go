package client_service

import (
	"log"
	"sync"

	"github.com/CGSG-2021-AE4/rc2/api"
)

type clientRegister struct {
	mutex sync.Mutex
	conns map[string]*Conn
}

func newClientRegister() *clientRegister {
	return &clientRegister{
		conns: make(map[string]*Conn),
	}
}

func (cr *clientRegister) Add(c *Conn) error {
	cr.mutex.Lock()
	defer cr.mutex.Unlock()

	if cr.conns[c.Login] != nil {
		return api.Error("Double registration")
	}

	cr.conns[c.Login] = c
	return nil
}

func (cr *clientRegister) Remove(c *Conn) error {
	cr.mutex.Lock()
	defer cr.mutex.Unlock()

	if c.Login != "" && cr.conns[c.Login] != nil {
		delete(cr.conns, c.Login)
		log.Println("Unregistered:", c.Login)
		return nil
	}
	return api.Error("Unregister failed")
}

func (cr *clientRegister) Exist(login string) bool {
	cr.mutex.Lock()
	defer cr.mutex.Unlock()

	return cr.conns[login] != nil
}

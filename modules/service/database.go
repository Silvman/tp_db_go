package service

import (
	"github.com/jackc/pgx"
	"log"
	"sync"
	"time"
)

type HandlerDB struct {
	pool        *pgx.ConnPool
	needVacuum  bool
	vacuumMutex *sync.Mutex
}

func (self *HandlerDB) Connect(config pgx.ConnConfig) (err error) {
	self.pool, err = pgx.NewConnPool(pgx.ConnPoolConfig{ConnConfig: config})

	tick := time.NewTicker(30 * time.Second)
	go func() {
		for range tick.C {
			log.Println("vacuum")
			self.Vacuum()
		}
	}()

	return err
}

func NewHandler() (*HandlerDB, error) {
	config := pgx.ConnConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "docker",
		Password: "docker",
		Database: "docker",
	}

	handler := &HandlerDB{
		vacuumMutex: &sync.Mutex{},
		needVacuum:  true,
	}
	var err = handler.Connect(config)

	return handler, err
}

func check(err interface{}) {
	//log.Println(err)
}

func (self *HandlerDB) askVacuum() {
	//self.vacuumMutex.Lock()
	//self.needVacuum = true
	//self.vacuumMutex.Unlock()
}

func (self *HandlerDB) checkVacuum() {
	//self.vacuumMutex.Lock()
	//if self.needVacuum {
	//	self.pool.Exec("VACUUM");
	//	self.vacuumMutex.Lock()
	//	self.needVacuum = false
	//	self.vacuumMutex.Unlock()
	//}

	//self.vacuumMutex.Unlock()
}

func (self *HandlerDB) Vacuum() {
	self.pool.Exec("VACUUM")
}

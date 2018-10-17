package service

import (
	"github.com/jackc/pgx"
	"log"
)

type HandlerDB struct {
	pool *pgx.ConnPool
}

func (self *HandlerDB) Connect(config pgx.ConnConfig) (err error) {
	self.pool, err = pgx.NewConnPool(pgx.ConnPoolConfig{ConnConfig: config})
	return err
}

func NewHandler() (*HandlerDB, error) {
	config := pgx.ConnConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "test",
		Password: "test",
		Database: "test",
	}

	handler := &HandlerDB{}
	var err = handler.Connect(config)

	return handler, err
}

func check(err error) {
	if err != nil {
		log.Panic(err)
	}
}

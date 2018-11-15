package service

import (
	"github.com/Silvman/tech-db-forum/models"
	"github.com/Silvman/tech-db-forum/restapi/operations"
	"github.com/go-openapi/runtime/middleware"
	"log"
)

func (self HandlerDB) Clear(params operations.ClearParams) middleware.Responder {
	tx, err := self.pool.Begin()
	if err != nil {
		log.Fatalln(err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`truncate table votes, forums_users, users, forums, threads, posts`); err != nil {
		check(err)
	}

	err = tx.Commit()
	if err != nil {
		check(err)
	}

	return operations.NewClearOK()
}

func (self HandlerDB) Status(params operations.StatusParams) middleware.Responder {
	tx, err := self.pool.Begin()
	if err != nil {
		log.Fatalln(err)
	}
	defer tx.Rollback()

	check("status")

	status := models.Status{}
	tx.QueryRow("select count(*) from users").Scan(&status.User)
	tx.QueryRow("select count(*) from forums").Scan(&status.Forum)
	tx.QueryRow("select count(*) from threads").Scan(&status.Thread)
	tx.QueryRow("select count(*) from posts").Scan(&status.Post)

	err = tx.Commit()
	if err != nil {
		check(err)
	}

	return operations.NewStatusOK().WithPayload(&status)
}

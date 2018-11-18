package mapper

import (
	"github.com/Silvman/tech-db-forum/models"
	"log"
)

func (self HandlerDB) Clear() error {
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

	return nil
}

func (self HandlerDB) Status() *models.Status {
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

	return &status
}

package mapper

import (
	"github.com/Silvman/tech-db-forum/models"
	"log"
	"sync/atomic"
)

func (self HandlerDB) Clear() error {
	if _, err := self.pool.Exec(`truncate table votes, forums_users, users, forums, threads, posts`); err != nil {
		log.Println(err)
	}

	atomic.StoreInt32(&totalPosts, 0)

	return nil
}

func (self HandlerDB) Status() *models.Status {
	status := models.Status{}
	self.pool.QueryRow("select count(*) from users").Scan(&status.User)
	self.pool.QueryRow("select count(*) from forums").Scan(&status.Forum)
	self.pool.QueryRow("select count(*) from threads").Scan(&status.Thread)
	self.pool.QueryRow("select count(*) from posts").Scan(&status.Post)

	return &status
}

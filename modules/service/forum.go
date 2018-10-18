package service

import (
	"fmt"
	"github.com/Silvman/tech-db-forum/models"
	"github.com/Silvman/tech-db-forum/restapi/operations"
	"github.com/go-openapi/runtime/middleware"
	"log"
)

const qSelectForumBySlug = `SELECT slug, title, posts, threads, owner FROM forums WHERE slug=$1`
const qInsertForum = `INSERT INTO forums (slug, title, owner) VALUES ($1, $2, $3) RETURNING owner`

/*
func (self HandlerDB) () middleware.Responder {

}

	tx, err := self.pool.Begin()
	if err != nil {
		log.Fatalln(err)
	}
	defer tx.Rollback()

	err = tx.Commit()
	if err != nil {
    	log.Println(err)
	}
*/

func (self HandlerDB) ForumCreate(params operations.ForumCreateParams) middleware.Responder {
	tx, err := self.pool.Begin()
	if err != nil {
		log.Fatalln(err)
	}
	defer tx.Rollback()

	forumExisting := models.Forum{}

	if err = tx.QueryRow(qSelectForumBySlug, &params.Forum.Slug).
		Scan(
			&forumExisting.Slug,
			&forumExisting.Title,
			&forumExisting.Posts,
			&forumExisting.Threads,
			&forumExisting.User,
		); err == nil {
		return operations.NewForumCreateConflict().WithPayload(&forumExisting)
	}

	if err := tx.QueryRow(qInsertForum, &params.Forum.Slug, &params.Forum.Title, &params.Forum.User).
		Scan(&params.Forum.User); err != nil {
		currError := models.Error{Message: fmt.Sprintf("Can't find forum with slug: %s", params.Forum.Slug)}
		return operations.NewForumCreateNotFound().WithPayload(&currError)
	}

	tx.Commit()
	return operations.NewForumCreateCreated().WithPayload(params.Forum)
}

func (self HandlerDB) ForumGetOne(params operations.ForumGetOneParams) middleware.Responder {
	tx, err := self.pool.Begin()
	if err != nil {
		log.Fatalln(err)
	}
	defer tx.Rollback()

	forumExisting := models.Forum{}

	if err = tx.QueryRow(qSelectForumBySlug, params.Slug).
		Scan(
			&forumExisting.Slug,
			&forumExisting.Title,
			&forumExisting.Posts,
			&forumExisting.Threads,
			&forumExisting.User,
		); err != nil {
		currError := models.Error{Message: fmt.Sprintf("Can't find forum with slug: %s", params.Slug)}
		operations.NewForumGetOneNotFound().WithPayload(&currError)
	}

	tx.Commit()

	return operations.NewForumGetOneOK().WithPayload(&forumExisting)
}

func (self HandlerDB) ForumGetThreads(params operations.ForumGetThreadsParams) middleware.Responder {
	tx, err := self.pool.Begin()
	if err != nil {
		log.Fatalln(err)
	}
	defer tx.Rollback()

	if rows, _ := tx.Query(qSelectForumBySlug, params.Slug); rows.Next() == false {
		currError := models.Error{Message: fmt.Sprintf("Can't find forum with slug: %s", params.Slug)}
		return operations.NewForumGetThreadsNotFound().WithPayload(&currError)
	}

	args := []interface{}{}
	query := `select (id, title, message, votes, slug, created, forum, author) from threads where slug = $1`
	args = append(args, params.Slug)

	if params.Since != nil {
		args = append(args, *params.Since)
		query += fmt.Sprintf(" and created >= $%d", len(args))
	}

	if params.Desc != nil {
		query += " order by created"
		if *params.Desc {
			query += " desc"
		} else {
			query += " asc"
		}
	}

	if params.Limit != nil {
		args = append(args, *params.Limit)
		query += fmt.Sprintf(" limit $%d", len(args))
	}

	rows, err := tx.Query(query, args...)
	existingThreads := models.Threads{}

	for rows.Next() {
		t := models.Thread{}
		rows.Scan(&t.ID, &t.Title, &t.Message, &t.Votes, &t.Slug, &t.Created, &t.Forum, &t.Author)
		existingThreads = append(existingThreads, &t)
	}
	tx.Commit()

	return operations.NewForumGetThreadsOK().WithPayload(existingThreads)
}

func (self HandlerDB) ForumGetUsers(params operations.ForumGetUsersParams) middleware.Responder {
	tx, err := self.pool.Begin()
	if err != nil {
		log.Fatalln(err)
	}
	defer tx.Rollback()

	if rows, _ := tx.Query(qSelectForumBySlug, params.Slug); rows.Next() == false {
		currError := models.Error{Message: fmt.Sprintf("Can't find forum with slug: %s", params.Slug)}
		return operations.NewForumGetThreadsNotFound().WithPayload(&currError)
	}

	args := []interface{}{}
	args = append(args, params.Slug)

	qOuter := `select (nickname, fullname, about, email) from users where nickname in (`
	qSelectUserThreads := `select distinct (author) from threads where forum = $1`
	qSelectUserPosts := `select distinct (p.author) from posts p`
	qJoin := `join threads t on p.thread = t.id where t.forum = $1`
	qOuterClose := `) order by nickname`

	if params.Since != nil {
		args = append(args, *params.Since)
		qSelectUserThreads += fmt.Sprintf(" and id > $%d", len(args))
		qSelectUserPosts += fmt.Sprintf(" and p.id > $%d", len(args))
	}

	if params.Desc != nil {
		if *params.Desc {
			qOuterClose += " desc"
		} else {
			qOuterClose += " asc"
		}
	}

	if params.Limit != nil {
		args = append(args, *params.Limit)
		qOuterClose += fmt.Sprintf(" limit $%d", len(args))
	}

	query := qOuter + qSelectUserThreads + " union " + qSelectUserPosts + " " + qJoin + qOuterClose

	rows, err := tx.Query(query, args...)
	existingUsers := models.Users{}

	for rows.Next() {
		t := models.User{}
		rows.Scan(&t.Nickname, &t.Fullname, &t.About, &t.Email)
		existingUsers = append(existingUsers, &t)
	}
	tx.Commit()

	return operations.NewForumGetUsersOK().WithPayload(existingUsers)

}

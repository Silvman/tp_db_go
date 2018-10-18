package service

import (
	"fmt"
	"github.com/Silvman/tech-db-forum/models"
	"github.com/Silvman/tech-db-forum/restapi/operations"
	"github.com/go-openapi/runtime/middleware"
	"log"
	"strconv"
	"strings"
)

// TODO УДАЛИТЬ ВСЕ id = $1 or slug = $1

func (self HandlerDB) ThreadCreate(params operations.ThreadCreateParams) middleware.Responder {
	tx, err := self.pool.Begin()
	if err != nil {
		log.Fatalln(err)
	}
	defer tx.Rollback()

	var slug string
	if err := tx.QueryRow("select slug from forums where slug = $1", params.Slug).Scan(&slug); err != nil {
		currentErr := models.Error{Message: fmt.Sprintf("Can't find thread forum by slug: %s", params.Slug)}
		return operations.NewThreadCreateNotFound().WithPayload(&currentErr)
	}

	var author string
	if err := tx.QueryRow("select nickname from users where nickname = $1", params.Thread.Author).Scan(&author); err != nil {
		currentErr := models.Error{Message: fmt.Sprintf("Can't find thread author by nickname: %s", params.Thread.Author)}
		return operations.NewThreadCreateNotFound().WithPayload(&currentErr)
	}

	eThread := models.Thread{}
	if err := tx.QueryRow(`select id, title, message, votes, slug, created, forum, author from threads where title = $1 and forum = $2`, params.Thread.Title, params.Slug).
		Scan(
			&eThread.ID,
			&eThread.Title,
			&eThread.Message,
			&eThread.Votes,
			&eThread.Slug,
			&eThread.Created,
			&eThread.Forum,
			&eThread.Author,
		); err == nil {
		return operations.NewThreadCreateConflict().WithPayload(&eThread)
	} else {
		log.Println(err)
	}

	args := []interface{}{}

	qFields := "title, message, author"
	qValues := "$1, $2, $3, $4"
	args = append(args, params.Thread.Title, params.Thread.Message, params.Thread.Author, params.Slug)

	if params.Thread.Created != nil {
		args = append(args, params.Thread.Created)
		qFields += ", created"
		qValues += fmt.Sprintf(", $%d", len(args))
	}

	if params.Thread.Slug != "" {
		args = append(args, params.Thread.Slug)
		qFields += ", slug"
		qValues += fmt.Sprintf(", $%d", len(args))
	}

	query := "insert into threads (" + qFields + ") values (" + qValues + ") returning id, title, message, votes, slug, created, forum, author"

	newThread := models.Thread{}
	if err := tx.QueryRow(query, args...).
		Scan(
			&newThread.ID,
			&newThread.Title,
			&newThread.Message,
			&newThread.Votes,
			&newThread.Slug,
			&newThread.Created,
			&newThread.Forum,
			&newThread.Author,
		); err != nil {
		log.Println(err)
	}

	err = tx.Commit()
	if err != nil {
		log.Println(err)
	}

	return operations.NewThreadCreateCreated().WithPayload(&newThread)
}

func (self HandlerDB) ThreadGetOne(params operations.ThreadGetOneParams) middleware.Responder {
	tx, err := self.pool.Begin()
	if err != nil {
		log.Fatalln(err)
	}
	defer tx.Rollback()

	eThread := models.Thread{}
	if err := tx.QueryRow(`select id, title, message, votes, slug, created, forum, author from threads where id = $1 or slug = $1`, params.SlugOrID).
		Scan(
			&eThread.ID,
			&eThread.Title,
			&eThread.Message,
			&eThread.Votes,
			&eThread.Slug,
			&eThread.Created,
			&eThread.Forum,
			&eThread.Author,
		); err != nil {
		currentErr := models.Error{}
		if _, err := strconv.Atoi(params.SlugOrID); err != nil {
			currentErr.Message = fmt.Sprintf("Can't find thread by slug: %s", params.SlugOrID)
		} else {
			currentErr.Message = fmt.Sprintf("Can't find thread by id: %s", params.SlugOrID)
		}
		return operations.NewThreadGetPostsNotFound().WithPayload(&currentErr)
	}

	err = tx.Commit()
	if err != nil {
		log.Println(err)
	}

	return operations.NewThreadGetOneOK().WithPayload(&eThread)
}

func (self HandlerDB) ThreadGetPosts(params operations.ThreadGetPostsParams) middleware.Responder {

	return operations.NewThreadGetPostsOK()
}

func (self HandlerDB) ThreadUpdate(params operations.ThreadUpdateParams) middleware.Responder {
	tx, err := self.pool.Begin()
	if err != nil {
		log.Fatalln(err)
	}
	defer tx.Rollback()

	query := "select id from threads where"
	currentErr := models.Error{}

	if _, err := strconv.Atoi(params.SlugOrID); err != nil {
		currentErr.Message = fmt.Sprintf("Can't find thread by slug: %s", params.SlugOrID)
		query += " slug = $1"
	} else {
		currentErr.Message = fmt.Sprintf("Can't find thread by id: %s", params.SlugOrID)
		query += " id = $1"
	}

	var tId int32
	if err := tx.QueryRow(query, params.SlugOrID).Scan(&tId); err != nil {
		return operations.NewThreadUpdateNotFound().WithPayload(&currentErr)
	}

	args := []interface{}{}

	var qValues []string

	if params.Thread.Message != "" {
		args = append(args, params.Thread.Message)
		qValues = append(qValues, fmt.Sprintf("message = $%d", len(args)))
	}

	if params.Thread.Title != "" {
		args = append(args, params.Thread.Title)
		qValues = append(qValues, fmt.Sprintf("title = $%d", len(args)))
	}

	updThread := models.Thread{}
	query = fmt.Sprintf("update threads set "+strings.Join(qValues, ",")+" where id = %d returning id, title, message, votes, slug, created, forum, author", tId)
	if err := tx.QueryRow(query, args...).Scan(
		&updThread.ID,
		&updThread.Title,
		&updThread.Message,
		&updThread.Votes,
		&updThread.Slug,
		&updThread.Created,
		&updThread.Forum,
		&updThread.Author,
	); err != nil {
		log.Println(err)
	}

	if err = tx.Commit(); err != nil {
		log.Println(err)
	}

	return operations.NewThreadUpdateOK().WithPayload(&updThread)
}

func (self HandlerDB) ThreadVote(params operations.ThreadVoteParams) middleware.Responder {
	tx, err := self.pool.Begin()
	if err != nil {
		log.Fatalln(err)
	}
	defer tx.Rollback()

	query := "select id from threads where"
	currentErr := models.Error{}
	if _, err := strconv.Atoi(params.SlugOrID); err != nil {
		currentErr.Message = fmt.Sprintf("Can't find thread by slug: %s", params.SlugOrID)
		query += " slug = $1"
	} else {
		currentErr.Message = fmt.Sprintf("Can't find thread by id: %s", params.SlugOrID)
		query += " id = $1"
	}

	var tId int32
	if err := tx.QueryRow(query, params.SlugOrID).Scan(&tId); err != nil {
		return operations.NewThreadVoteNotFound().WithPayload(&currentErr)
	}

	if _, err := tx.Exec("insert into votes (author, thread, vote) values ($1, $2, $3) on conflict do update set vote = $3", params.Vote.Nickname, tId, params.Vote.Voice); err != nil {
		currentErr.Message = fmt.Sprintf("Can't find user by nickname: %s", params.Vote.Nickname)
		return operations.NewThreadVoteNotFound().WithPayload(&currentErr)
	}

	updThread := models.Thread{}
	if err := tx.QueryRow(`select id, title, message, votes, slug, created, forum, author from threads where id = $1`, tId).
		Scan(
			&updThread.ID,
			&updThread.Title,
			&updThread.Message,
			&updThread.Votes,
			&updThread.Slug,
			&updThread.Created,
			&updThread.Forum,
			&updThread.Author,
		); err != nil {
		log.Println(err)
	}

	if err = tx.Commit(); err != nil {
		log.Println(err)
	}

	return operations.NewThreadUpdateOK().WithPayload(&updThread)

}

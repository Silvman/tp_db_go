package service

import (
	"fmt"
	"github.com/go-openapi/runtime/middleware"
	"log"
	"strconv"
	"github.com/Silvman/tech-db-forum/models"
	"github.com/Silvman/tech-db-forum/restapi/operations"
)

func (self HandlerDB) PostGetOne(params operations.PostGetOneParams) middleware.Responder {
	tx, err := self.pool.Begin()
	if err != nil {
		log.Fatalln(err)
	}
	defer tx.Rollback()

	ePostFull := models.PostFull{}
	if err := tx.QueryRow(`select id, parent, message, isEdit, forum, created, thread, author from posts where id = $1`, params.ID).
		Scan(
			&ePostFull.Post.ID,
			&ePostFull.Post.Parent,
			&ePostFull.Post.Message,
			&ePostFull.Post.IsEdited,
			&ePostFull.Post.Forum,
			&ePostFull.Post.Created,
			&ePostFull.Post.Thread,
			&ePostFull.Post.Author,
		); err != nil {
		currentErr := models.Error{Message: fmt.Sprintf("Can't find post with id: %s", params.ID)}
		return operations.NewPostGetOneNotFound().WithPayload(&currentErr)
	}

	if params.Related != nil {
		for _, value := range params.Related {
			switch value {
			case "user":
				{
					if err := tx.QueryRow(`select nickname, fullname, about, email from users where nickname = $1`, ePostFull.Post.Author).
						Scan(&ePostFull.Author.Nickname, &ePostFull.Author.Fullname, &ePostFull.Author.About, &ePostFull.Author.Email); err != nil {
						log.Println(err)
					}
				}

			case "forum":
				{
					if err := tx.QueryRow(qSelectForumBySlug, ePostFull.Post.Forum).
						Scan(
							&ePostFull.Forum.Slug,
							&ePostFull.Forum.Title,
							&ePostFull.Forum.Posts,
							&ePostFull.Forum.Threads,
							&ePostFull.Forum.User,
						); err != nil {
						log.Println(err)
					}
				}

			case "thread":
				{
					if err := tx.QueryRow(`select id, title, message, votes, slug, created, forum, author from threads where id = $1`, ePostFull.Post.Thread).
						Scan(
							&ePostFull.Thread.ID,
							&ePostFull.Thread.Title,
							&ePostFull.Thread.Message,
							&ePostFull.Thread.Votes,
							&ePostFull.Thread.Slug,
							&ePostFull.Thread.Created,
							&ePostFull.Thread.Forum,
							&ePostFull.Thread.Author,
						); err != nil {
						log.Println(err)
					}
				}

			}
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Println(err)
	}

	return operations.NewPostGetOneOK().WithPayload(&ePostFull)
}

func (self HandlerDB) PostUpdate(params operations.PostUpdateParams) middleware.Responder {
	tx, err := self.pool.Begin()
	if err != nil {
		log.Fatalln(err)
	}
	defer tx.Rollback()

	ePost := models.Post{}
	if err := tx.QueryRow(`update posts set message = $1, isEdit = true where id = $2 returning id, parent, message, isEdit, forum, created, thread, author`, params.Post.Message, params.ID).
		Scan(
			&ePost.ID,
			&ePost.Parent,
			&ePost.Message,
			&ePost.IsEdited,
			&ePost.Forum,
			&ePost.Created,
			&ePost.Thread,
			&ePost.Author,
		); err != nil {
		currentErr := models.Error{Message:fmt.Sprintf("Can't find post with id: %d", params.ID)}
		return operations.NewPostUpdateNotFound().WithPayload(&currentErr)
	}

	err = tx.Commit()
	if err != nil {
		log.Println(err)
	}

	return operations.NewPostUpdateOK().WithPayload(&ePost)
}

func (self HandlerDB) PostsCreate(params operations.PostsCreateParams) middleware.Responder {
	tx, err := self.pool.Begin()
	if err != nil {
		log.Fatalln(err)
	}
	defer tx.Rollback()

	// проверки slug или id
	// заселектить тред со слагом\айди. получить айди и дальше с ним работать

	query := `with now_time as (select current_time as ct)
insert into posts (parent, message, isEdit, created, thread, author) values ($1, $2, false, now_time.ct, $3, $4)`
	queryEnd := " returning id, parent, message, isEdit, created, thread, author"


	args := []interface{}{}
	post1 := params.Posts[0]
	args = append(args, post1.Parent, post1.Message, post1.Author, params.SlugOrID)

	for _, value := range params.Posts {
		if value.Parent != 0 {
			var pId int64
			var tId int32

			if err := tx.QueryRow(`select id, thread from posts where id = $1`, value.Parent, tId).Scan(&pId); err != nil {
				currentErr := models.Error{Message:fmt.Sprintf("Parent post is not found")}
				return operations.NewPostsCreateNotFound().WithPayload(&currentErr)
			}

		}

		query += fmt.Sprintf(", ($%d, $%d, false, now_time.ct, $%d, $%d)", len(args) + 1, len(args) + 2, len(args) + 3, len(args) + 4)
		args = append(args, value.Parent, value.Message, value.Author, params.SlugOrID)
	}

	posts := models.Posts{}
	if rows, err :=

	err = tx.Commit()
	if err != nil {
		log.Println(err)
	}

	return operations.NewPostsCreateCreated()
}

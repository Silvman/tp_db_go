package service

import (
	"fmt"
	"github.com/Silvman/tech-db-forum/models"
	"github.com/Silvman/tech-db-forum/restapi/operations"
	"github.com/go-openapi/runtime/middleware"
	"log"
	"strconv"
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
		currentErr := models.Error{Message: fmt.Sprintf("Can't find post with id: %d", params.ID)}
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

	var tIdCurrent int32
	if err := tx.QueryRow("select id, slug form threads where id = $1 or slug = $1", params.SlugOrID).Scan(&tIdCurrent); err != nil {
		currentErr := models.Error{}
		if _, err := strconv.Atoi(params.SlugOrID); err != nil {
			currentErr.Message = fmt.Sprintf("Can't find post thread by slug: %s", params.SlugOrID)
		} else {
			currentErr.Message = fmt.Sprintf("Can't find post thread by id: %s", params.SlugOrID)
		}

		return operations.NewPostsCreateNotFound().WithPayload(&currentErr)
	}

	query := `with now_time as (select current_time as ct) insert into posts (parent, message, created, thread, author) values ($1, $2, now_time.ct, $3, $4)`
	queryEnd := " returning id, parent, message, isEdit, created, thread, author"

	args := []interface{}{}
	post1 := params.Posts[0]
	args = append(args, post1.Parent, post1.Message, post1.Author, tIdCurrent)

	for key, value := range params.Posts {
		if value.Parent != 0 {
			var tId int32

			// Нет в документации к апи!
			if err := tx.QueryRow(`select thread from posts where id = $1`, value.Parent).Scan(&tId); err != nil {
				currentErr := models.Error{Message: fmt.Sprintf("Parent post is not found")}
				return operations.NewPostsCreateNotFound().WithPayload(&currentErr)
			}

			if tId != tIdCurrent {
				currentErr := models.Error{Message: fmt.Sprintf("Parent post was created in another thread")}
				return operations.NewPostsCreateConflict().WithPayload(&currentErr)
			}

			var nick string
			if err := tx.QueryRow("select nickname from users where id = $1", value.Author).Scan(&nick); err != nil {
				currentErr := models.Error{Message: fmt.Sprintf("Can't find post author by nickname: %s", value.Author)}
				return operations.NewPostsCreateNotFound().WithPayload(&currentErr)
			}
		}

		if key == 0 {
			continue
		}

		query += fmt.Sprintf(", ($%d, $%d, now_time.ct, $%d, $%d)", len(args)+1, len(args)+2, len(args)+3, len(args)+4)
		args = append(args, value.Parent, value.Message, value.Author, tIdCurrent)
	}

	query += queryEnd

	posts := models.Posts{}
	rows, err := tx.Query(query, args...)
	if err != nil {
		log.Println(err)
	}

	for rows.Next() {
		t := models.Post{}
		rows.Scan(&t.ID, &t.Parent, &t.Message, &t.IsEdited, &t.Created, &t.Thread, &t.Author)
		posts = append(posts, &t)
	}

	err = tx.Commit()
	if err != nil {
		log.Println(err)
	}

	return operations.NewPostsCreateCreated().WithPayload(posts)
}

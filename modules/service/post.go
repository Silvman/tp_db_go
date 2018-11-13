package service

import (
	"fmt"
	"github.com/Silvman/tech-db-forum/models"
	"github.com/Silvman/tech-db-forum/restapi/operations"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/jackc/pgx/pgtype"
	"log"
	"strconv"
	"strings"
)

func (self HandlerDB) PostGetOne(params operations.PostGetOneParams) middleware.Responder {
	tx, err := self.pool.Begin()
	if err != nil {
		log.Fatalln(err)
	}
	defer tx.Rollback()

	ePostFull := models.PostFull{}

	ePostFull.Post = &models.Post{}

	pgTime := pgtype.Timestamptz{}

	if err := tx.QueryRow(`select id, parent, message, isEdit, forum, created, thread, author from posts where id = $1`, params.ID).
		Scan(
			&ePostFull.Post.ID,
			&ePostFull.Post.Parent,
			&ePostFull.Post.Message,
			&ePostFull.Post.IsEdited,
			&ePostFull.Post.Forum,
			&pgTime,
			&ePostFull.Post.Thread,
			&ePostFull.Post.Author,
		); err != nil {

		log.Println(err)
		currentErr := models.Error{Message: fmt.Sprintf("Can't find post with id: %s", params.ID)}
		return operations.NewPostGetOneNotFound().WithPayload(&currentErr)
	}

	t := strfmt.NewDateTime()
	t.Scan(pgTime.Time)

	ePostFull.Post.Created = &t

	if params.Related != nil {
		for _, value := range params.Related {
			switch value {
			case "user":
				{
					ePostFull.Author = &models.User{}
					if err := tx.QueryRow(`select nickname, fullname, about, email from users where nickname = $1`, ePostFull.Post.Author).
						Scan(&ePostFull.Author.Nickname, &ePostFull.Author.Fullname, &ePostFull.Author.About, &ePostFull.Author.Email); err != nil {
						log.Println(err)
					}
				}

			case "forum":
				{
					ePostFull.Forum = &models.Forum{}
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
					pgSlug := pgtype.Text{}
					ePostFull.Thread = &models.Thread{}
					if err := tx.QueryRow(`select id, title, message, votes, slug, created, forum, author from threads where id = $1`, ePostFull.Post.Thread).
						Scan(
							&ePostFull.Thread.ID,
							&ePostFull.Thread.Title,
							&ePostFull.Thread.Message,
							&ePostFull.Thread.Votes,
							&pgSlug,
							&pgTime,
							&ePostFull.Thread.Forum,
							&ePostFull.Thread.Author,
						); err != nil {
						log.Println(err)
					}

					if pgSlug.Status != pgtype.Null {
						ePostFull.Thread.Slug = pgSlug.String
					}

					t := strfmt.NewDateTime()
					t.Scan(pgTime.Time)

					ePostFull.Thread.Created = &t
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

	pgTime := pgtype.Timestamptz{}
	ePost := models.Post{}

	log.Println("post_update")

	if err := tx.QueryRow(`select id, parent, message, isEdit, forum, created, thread, author from posts where id = $1`, params.ID).Scan(
		&ePost.ID,
		&ePost.Parent,
		&ePost.Message,
		&ePost.IsEdited,
		&ePost.Forum,
		&pgTime,
		&ePost.Thread,
		&ePost.Author,
	); err != nil {
		log.Println(err)
		currentErr := models.Error{Message: fmt.Sprintf("Can't find post with id: %d", params.ID)}
		return operations.NewPostUpdateNotFound().WithPayload(&currentErr)
	}
	t := strfmt.NewDateTime()
	t.Scan(pgTime.Time)
	ePost.Created = &t

	if params.Post.Message != "" && params.Post.Message != ePost.Message {
		if err := tx.QueryRow(`update posts set isEdit = true, message = $1 where id = $2 returning id, parent, message, isEdit, forum, created, thread, author`, params.Post.Message, params.ID).Scan(
			&ePost.ID,
			&ePost.Parent,
			&ePost.Message,
			&ePost.IsEdited,
			&ePost.Forum,
			&pgTime,
			&ePost.Thread,
			&ePost.Author,
		); err != nil {
			log.Println(err)
		}
		t := strfmt.NewDateTime()
		t.Scan(pgTime.Time)
		ePost.Created = &t
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

	log.Println("posts_create")

	queryCheck := "select id, forum from threads where"
	currentErr := models.Error{}

	if _, err := strconv.Atoi(params.SlugOrID); err != nil {
		currentErr.Message = fmt.Sprintf("Can't find thread by slug: %s", params.SlugOrID)
		queryCheck += " slug = $1"
	} else {
		currentErr.Message = fmt.Sprintf("Can't find thread by id: %s", params.SlugOrID)
		queryCheck += " id = $1::bigint"
	}

	var tIdCurrent int32
	var tForumCurrent string
	if err := tx.QueryRow(queryCheck, params.SlugOrID).Scan(&tIdCurrent, &tForumCurrent); err != nil {
		return operations.NewPostsCreateNotFound().WithPayload(&currentErr)
	}

	if len(params.Posts) == 0 {
		return operations.NewPostsCreateCreated().WithPayload(params.Posts)
	}

	query := `insert into posts (parent, message, thread, author, forum) values `
	queryEnd := " returning id, parent, message, isEdit, forum, created, thread, author"
	var queryValues []string

	args := []interface{}{}
	for _, value := range params.Posts {
		var tId int32

		if value.Parent != 0 {
			// Нет в документации к апи!
			if err := tx.QueryRow(`select thread from posts where id = $1`, value.Parent).Scan(&tId); err != nil {
				currentErr := models.Error{Message: fmt.Sprintf("Parent post was created in another thread")}
				return operations.NewPostsCreateConflict().WithPayload(&currentErr)
			}

			if tId != tIdCurrent {
				currentErr := models.Error{Message: fmt.Sprintf("Parent post was created in another thread")}
				return operations.NewPostsCreateConflict().WithPayload(&currentErr)
			}
		}

		var nick string
		if err := tx.QueryRow("select nickname from users where nickname = $1", value.Author).Scan(&nick); err != nil {
			currentErr := models.Error{Message: fmt.Sprintf("Can't find post author by nickname: %s", value.Author)}
			return operations.NewPostsCreateNotFound().WithPayload(&currentErr)
		}

		queryValues = append(queryValues, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d)", len(args)+1, len(args)+2, len(args)+3, len(args)+4, len(args)+5))
		args = append(args, value.Parent, value.Message, tIdCurrent, nick, tForumCurrent)
	}

	query += strings.Join(queryValues, ",") + queryEnd

	posts := models.Posts{}

	rows, err := tx.Query(query, args...)
	if err != nil {
		log.Println(err)
	}

	for rows.Next() {
		t := models.Post{}
		pgTime := pgtype.Timestamptz{}
		err = rows.Scan(&t.ID, &t.Parent, &t.Message, &t.IsEdited, &t.Forum, &pgTime, &t.Thread, &t.Author)
		log.Println(err)

		time := strfmt.NewDateTime()
		time.Scan(pgTime.Time)

		t.Created = &time

		posts = append(posts, &t)
	}

	err = tx.Commit()
	if err != nil {
		log.Println(err)
	}

	return operations.NewPostsCreateCreated().WithPayload(posts)
}

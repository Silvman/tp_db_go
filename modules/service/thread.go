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

	pgTime := pgtype.Timestamptz{}
	pgSlug := pgtype.Text{}
	eThread := models.Thread{}
	if err := tx.QueryRow(`select id, title, message, votes, slug, created, forum, author from threads where (title = $1 and forum = $2) or (slug = $3)`, params.Thread.Title, params.Slug, params.Thread.Slug).
		Scan(
			&eThread.ID,
			&eThread.Title,
			&eThread.Message,
			&eThread.Votes,
			&pgSlug,
			&pgTime,
			&eThread.Forum,
			&eThread.Author,
		); err == nil {
		if pgSlug.Status != pgtype.Null {
			eThread.Slug = pgSlug.String
		}
		t := strfmt.NewDateTime()
		t.Scan(pgTime.Time)

		eThread.Created = &t
		return operations.NewThreadCreateConflict().WithPayload(&eThread)
	} else {
		log.Println(err)
	}

	args := []interface{}{}

	qFields := "title, message, author, forum"
	qValues := "$1, $2, $3, $4"
	args = append(args, params.Thread.Title, params.Thread.Message, params.Thread.Author, slug)

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
			&pgSlug,
			&pgTime,
			&newThread.Forum,
			&newThread.Author,
		); err != nil {

	}

	t := strfmt.NewDateTime()
	t.Scan(pgTime.Time)
	newThread.Created = &t

	if pgSlug.Status != pgtype.Null {
		newThread.Slug = pgSlug.String
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

	pgTime := pgtype.Timestamptz{}
	pgSlug := pgtype.Text{}
	eThread := models.Thread{}

	query := "select id, title, message, votes, slug, created, forum, author from threads where"
	currentErr := models.Error{}
	if _, err := strconv.Atoi(params.SlugOrID); err != nil {
		currentErr.Message = fmt.Sprintf("Can't find thread by slug: %s", params.SlugOrID)
		query += " slug = $1"
	} else {
		currentErr.Message = fmt.Sprintf("Can't find thread by id: %s", params.SlugOrID)
		query += " id = $1::bigint"
	}

	if err := tx.QueryRow(query, params.SlugOrID).
		Scan(
			&eThread.ID,
			&eThread.Title,
			&eThread.Message,
			&eThread.Votes,
			&pgSlug,
			&pgTime,
			&eThread.Forum,
			&eThread.Author,
		); err != nil {
		return operations.NewThreadGetPostsNotFound().WithPayload(&currentErr)
	}

	// парсим время и слаг
	if pgSlug.Status != pgtype.Null {
		eThread.Slug = pgSlug.String
	}
	t := strfmt.NewDateTime()
	t.Scan(pgTime.Time)

	eThread.Created = &t

	err = tx.Commit()
	if err != nil {
		log.Println(err)
	}

	return operations.NewThreadGetOneOK().WithPayload(&eThread)
}

func (self HandlerDB) ThreadGetPosts(params operations.ThreadGetPostsParams) middleware.Responder {
	tx, err := self.pool.Begin()
	if err != nil {
		log.Fatalln(err)
	}
	defer tx.Rollback()

	queryCheck := "select id from threads where"
	currentErr := models.Error{}

	if _, err := strconv.Atoi(params.SlugOrID); err != nil {
		currentErr.Message = fmt.Sprintf("Can't find thread by slug: %s", params.SlugOrID)
		queryCheck += " slug = $1"
	} else {
		currentErr.Message = fmt.Sprintf("Can't find thread by id: %s", params.SlugOrID)
		queryCheck += " id = $1::bigint"
	}

	var tId int32
	if err := tx.QueryRow(queryCheck, params.SlugOrID).Scan(&tId); err != nil {
		return operations.NewThreadGetPostsNotFound().WithPayload(&currentErr)
	}

	args := []interface{}{}
	//var qValues []string

	args = append(args, tId)

	var query string

	switch *params.Sort {
	default:
		fallthrough
	case "flat":
		{
			query = "select id, parent, message, isEdit, forum, created, thread, author from posts where thread = $1::bigint"

			if params.Since != nil {
				args = append(args, *params.Since)
				if params.Desc != nil && *params.Desc {
					query += fmt.Sprintf(" and id < $%d", len(args))
				} else {
					query += fmt.Sprintf(" and id > $%d", len(args))
				}
			}

			if params.Desc != nil && *params.Desc {
				query += " order by id desc"
			} else {
				query += " order by id"
			}

			if params.Limit != nil {
				args = append(args, *params.Limit)
				query += fmt.Sprintf(" limit $%d", len(args))
			}
		}

	case "tree":
		{
			query = `with recursive posts_tree_b (id, mPath) as (
    		select id, array_append('{}'::bigint[], id) as mArray from posts where parent = 0 and thread = $1
		union all
    		select p.id, array_append(mPath, p.id) from posts p
  			join posts_tree_b as pt on pt.id = p.parent
		)
		select posts_tree_b.id as id, parent, message, isEdit, forum, created, thread, author from posts_tree_b
		join posts p on posts_tree_b.id = p.id`

			if params.Since != nil {
				args = append(args, *params.Since)
				if params.Desc != nil && *params.Desc {
					query += fmt.Sprintf(" where mPath < (select mPath from posts_tree_b where id = $%d) ", len(args))
				} else {
					query += fmt.Sprintf(" where mPath > (select mPath from posts_tree_b where id = $%d) ", len(args))
				}
			}

			if params.Desc != nil && *params.Desc {
				query += " order by mPath[1] desc, mPath desc"
			} else {
				query += " order by mPath[1], mPath"
			}

			if params.Limit != nil {
				args = append(args, *params.Limit)
				query += fmt.Sprintf(" limit $%d", len(args))
			}
		}

	case "parent_tree":
		{
			query = `select rr.id, parent, message, isEdit, forum, created, thread, author from (
		with recursive posts_tree (id, mPath) as (
    			select
    			id, array_append('{}'::bigint[], id) as mArray from posts where parent = 0 and thread = $1
union all
  			select p.id, array_append(mPath, p.id) from posts p
  			join posts_tree as pt on pt.id = p.parent
		) select mPath, posts_tree.id as id, dense_rank() over (order by mPath[1]`

			if params.Desc != nil && *params.Desc {
				query += " desc"
			}

			query += `) as r from posts_tree `

			if params.Since != nil {
				args = append(args, *params.Since)
				if params.Desc != nil && *params.Desc {
					query += fmt.Sprintf(" where mPath[1] < (select mPath[1] from posts_tree where id = $%d) ", len(args))
				} else {
					query += fmt.Sprintf(" where mPath[1] > (select mPath[1] from posts_tree where id = $%d) ", len(args))
				}
			}

			query += `) as rr join posts on rr.id = posts.id where`

			if params.Limit != nil {
				args = append(args, *params.Limit)
				query += fmt.Sprintf(" r <= $%d", len(args))
			}

			if params.Desc != nil && *params.Desc {
				query += " order by mPath[1] desc, mPath"
			} else {
				query += " order by mPath[1], mPath"
			}
		}
	}

	log.Println(query)
	rows, err := tx.Query(query, args...)
	if err != nil {
		log.Println(err)
	}

	log.Printf("%#v\n", rows)

	fetchPosts := models.Posts{}
	pgTime := pgtype.Timestamptz{}
	pgSlug := pgtype.Text{}
	for rows.Next() {
		log.Printf("fit\n")
		post := models.Post{}
		err := rows.Scan(&post.ID, &post.Parent, &post.Message, &post.IsEdited, &pgSlug, &pgTime, &post.Thread, &post.Author)
		if err != nil {
			log.Println(err)
		}

		if pgSlug.Status != pgtype.Null {
			post.Forum = pgSlug.String
		}

		t := strfmt.NewDateTime()
		t.Scan(pgTime.Time)

		post.Created = &t

		fetchPosts = append(fetchPosts, &post)
	}

	if err = tx.Commit(); err != nil {
		log.Println(err)
	}

	return operations.NewThreadGetPostsOK().WithPayload(fetchPosts)
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
		query += " id = $1::bigint"
	}

	var tId int32
	if err := tx.QueryRow(query, params.SlugOrID).Scan(&tId); err != nil {
		return operations.NewThreadUpdateNotFound().WithPayload(&currentErr)
	}

	if params.Thread.Message == "" && params.Thread.Title == "" {
		newParams := operations.NewThreadGetOneParams()
		newParams.SlugOrID = params.SlugOrID
		return self.ThreadGetOne(newParams)
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

	pgTime := pgtype.Timestamptz{}
	pgSlug := pgtype.Text{}
	updThread := models.Thread{}
	query = fmt.Sprintf("update threads set "+strings.Join(qValues, ",")+" where id = %d returning id, title, message, votes, slug, created, forum, author", tId)
	if err := tx.QueryRow(query, args...).Scan(
		&updThread.ID,
		&updThread.Title,
		&updThread.Message,
		&updThread.Votes,
		&pgSlug,
		&pgTime,
		&updThread.Forum,
		&updThread.Author,
	); err != nil {
		log.Println(err)
	}

	if pgSlug.Status != pgtype.Null {
		updThread.Slug = pgSlug.String
	}
	t := strfmt.NewDateTime()
	t.Scan(pgTime.Time)

	updThread.Created = &t

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
		query += " id = $1::bigint"
	}

	var tId int32
	if err := tx.QueryRow(query, params.SlugOrID).Scan(&tId); err != nil {
		return operations.NewThreadVoteNotFound().WithPayload(&currentErr)
	}

	if _, err := tx.Exec("insert into votes (author, thread, vote) values ($1, $2, $3) on conflict (author, thread) do update set vote = $3", params.Vote.Nickname, tId, params.Vote.Voice); err != nil {
		log.Println(err)
		currentErr.Message = fmt.Sprintf("Can't find user by nickname: %s", params.Vote.Nickname)
		return operations.NewThreadVoteNotFound().WithPayload(&currentErr)
	}

	pgTime := pgtype.Timestamptz{}
	pgSlug := pgtype.Text{}
	updThread := models.Thread{}
	if err := tx.QueryRow(`select id, title, message, votes, slug, created, forum, author from threads where id = $1`, tId).
		Scan(
			&updThread.ID,
			&updThread.Title,
			&updThread.Message,
			&updThread.Votes,
			&pgSlug,
			&pgTime,
			&updThread.Forum,
			&updThread.Author,
		); err != nil {
		log.Println(err)
	}

	if pgSlug.Status != pgtype.Null {
		updThread.Slug = pgSlug.String
	}
	t := strfmt.NewDateTime()
	t.Scan(pgTime.Time)

	updThread.Created = &t

	if err = tx.Commit(); err != nil {
		log.Println(err)
	}

	return operations.NewThreadUpdateOK().WithPayload(&updThread)
}

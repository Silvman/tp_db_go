package service

import (
	"fmt"
	"github.com/Silvman/tech-db-forum/models"
	"github.com/Silvman/tech-db-forum/restapi/operations"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/jackc/pgx/pgtype"
	"strconv"
	"strings"
)

func (self HandlerDB) ThreadCreate(params operations.ThreadCreateParams) middleware.Responder {
	tx, err := self.pool.Begin()
	if err != nil {
		check(err)
	}
	defer tx.Rollback()

	check("thread_create")

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

	argsC := []interface{}{}
	argsC = append(argsC, params.Thread.Title, params.Slug, params.Thread.Message)

	queryConflict := "select id, title, message, votes, slug, created, forum, author from threads where ((title = $1) and (forum = $2) and (message = $3))"

	if params.Thread.Slug != "" {
		queryConflict += " or (slug = $4)"
		argsC = append(argsC, params.Thread.Slug)
	}

	pgTime := pgtype.Timestamptz{}
	pgSlug := pgtype.Text{}
	eThread := models.Thread{}
	if err := tx.QueryRow(queryConflict, argsC...).
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
		check(err)
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
		check(err)
	}

	t := strfmt.NewDateTime()
	t.Scan(pgTime.Time)
	newThread.Created = &t

	if pgSlug.Status != pgtype.Null {
		newThread.Slug = pgSlug.String
	}

	err = tx.Commit()
	if err != nil {
		check(err)
	}

	return operations.NewThreadCreateCreated().WithPayload(&newThread)
}

func (self HandlerDB) ThreadGetOne(params operations.ThreadGetOneParams) middleware.Responder {
	self.checkVacuum()
	tx, err := self.pool.Begin()
	if err != nil {
		check(err)
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
		check(err)
	}

	return operations.NewThreadGetOneOK().WithPayload(&eThread)
}

func (self HandlerDB) ThreadGetPosts(params operations.ThreadGetPostsParams) middleware.Responder {
	tx, err := self.pool.Begin()
	if err != nil {
		check(err)
	}
	defer tx.Rollback()

	queryCheck := "select id from threads where"
	currentErr := models.Error{}

	if _, err := strconv.Atoi(params.SlugOrID); err != nil {
		currentErr.Message = fmt.Sprintf("Can't find thread by slug: %s", params.SlugOrID)
		queryCheck += " slug = $1"
	} else {
		currentErr.Message = fmt.Sprintf("Can't find thread by id: %s", params.SlugOrID)
		queryCheck += " id = $1"
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
			query = "select id, parent, message, isEdit, forum, created, thread, author from posts where thread = $1"

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
			query = `select id, parent, message, isEdit, forum, created, thread, author from posts where thread = $1`

			if params.Since != nil {
				args = append(args, *params.Since)
				if params.Desc != nil && *params.Desc {
					query += fmt.Sprintf(" and mPath < (select mPath from posts where id = $%d) ", len(args))
				} else {
					query += fmt.Sprintf(" and mPath > (select mPath from posts where id = $%d) ", len(args))
				}
			}

			if params.Desc != nil && *params.Desc {
				query += " order by mPath desc"
			} else {
				query += " order by mPath"
			}

			if params.Limit != nil {
				args = append(args, *params.Limit)
				query += fmt.Sprintf(" limit $%d", len(args))
			}
		}

	case "parent_tree":
		{
			query = `select id, parent, message, isEdit, forum, created, thread, author
			from posts where mPath[1] in
			(select id from posts where thread = $1 and parent = 0`

			if params.Since != nil {
				args = append(args, *params.Since)
				if params.Desc != nil && *params.Desc {
					query += fmt.Sprintf(" and id < (select mPath[1] from posts where id = $%d) ", len(args))
				} else {
					query += fmt.Sprintf(" and id > (select mPath[1] from posts where id = $%d) ", len(args))
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

			query += `)`

			if params.Desc != nil && *params.Desc {
				query += " order by mPath[1] desc, mPath"
			} else {
				query += " order by mPath"
			}
		}
	}

	//check("- begin -----")
	//check(params.HTTPRequest.URL)
	//check(query)
	//check("- end -------")

	check(query)
	rows, err := tx.Query(query, args...)
	if err != nil {
		check(err)
	}

	//log.Printf("%#v\n", rows)

	fetchPosts := models.Posts{}
	pgTime := pgtype.Timestamptz{}
	pgSlug := pgtype.Text{}
	for rows.Next() {
		//log.Printf("fit\n")
		post := models.Post{}
		err := rows.Scan(&post.ID, &post.Parent, &post.Message, &post.IsEdited, &pgSlug, &pgTime, &post.Thread, &post.Author)
		if err != nil {
			check(err)
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
		check(err)
	}

	return operations.NewThreadGetPostsOK().WithPayload(fetchPosts)
}

func (self HandlerDB) ThreadUpdate(params operations.ThreadUpdateParams) middleware.Responder {
	tx, err := self.pool.Begin()
	if err != nil {
		check(err)
	}
	defer tx.Rollback()

	check("thread_update")

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
		check(err)
	}

	if pgSlug.Status != pgtype.Null {
		updThread.Slug = pgSlug.String
	}
	t := strfmt.NewDateTime()
	t.Scan(pgTime.Time)

	updThread.Created = &t

	if err = tx.Commit(); err != nil {
		check(err)
	}

	return operations.NewThreadUpdateOK().WithPayload(&updThread)
}

func (self HandlerDB) ThreadVote(params operations.ThreadVoteParams) middleware.Responder {
	tx, err := self.pool.Begin()
	if err != nil {
		check(err)
	}
	defer tx.Rollback()

	check("vote")

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
		check(err)
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
		check(err)
	}

	if pgSlug.Status != pgtype.Null {
		updThread.Slug = pgSlug.String
	}
	t := strfmt.NewDateTime()
	t.Scan(pgTime.Time)

	updThread.Created = &t

	if err = tx.Commit(); err != nil {
		check(err)
	}

	return operations.NewThreadUpdateOK().WithPayload(&updThread)
}

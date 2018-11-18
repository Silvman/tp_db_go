package mapper

import (
	"errors"
	"fmt"
	"github.com/Silvman/tech-db-forum/models"
	"github.com/jackc/pgx/pgtype"
	"strconv"
	"strings"
)

func (self HandlerDB) ThreadCreate(Slug string, Thread *models.Thread) (*models.Thread, error) {
	tx, err := self.pool.Begin()
	if err != nil {
		check(err)
	}
	defer tx.Rollback()

	check("thread_create")

	var slug string
	if err := tx.QueryRow("select slug from forums where slug = $1", Slug).Scan(&slug); err != nil {
		return nil, errors.New(fmt.Sprintf("Can't find thread forum by slug: %s", Slug))
	}

	var author string
	if err := tx.QueryRow("select nickname from users where nickname = $1", Thread.Author).Scan(&author); err != nil {
		return nil, errors.New(fmt.Sprintf("Can't find thread author by nickname: %s", Thread.Author))
	}

	argsC := []interface{}{}
	argsC = append(argsC, Thread.Title, Slug, Thread.Message)
	queryConflict := "select id, title, message, votes, slug, created, forum, author from threads where ((title = $1) and (forum = $2) and (message = $3))"

	if Thread.Slug != "" {
		queryConflict += " or (slug = $4)"
		argsC = append(argsC, Thread.Slug)
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

		tgtimeToString(&pgTime, &eThread.Created)

		return &eThread, errors.New("already exists")
	}

	args := []interface{}{}

	qFields := "title, message, author, forum"
	qValues := "$1, $2, $3, $4"
	args = append(args, Thread.Title, Thread.Message, Thread.Author, slug)

	if Thread.Created != "" {
		args = append(args, Thread.Created)
		qFields += ", created"
		qValues += fmt.Sprintf(", $%d", len(args))
	}

	if Thread.Slug != "" {
		args = append(args, Thread.Slug)
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

	tgtimeToString(&pgTime, &newThread.Created)

	if pgSlug.Status != pgtype.Null {
		newThread.Slug = pgSlug.String
	}

	err = tx.Commit()
	if err != nil {
		check(err)
	}

	return &newThread, nil
}

func (self HandlerDB) ThreadGetOne(SlugOrID string) (*models.Thread, error) {
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
	var currentErr string
	if _, err := strconv.Atoi(SlugOrID); err != nil {
		currentErr = fmt.Sprintf("Can't find thread by slug: %s", SlugOrID)
		query += " slug = $1"
	} else {
		currentErr = fmt.Sprintf("Can't find thread by id: %s", SlugOrID)
		query += " id = $1::bigint"
	}

	if err := tx.QueryRow(query, SlugOrID).
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
		return nil, errors.New(currentErr)
	}

	// парсим время и слаг
	if pgSlug.Status != pgtype.Null {
		eThread.Slug = pgSlug.String
	}

	tgtimeToString(&pgTime, &eThread.Created)

	err = tx.Commit()
	if err != nil {
		check(err)
	}

	return &eThread, nil
}

func (self HandlerDB) ThreadGetPosts(SlugOrID string, Sort *string, Since *int, Desc *bool, Limit *int) (models.Posts, error) {
	tx, err := self.pool.Begin()
	if err != nil {
		check(err)
	}
	defer tx.Rollback()

	check("get_posts")

	queryCheck := "select id from threads where"
	var currentErr string

	if _, err := strconv.Atoi(SlugOrID); err != nil {
		currentErr = fmt.Sprintf("Can't find thread by slug: %s", SlugOrID)
		queryCheck += " slug = $1"
	} else {
		currentErr = fmt.Sprintf("Can't find thread by id: %s", SlugOrID)
		queryCheck += " id = $1"
	}

	var tId int32
	if err := tx.QueryRow(queryCheck, SlugOrID).Scan(&tId); err != nil {
		return nil, errors.New(currentErr)
	}

	check("thread_valid")

	args := []interface{}{}
	//var qValues []string

	args = append(args, tId)

	var query string

	switch *Sort {
	default:
		fallthrough
	case "flat":
		{
			query = "select id, parent, message, isEdit, forum, created, thread, author from posts where thread = $1"

			if Since != nil {
				args = append(args, *Since)
				if Desc != nil && *Desc {
					query += fmt.Sprintf(" and id < $%d", len(args))
				} else {
					query += fmt.Sprintf(" and id > $%d", len(args))
				}
			}

			if Desc != nil && *Desc {
				query += " order by id desc"
			} else {
				query += " order by id"
			}

			if Limit != nil {
				args = append(args, *Limit)
				query += fmt.Sprintf(" limit $%d", len(args))
			}
		}

	case "tree":
		{
			query = `select id, parent, message, isEdit, forum, created, thread, author from posts where thread = $1`

			if Since != nil {
				args = append(args, *Since)
				if Desc != nil && *Desc {
					query += fmt.Sprintf(" and mPath < (select mPath from posts where id = $%d) ", len(args))
				} else {
					query += fmt.Sprintf(" and mPath > (select mPath from posts where id = $%d) ", len(args))
				}
			}

			if Desc != nil && *Desc {
				query += " order by mPath desc"
			} else {
				query += " order by mPath"
			}

			if Limit != nil {
				args = append(args, *Limit)
				query += fmt.Sprintf(" limit $%d", len(args))
			}
		}

	case "parent_tree":
		{
			query = `select id, parent, message, isEdit, forum, created, thread, author
			from posts where rootParent in
			(select id from posts where thread = $1 and parent = 0`

			if Since != nil {
				args = append(args, *Since)
				if Desc != nil && *Desc {
					query += fmt.Sprintf(" and id < (select rootParent from posts where id = $%d) ", len(args))
				} else {
					query += fmt.Sprintf(" and id > (select rootParent from posts where id = $%d) ", len(args))
				}
			}

			if Desc != nil && *Desc {
				query += " order by id desc"
			} else {
				query += " order by id"
			}

			if Limit != nil {
				args = append(args, *Limit)
				query += fmt.Sprintf(" limit $%d", len(args))
			}

			query += `)`

			if Desc != nil && *Desc {
				query += " order by rootParent desc, mPath"
			} else {
				query += " order by mPath"
			}
		}
	}

	check("- begin -----")
	check(query)
	check("- end -------")

	rows, err := tx.Query(query, args...)
	if err != nil {
		check(err)
	}

	//log.Printf("%#v\n", rows)

	fetchPosts := models.Posts{}
	pgTime := pgtype.Timestamptz{}
	pgSlug := pgtype.Text{}
	for rows.Next() {
		check("fit\n")
		post := models.Post{}
		err := rows.Scan(&post.ID, &post.Parent, &post.Message, &post.IsEdited, &pgSlug, &pgTime, &post.Thread, &post.Author)
		if err != nil {
			check(err)
		}

		if pgSlug.Status != pgtype.Null {
			post.Forum = pgSlug.String
		}

		tgtimeToString(&pgTime, &post.Created)

		fetchPosts = append(fetchPosts, &post)
	}

	rows.Close()

	check("precommit")
	if err = tx.Commit(); err != nil {
		check(err)
	}

	return fetchPosts, nil
}

func (self HandlerDB) ThreadUpdate(SlugOrID string, Thread *models.ThreadUpdate) (*models.Thread, error) {
	tx, err := self.pool.Begin()
	if err != nil {
		check(err)
	}
	defer tx.Rollback()

	check("thread_update")

	query := "select id from threads where"
	var currentErr string

	if _, err := strconv.Atoi(SlugOrID); err != nil {
		currentErr = fmt.Sprintf("Can't find thread by slug: %s", SlugOrID)
		query += " slug = $1"
	} else {
		currentErr = fmt.Sprintf("Can't find thread by id: %s", SlugOrID)
		query += " id = $1"
	}

	var tId int32
	if err := tx.QueryRow(query, SlugOrID).Scan(&tId); err != nil {
		return nil, errors.New(currentErr)
	}

	if Thread.Message == "" && Thread.Title == "" {
		return self.ThreadGetOne(SlugOrID)
	}

	args := []interface{}{}
	var qValues []string

	if Thread.Message != "" {
		args = append(args, Thread.Message)
		qValues = append(qValues, fmt.Sprintf("message = $%d", len(args)))
	}

	if Thread.Title != "" {
		args = append(args, Thread.Title)
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

	tgtimeToString(&pgTime, &updThread.Created)

	if err = tx.Commit(); err != nil {
		check(err)
	}

	return &updThread, nil
}

func (self HandlerDB) ThreadVote(SlugOrID string, Vote *models.Vote) (*models.Thread, error) {
	tx, err := self.pool.Begin()
	if err != nil {
		check(err)
	}
	defer tx.Rollback()

	check("vote")

	query := "select id from threads where"
	var currentErr string
	if _, err := strconv.Atoi(SlugOrID); err != nil {
		currentErr = fmt.Sprintf("Can't find thread by slug: %s", SlugOrID)
		query += " slug = $1"
	} else {
		currentErr = fmt.Sprintf("Can't find thread by id: %s", SlugOrID)
		query += " id = $1::bigint"
	}

	var tId int32
	if err := tx.QueryRow(query, SlugOrID).Scan(&tId); err != nil {
		return nil, errors.New(currentErr)
	}

	if _, err := tx.Exec("insert into votes (author, thread, vote) values ($1, $2, $3) on conflict (author, thread) do update set vote = $3", Vote.Nickname, tId, Vote.Voice); err != nil {
		check(err)
		currentErr = fmt.Sprintf("Can't find user by nickname: %s", Vote.Nickname)
		return nil, errors.New(currentErr)
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

	tgtimeToString(&pgTime, &updThread.Created)

	if err = tx.Commit(); err != nil {
		check(err)
	}

	return &updThread, nil
}

package mapper

import (
	"errors"
	"fmt"
	"github.com/Silvman/tech-db-forum/models"
	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"strconv"
	"strings"
)

const qSelectThreadsForumTitle = `select id, title, message, votes, slug, created, forum, author from threads where ((title = $1) and (forum = $2) and (message = $3))`
const qSelectThreadsForumSlug = `select id, title, message, votes, slug, created, forum, author from threads where (slug = $1)`

const qInsertThread = `insert into threads (title, message, author, forum) values ($1, $2, $3, $4) returning id, title, message, votes, slug, created, forum, author`
const qInsertThreadCreated = `insert into threads (title, message, author, forum, created) values ($1, $2, $3, $4, $5) returning id, title, message, votes, slug, created, forum, author`
const qInsertThreadCreatedSlug = `insert into threads (title, message, author, forum, created, slug) values ($1, $2, $3, $4, $5, $6) returning id, title, message, votes, slug, created, forum, author`
const qInsertThreadSlug = `insert into threads (title, message, author, forum, slug) values ($1, $2, $3, $4, $5) returning id, title, message, votes, slug, created, forum, author`

const qSelectThreadBySlug = `select id, title, message, votes, slug, created, forum, author from threads where slug = $1`

const qSelectIdFromThreadsId = `select id from threads where id = $1::bigint`
const qSelectIdFromThreadsSlug = `select id from threads where slug = $1`

const qSelectPostsPTDesc = `select id, parent, message, isEdit, forum, created, thread, author from posts where rootParent in (select id from posts where thread = $1 and parent = 0 order by id desc limit $2) order by rootParent desc, mPath`
const qSelectPostsPT = `select id, parent, message, isEdit, forum, created, thread, author from posts where rootParent in (select id from posts where thread = $1 and parent = 0 order by id limit $2) order by mPath`
const qSelectPostsPTSinceDesc = `select id, parent, message, isEdit, forum, created, thread, author from posts where rootParent in (select id from posts where thread = $1 and parent = 0 and id < (select rootParent from posts where id = $2)  order by id desc limit $3) order by rootParent desc, mPath`
const qSelectPostsPTSince = `select id, parent, message, isEdit, forum, created, thread, author from posts where rootParent in (select id from posts where thread = $1 and parent = 0 and id > (select rootParent from posts where id = $2)  order by id limit $3) order by mPath`

const qSelectPostsTDesc = `select id, parent, message, isEdit, forum, created, thread, author from posts where thread = $1 order by mPath desc limit $2`
const qSelectPostsT = `select id, parent, message, isEdit, forum, created, thread, author from posts where thread = $1 order by mPath limit $2`
const qSelectPostsTSinceDesc = `select id, parent, message, isEdit, forum, created, thread, author from posts where thread = $1 and mPath < (select mPath from posts where id = $2)  order by mPath desc limit $3`
const qSelectPostsTSince = `select id, parent, message, isEdit, forum, created, thread, author from posts where thread = $1 and mPath > (select mPath from posts where id = $2)  order by mPath limit $3`

const qSelectPostsFDesc = `select id, parent, message, isEdit, forum, created, thread, author from posts where thread = $1 order by id desc limit $2`
const qSelectPostsF = `select id, parent, message, isEdit, forum, created, thread, author from posts where thread = $1 order by id limit $2`
const qSelectPostsFSinceDesc = `select id, parent, message, isEdit, forum, created, thread, author from posts where thread = $1 and id < $2 order by id desc limit $3`
const qSelectPostsFSince = `select id, parent, message, isEdit, forum, created, thread, author from posts where thread = $1 and id > $2 order by id limit $3`

func (self HandlerDB) ThreadCreate(Slug string, Thread *models.Thread) (*models.Thread, error) {
	tx, err := self.pool.Begin()
	if err != nil {
		check(err)
	}
	defer tx.Rollback()

	var slug string
	if err := tx.QueryRow(qSelectSlug, Slug).Scan(&slug); err != nil {
		return nil, errors.New(fmt.Sprintf("Can't find thread forum by slug: %s", Slug))
	}

	var author string
	if err := tx.QueryRow(qSelectUsersNickname, Thread.Author).Scan(&author); err != nil {
		return nil, errors.New(fmt.Sprintf("Can't find thread author by nickname: %s", Thread.Author))
	}

	pgSlug := pgtype.Text{}
	eThread := models.Thread{}

	if Thread.Slug == "" {
		err = tx.QueryRow(qSelectThreadsForumTitle, Thread.Title, Slug, Thread.Message).Scan(&eThread.ID, &eThread.Title, &eThread.Message, &eThread.Votes, &pgSlug, &eThread.Created, &eThread.Forum, &eThread.Author)
	} else {
		err = tx.QueryRow(qSelectThreadsForumSlug, Thread.Slug).Scan(&eThread.ID, &eThread.Title, &eThread.Message, &eThread.Votes, &pgSlug, &eThread.Created, &eThread.Forum, &eThread.Author)
	}

	if err == nil {
		if pgSlug.Status != pgtype.Null {
			eThread.Slug = pgSlug.String
		}

		return &eThread, errors.New("already exists")
	} else {
		check(err)
	}

	newThread := models.Thread{}
	if Thread.Created.IsZero() {
		if Thread.Slug != "" {
			err = tx.QueryRow(qInsertThreadSlug, Thread.Title, Thread.Message, Thread.Author, slug, Thread.Slug).Scan(&newThread.ID, &newThread.Title, &newThread.Message, &newThread.Votes, &pgSlug, &newThread.Created, &newThread.Forum, &newThread.Author)
		} else {
			err = tx.QueryRow(qInsertThread, Thread.Title, Thread.Message, Thread.Author, slug).Scan(&newThread.ID, &newThread.Title, &newThread.Message, &newThread.Votes, &pgSlug, &newThread.Created, &newThread.Forum, &newThread.Author)
		}
	} else {
		if Thread.Slug != "" {
			err = tx.QueryRow(qInsertThreadCreatedSlug, Thread.Title, Thread.Message, Thread.Author, slug, Thread.Created, Thread.Slug).Scan(&newThread.ID, &newThread.Title, &newThread.Message, &newThread.Votes, &pgSlug, &newThread.Created, &newThread.Forum, &newThread.Author)
		} else {
			err = tx.QueryRow(qInsertThreadCreated, Thread.Title, Thread.Message, Thread.Author, slug, Thread.Created).Scan(&newThread.ID, &newThread.Title, &newThread.Message, &newThread.Votes, &pgSlug, &newThread.Created, &newThread.Forum, &newThread.Author)
		}
	}

	if err != nil {
		check(err)
	}

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
	tx, err := self.pool.Begin()
	if err != nil {
		check(err)
	}
	defer tx.Rollback()

	//pgTime := pgtype.Timestamptz{}
	pgSlug := pgtype.Text{}
	eThread := models.Thread{}

	if _, err := strconv.Atoi(SlugOrID); err != nil {
		if err := tx.QueryRow(qSelectThreadBySlug, SlugOrID).Scan(&eThread.ID, &eThread.Title, &eThread.Message, &eThread.Votes, &pgSlug, &eThread.Created, &eThread.Forum, &eThread.Author); err != nil {
			return nil, errors.New(fmt.Sprintf("Can't find thread by slug: %s", SlugOrID))
		}
	} else {
		if err := tx.QueryRow(qSelectThreadById, SlugOrID).Scan(&eThread.ID, &eThread.Title, &eThread.Message, &eThread.Votes, &pgSlug, &eThread.Created, &eThread.Forum, &eThread.Author); err != nil {
			return nil, errors.New(fmt.Sprintf("Can't find thread by id: %s", SlugOrID))
		}
	}

	if pgSlug.Status != pgtype.Null {
		eThread.Slug = pgSlug.String
	}

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

	var tId int32
	if _, err := strconv.Atoi(SlugOrID); err != nil {
		if err := tx.QueryRow(qSelectIdFromThreadsSlug, SlugOrID).Scan(&tId); err != nil {
			return nil, errors.New(fmt.Sprintf("Can't find thread by slug: %s", SlugOrID))
		}
	} else {
		if err := tx.QueryRow(qSelectIdFromThreadsId, SlugOrID).Scan(&tId); err != nil {
			return nil, errors.New(fmt.Sprintf("Can't find thread by id: %s", SlugOrID))
		}
	}

	var rows *pgx.Rows
	switch *Sort {
	default:
		fallthrough

	case "flat":
		{
			if Desc != nil && *Desc {
				if Since != nil {
					rows, err = tx.Query(qSelectPostsFSinceDesc, tId, *Since, *Limit)
				} else {
					rows, err = tx.Query(qSelectPostsFDesc, tId, *Limit)
				}
			} else {
				if Since != nil {
					rows, err = tx.Query(qSelectPostsFSince, tId, *Since, *Limit)
				} else {
					rows, err = tx.Query(qSelectPostsF, tId, *Limit)
				}
			}
		}

	case "tree":
		{
			if Desc != nil && *Desc {
				if Since != nil {
					rows, err = tx.Query(qSelectPostsTSinceDesc, tId, *Since, *Limit)
				} else {
					rows, err = tx.Query(qSelectPostsTDesc, tId, *Limit)
				}
			} else {
				if Since != nil {
					rows, err = tx.Query(qSelectPostsTSince, tId, *Since, *Limit)
				} else {
					rows, err = tx.Query(qSelectPostsT, tId, *Limit)
				}
			}
		}

	case "parent_tree":
		{
			if Desc != nil && *Desc {
				if Since != nil {
					rows, err = tx.Query(qSelectPostsPTSinceDesc, tId, *Since, *Limit)
				} else {
					rows, err = tx.Query(qSelectPostsPTDesc, tId, *Limit)
				}
			} else {
				if Since != nil {
					rows, err = tx.Query(qSelectPostsPTSince, tId, *Since, *Limit)
				} else {
					rows, err = tx.Query(qSelectPostsPT, tId, *Limit)
				}
			}
		}
	}

	if err != nil {
		check(err)
	}

	fetchPosts := models.Posts{}
	pgSlug := pgtype.Text{}
	for rows.Next() {
		post := models.Post{}
		err := rows.Scan(&post.ID, &post.Parent, &post.Message, &post.IsEdited, &pgSlug, &post.Created, &post.Thread, &post.Author)
		if err != nil {
			check(err)
		}

		if pgSlug.Status != pgtype.Null {
			post.Forum = pgSlug.String
		}

		fetchPosts = append(fetchPosts, &post)
	}

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

	var tId int32
	if _, err := strconv.Atoi(SlugOrID); err != nil {
		if err := tx.QueryRow(qSelectIdFromThreadsSlug, SlugOrID).Scan(&tId); err != nil {
			return nil, errors.New(fmt.Sprintf("Can't find thread by slug: %s", SlugOrID))
		}
	} else {
		if err := tx.QueryRow(qSelectIdFromThreadsId, SlugOrID).Scan(&tId); err != nil {
			return nil, errors.New(fmt.Sprintf("Can't find thread by id: %s", SlugOrID))
		}
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

	pgSlug := pgtype.Text{}
	updThread := models.Thread{}
	query := fmt.Sprintf("update threads set "+strings.Join(qValues, ",")+" where id = %d returning id, title, message, votes, slug, created, forum, author", tId)
	if err := tx.QueryRow(query, args...).Scan(
		&updThread.ID,
		&updThread.Title,
		&updThread.Message,
		&updThread.Votes,
		&pgSlug,
		&updThread.Created,
		&updThread.Forum,
		&updThread.Author,
	); err != nil {
		check(err)
	}

	if pgSlug.Status != pgtype.Null {
		updThread.Slug = pgSlug.String
	}

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

	var tId int32
	if _, err := strconv.Atoi(SlugOrID); err != nil {
		if err := tx.QueryRow(qSelectIdFromThreadsSlug, SlugOrID).Scan(&tId); err != nil {
			return nil, errors.New(fmt.Sprintf("Can't find thread by slug: %s", SlugOrID))
		}
	} else {
		if err := tx.QueryRow(qSelectIdFromThreadsId, SlugOrID).Scan(&tId); err != nil {
			return nil, errors.New(fmt.Sprintf("Can't find thread by id: %s", SlugOrID))
		}
	}

	if _, err := tx.Exec("insert into votes (author, thread, vote) values ($1, $2, $3) on conflict (author, thread) do update set vote = $3", Vote.Nickname, tId, Vote.Voice); err != nil {
		return nil, errors.New(fmt.Sprintf("Can't find user by nickname: %s", Vote.Nickname))
	}

	pgSlug := pgtype.Text{}
	updThread := models.Thread{}
	if err := tx.QueryRow(`select id, title, message, votes, slug, created, forum, author from threads where id = $1`, tId).
		Scan(
			&updThread.ID,
			&updThread.Title,
			&updThread.Message,
			&updThread.Votes,
			&pgSlug,
			&updThread.Created,
			&updThread.Forum,
			&updThread.Author,
		); err != nil {
		check(err)
	}

	if pgSlug.Status != pgtype.Null {
		updThread.Slug = pgSlug.String
	}

	if err = tx.Commit(); err != nil {
		check(err)
	}

	return &updThread, nil
}

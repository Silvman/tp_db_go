package service

import (
	"fmt"
	"github.com/Silvman/tech-db-forum/models"
	"github.com/Silvman/tech-db-forum/restapi/operations"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/jackc/pgx/pgtype"
)

const qSelectForumBySlug = `select slug, title, posts, threads, owner from forums where slug=$1`
const qInsertForum = `insert into forums (slug, title, owner) values ($1, $2, $3) returning owner`

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
		//log.Fatalln(err)
	}
	defer tx.Rollback()

	//log.Println("forum_create")

	forumExisting := models.Forum{}

	if err = tx.QueryRow(qSelectForumBySlug, params.Forum.Slug).
		Scan(
			&forumExisting.Slug,
			&forumExisting.Title,
			&forumExisting.Posts,
			&forumExisting.Threads,
			&forumExisting.User,
		); err == nil {
		return operations.NewForumCreateConflict().WithPayload(&forumExisting)
	}

	var nickname string
	if err := tx.QueryRow(`select nickname from users where nickname = $1`, params.Forum.User).Scan(&nickname); err != nil {
		currError := models.Error{Message: fmt.Sprintf("Can't find user with nickname: %s", params.Forum.User)}
		return operations.NewForumCreateNotFound().WithPayload(&currError)
	}

	if err := tx.QueryRow(qInsertForum, params.Forum.Slug, params.Forum.Title, nickname).
		Scan(&params.Forum.User); err != nil {
		//log.Println(err)
	}

	tx.Commit()
	return operations.NewForumCreateCreated().WithPayload(params.Forum)
}

func (self HandlerDB) ForumGetOne(params operations.ForumGetOneParams) middleware.Responder {
	tx, err := self.pool.Begin()
	if err != nil {
		//log.Fatalln(err)
	}
	defer tx.Rollback()

	forumExisting := models.Forum{}

	if err = tx.QueryRow("select slug, title, posts, threads, owner from forums where slug = $1", params.Slug).
		Scan(
			&forumExisting.Slug,
			&forumExisting.Title,
			&forumExisting.Posts,
			&forumExisting.Threads,
			&forumExisting.User,
		); err != nil {
		currError := models.Error{Message: fmt.Sprintf("Can't find forum with slug: %s", params.Slug)}
		return operations.NewForumGetOneNotFound().WithPayload(&currError)
	}

	tx.Commit()
	return operations.NewForumGetOneOK().WithPayload(&forumExisting)
}

func (self HandlerDB) ForumGetThreads(params operations.ForumGetThreadsParams) middleware.Responder {
	tx, err := self.pool.Begin()
	if err != nil {
		//log.Fatalln(err)
	}
	defer tx.Rollback()

	var eSlug string
	if err = tx.QueryRow("select slug from forums where slug = $1", params.Slug).Scan(&eSlug); err != nil {
		currError := models.Error{Message: fmt.Sprintf("Can't find forum with slug: %s", params.Slug)}
		return operations.NewForumGetOneNotFound().WithPayload(&currError)
	}

	args := []interface{}{}
	query := `select id, title, message, votes, slug, created, forum, author from threads where forum = $1`
	args = append(args, params.Slug)

	if params.Since != nil {
		args = append(args, params.Since.String())
		if params.Desc != nil && *params.Desc {
			query += fmt.Sprintf(" and created <= $%d::timestamptz", len(args))
		} else {
			query += fmt.Sprintf(" and created >= $%d::timestamptz", len(args))
		}
	}

	query += " order by created"

	if params.Desc != nil && *params.Desc {
		query += " desc"
	}

	if params.Limit != nil {
		args = append(args, *params.Limit)
		query += fmt.Sprintf(" limit $%d", len(args))
	}

	rows, err := tx.Query(query, args...)

	existingThreads := models.Threads{}
	pgTime := pgtype.Timestamptz{}
	pgSlug := pgtype.Text{}
	for rows.Next() {
		thread := models.Thread{}
		err := rows.Scan(&thread.ID, &thread.Title, &thread.Message, &thread.Votes, &pgSlug, &pgTime, &thread.Forum, &thread.Author)
		if err != nil {
			//log.Println(err)
		}

		if pgSlug.Status != pgtype.Null {
			thread.Slug = pgSlug.String
		}

		t := strfmt.NewDateTime()
		t.Scan(pgTime.Time)

		thread.Created = &t

		existingThreads = append(existingThreads, &thread)
	}
	tx.Commit()

	return operations.NewForumGetThreadsOK().WithPayload(existingThreads)
}

func (self HandlerDB) ForumGetUsers(params operations.ForumGetUsersParams) middleware.Responder {
	tx, err := self.pool.Begin()
	if err != nil {
		//log.Fatalln(err)
	}
	defer tx.Rollback()

	var eSlug string
	if err = tx.QueryRow("select slug from forums where slug = $1", params.Slug).Scan(&eSlug); err != nil {
		currError := models.Error{Message: fmt.Sprintf("Can't find forum by slug: %s", params.Slug)}
		return operations.NewForumGetOneNotFound().WithPayload(&currError)
	}

	args := []interface{}{}
	args = append(args, params.Slug)

	qOuter := `select nickname, fullname, about, email from users where nickname in (`
	qSelectUserThreads := `select distinct (author) from threads where forum = $1`
	qSelectUserPosts := `select distinct (p.author) from posts p`
	qJoin := `join threads t on p.thread = t.id where t.forum = $1`
	qOuterClose := `) order by nickname`

	if params.Since != nil {
		if params.Desc != nil && *params.Desc {
			args = append(args, *params.Since)
			qSelectUserThreads += fmt.Sprintf(" and author < $%d", len(args))
			qJoin += fmt.Sprintf(" and p.author < $%d", len(args))
		} else {
			args = append(args, *params.Since)
			qSelectUserThreads += fmt.Sprintf(" and author > $%d", len(args))
			qJoin += fmt.Sprintf(" and p.author > $%d", len(args))
		}
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

	//log.Println(query)

	rows, err := tx.Query(query, args...)

	//log.Println(err)

	existingUsers := models.Users{}

	for rows.Next() {
		t := models.User{}
		rows.Scan(&t.Nickname, &t.Fullname, &t.About, &t.Email)
		existingUsers = append(existingUsers, &t)
	}
	tx.Commit()

	return operations.NewForumGetUsersOK().WithPayload(existingUsers)

}

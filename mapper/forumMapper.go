package mapper

import (
	"errors"
	"fmt"
	"github.com/Silvman/tech-db-forum/models"
	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
)

const qSelectForumBySlug = `select slug, title, posts, threads, owner from forums where slug = $1`
const qSelectUsersNickname = `select nickname from users where nickname = $1`
const qInsertForum = `insert into forums (slug, title, owner) values ($1, $2, $3) returning owner`
const qSelectSlug = `select slug from forums where slug = $1`

const qSelectUsersSinceDesc = `select u.nickname, fullname, about, email from forums_users join users u on forums_users.uid = u.id where forum = $1 and u.nickname < $2 order by u.nickname desc limit $3`
const qSelectUsersDesc = `select u.nickname, fullname, about, email from forums_users join users u on forums_users.uid = u.id where forum = $1 order by u.nickname desc limit $2`
const qSelectUsersSince = `select u.nickname, fullname, about, email from forums_users join users u on forums_users.uid = u.id where forum = $1 and u.nickname > $2 order by u.nickname limit $3`
const qSelectUsers = `select u.nickname, fullname, about, email from forums_users join users u on forums_users.uid = u.id where forum = $1 order by u.nickname limit $2`

const qSelectThreadsCreatedDesc = `select id, title, message, votes, slug, created, forum, author from threads where forum = $1 and created <= $2::timestamptz order by created desc limit $3`
const qSelectThreadsCreated = `select id, title, message, votes, slug, created, forum, author from threads where forum = $1 and created >= $2::timestamptz order by created limit $3`
const qSelectThreadsDesc = `select id, title, message, votes, slug, created, forum, author from threads where forum = $1 order by created desc limit $2`
const qSelectThreads = `select id, title, message, votes, slug, created, forum, author from threads where forum = $1 order by created limit $2`

func (self *HandlerDB) ForumCreate(Forum *models.Forum) (*models.Forum, error) {
	tx, err := self.pool.Begin()
	if err != nil {
		check(err)
	}
	defer tx.Rollback()

	check("forum_create")
	forumExisting := models.Forum{}
	if err = tx.QueryRow(qSelectForumBySlug, Forum.Slug).
		Scan(
			&forumExisting.Slug,
			&forumExisting.Title,
			&forumExisting.Posts,
			&forumExisting.Threads,
			&forumExisting.User,
		); err == nil {
		return &forumExisting, errors.New("already exists")
	}

	var nickname string
	if err := tx.QueryRow(qSelectUsersNickname, Forum.User).Scan(&nickname); err != nil {
		return nil, errors.New(fmt.Sprintf("Can't find user with nickname: %s", Forum.User))
	}

	if err := tx.QueryRow(qInsertForum, Forum.Slug, Forum.Title, nickname).
		Scan(&Forum.User); err != nil {
		check(err)
	}

	err = tx.Commit()
	if err != nil {
		check(err)
	}
	return Forum, nil
}

func (self *HandlerDB) ForumGetOne(Slug string) (*models.Forum, error) {
	tx, err := self.pool.Begin()
	if err != nil {
		check(err)
	}
	defer tx.Rollback()

	forumExisting := models.Forum{}
	if err = tx.QueryRow(qSelectForumBySlug, Slug).
		Scan(
			&forumExisting.Slug,
			&forumExisting.Title,
			&forumExisting.Posts,
			&forumExisting.Threads,
			&forumExisting.User,
		); err != nil {
		return nil, errors.New(fmt.Sprintf("Can't find forum with slug: %s", Slug))
	}

	err = tx.Commit()
	if err != nil {
		check(err)
	}
	return &forumExisting, nil
}

func (self HandlerDB) ForumGetThreads(Slug string, Desc *bool, Since *string, Limit *int) (models.Threads, error) {
	tx, err := self.pool.Begin()
	if err != nil {
		check(err)
	}
	defer tx.Rollback()

	var eSlug string
	if err = tx.QueryRow(qSelectSlug, Slug).Scan(&eSlug); err != nil {
		return nil, errors.New(fmt.Sprintf("Can't find forum with slug: %s", Slug))
	}

	var rows *pgx.Rows
	if Desc != nil && *Desc {
		if Since != nil {
			rows, _ = tx.Query(qSelectThreadsCreatedDesc, eSlug, *Since, *Limit)
		} else {
			rows, _ = tx.Query(qSelectThreadsDesc, eSlug, *Limit)
		}
	} else {
		if Since != nil {
			rows, _ = tx.Query(qSelectThreadsCreated, eSlug, *Since, *Limit)
		} else {
			rows, _ = tx.Query(qSelectThreads, eSlug, *Limit)
		}
	}

	existingThreads := make(models.Threads, 0, 20)
	pgSlug := pgtype.Text{}
	for rows.Next() {
		thread := models.Thread{}
		if err := rows.Scan(&thread.ID, &thread.Title, &thread.Message, &thread.Votes, &pgSlug, &thread.Created, &thread.Forum, &thread.Author); err != nil {
			check(err)
		}

		if pgSlug.Status != pgtype.Null {
			thread.Slug = pgSlug.String
		}

		existingThreads = append(existingThreads, &thread)
	}

	if err = tx.Commit(); err != nil {
		check(err)
	}

	return existingThreads, nil
}

func (self HandlerDB) ForumGetUsers(Slug string, Desc *bool, Since *string, Limit *int) (models.Users, error) {
	tx, err := self.pool.Begin()
	if err != nil {
		check(err)
	}
	defer tx.Rollback()

	var eSlug string
	if err = tx.QueryRow(qSelectSlug, Slug).Scan(&eSlug); err != nil {
		return nil, errors.New(fmt.Sprintf("Can't find forum by slug: %s", Slug))
	}

	var rows *pgx.Rows
	if Desc != nil && *Desc {
		if Since != nil {
			rows, _ = tx.Query(qSelectUsersSinceDesc, eSlug, *Since, *Limit)
		} else {
			rows, _ = tx.Query(qSelectUsersDesc, eSlug, *Limit)
		}
	} else {
		if Since != nil {
			rows, _ = tx.Query(qSelectUsersSince, eSlug, *Since, *Limit)
		} else {
			rows, _ = tx.Query(qSelectUsers, eSlug, *Limit)
		}
	}

	existingUsers := models.Users{}
	for rows.Next() {
		t := models.User{}
		rows.Scan(&t.Nickname, &t.Fullname, &t.About, &t.Email)
		existingUsers = append(existingUsers, &t)
	}

	if err = tx.Commit(); err != nil {
		check(err)
	}

	return existingUsers, nil
}

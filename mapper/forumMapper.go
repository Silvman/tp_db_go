package mapper

import (
	"errors"
	"fmt"
	"github.com/Silvman/tech-db-forum/models"
	"github.com/jackc/pgx/pgtype"
)

const qSelectForumBySlug = `select slug, title, posts, threads, owner from forums where slug=$1`
const qInsertForum = `insert into forums (slug, title, owner) values ($1, $2, $3) returning owner`

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

	if err := tx.QueryRow(qInsertForum, Forum.Slug, Forum.Title, Forum.User).
		Scan(&Forum.User); err != nil {
		return nil, errors.New(fmt.Sprintf("Can't find user with nickname: %s", Forum.User))
	}

	err = tx.Commit()
	if err != nil {
		check(err)
	}
	return nil, nil
}

func (self *HandlerDB) ForumGetOne(Slug string) (*models.Forum, error) {
	tx, err := self.pool.Begin()
	if err != nil {
		check(err)
	}
	defer tx.Rollback()

	forumExisting := models.Forum{}
	if err = tx.QueryRow("select slug, title, posts, threads, owner from forums where slug = $1", Slug).
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
	if err = tx.QueryRow("select slug from forums where slug = $1", Slug).Scan(&eSlug); err != nil {
		return nil, errors.New(fmt.Sprintf("Can't find forum with slug: %s", Slug))
	}

	args := []interface{}{}
	query := `select id, title, message, votes, slug, created, forum, author from threads where forum = $1`
	args = append(args, Slug)

	if Since != nil {
		args = append(args, *Since)
		if Desc != nil && *Desc {
			query += fmt.Sprintf(" and created <= $%d::timestamptz", len(args))
		} else {
			query += fmt.Sprintf(" and created >= $%d::timestamptz", len(args))
		}
	}

	query += " order by created"

	if Desc != nil && *Desc {
		query += " desc"
	}

	if Limit != nil {
		args = append(args, *Limit)
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
			check(err)
		}

		if pgSlug.Status != pgtype.Null {
			thread.Slug = pgSlug.String
		}

		tgtimeToString(&pgTime, &thread.Created)

		existingThreads = append(existingThreads, &thread)
	}
	tx.Commit()
	if err != nil {
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
	if err = tx.QueryRow("select slug from forums where slug = $1", Slug).Scan(&eSlug); err != nil {
		return nil, errors.New(fmt.Sprintf("Can't find forum by slug: %s", Slug))
	}

	args := []interface{}{}
	args = append(args, Slug)

	query := `select u.nickname, fullname, about, email from forums_users
join users u on forums_users.uid = u.id
where forum = $1`

	if Since != nil {
		if Desc != nil && *Desc {
			args = append(args, *Since)
			query += fmt.Sprintf(" and u.nickname < $%d", len(args))
		} else {
			args = append(args, *Since)
			query += fmt.Sprintf(" and u.nickname > $%d", len(args))
		}
	}

	query += " order by u.nickname"
	if Desc != nil {
		if *Desc {
			query += " desc"
		}
	}

	if Limit != nil {
		args = append(args, *Limit)
		query += fmt.Sprintf(" limit $%d", len(args))
	}
	check(query)

	rows, err := tx.Query(query, args...)
	if err != nil {
		check(err)
	}

	existingUsers := models.Users{}
	for rows.Next() {
		t := models.User{}
		rows.Scan(&t.Nickname, &t.Fullname, &t.About, &t.Email)
		existingUsers = append(existingUsers, &t)
	}

	tx.Commit()
	if err != nil {
		check(err)
	}

	return existingUsers, nil

}

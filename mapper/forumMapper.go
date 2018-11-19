package mapper

import (
	"errors"
	"fmt"
	"github.com/Silvman/tech-db-forum/models"
	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
)

func (self *HandlerDB) ForumCreate(Forum *models.Forum) (*models.Forum, error) {
	forumExisting := models.Forum{}
	if err := self.pool.QueryRow(qSelectForumBySlug, Forum.Slug).
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
	if err := self.pool.QueryRow(qSelectUsersNickname, Forum.User).Scan(&nickname); err != nil {
		return nil, errors.New(fmt.Sprintf("Can't find user with nickname: %s", Forum.User))
	}

	if err := self.pool.QueryRow(qInsertForum, Forum.Slug, Forum.Title, nickname).
		Scan(&Forum.User); err != nil {
		//log.Println(err)
	}

	return Forum, nil
}

func (self *HandlerDB) ForumGetOne(Slug string) (*models.Forum, error) {
	forumExisting := models.Forum{}
	if err := self.pool.QueryRow(qSelectForumBySlug, Slug).
		Scan(
			&forumExisting.Slug,
			&forumExisting.Title,
			&forumExisting.Posts,
			&forumExisting.Threads,
			&forumExisting.User,
		); err != nil {
		return nil, errors.New(fmt.Sprintf("Can't find forum with slug: %s", Slug))
	}

	return &forumExisting, nil
}

func (self HandlerDB) ForumGetThreads(Slug string, Desc *bool, Since *string, Limit *int) (models.Threads, error) {
	var eSlug string
	if err := self.pool.QueryRow(qSelectSlug, Slug).Scan(&eSlug); err != nil {
		return nil, errors.New(fmt.Sprintf("Can't find forum with slug: %s", Slug))
	}

	var rows *pgx.Rows
	if Desc != nil && *Desc {
		if Since != nil {
			rows, _ = self.pool.Query(qSelectThreadsCreatedDesc, eSlug, *Since, *Limit)
		} else {
			rows, _ = self.pool.Query(qSelectThreadsDesc, eSlug, *Limit)
		}
	} else {
		if Since != nil {
			rows, _ = self.pool.Query(qSelectThreadsCreated, eSlug, *Since, *Limit)
		} else {
			rows, _ = self.pool.Query(qSelectThreads, eSlug, *Limit)
		}
	}

	existingThreads := make(models.Threads, 0, 20)
	pgSlug := pgtype.Text{}
	for rows.Next() {
		thread := models.Thread{}
		if err := rows.Scan(&thread.ID, &thread.Title, &thread.Message, &thread.Votes, &pgSlug, &thread.Created, &thread.Forum, &thread.Author); err != nil {
			//log.Println(err)
		}

		if pgSlug.Status != pgtype.Null {
			thread.Slug = pgSlug.String
		}

		existingThreads = append(existingThreads, &thread)
	}

	return existingThreads, nil
}

func (self HandlerDB) ForumGetUsers(Slug string, Desc *bool, Since *string, Limit *int) (models.Users, error) {
	var eSlug string
	if err := self.pool.QueryRow(qSelectSlug, Slug).Scan(&eSlug); err != nil {
		return nil, errors.New(fmt.Sprintf("Can't find forum by slug: %s", Slug))
	}

	var rows *pgx.Rows
	if Desc != nil && *Desc {
		if Since != nil {
			rows, _ = self.pool.Query(qSelectUsersSinceDesc, eSlug, *Since, *Limit)
		} else {
			rows, _ = self.pool.Query(qSelectUsersDesc, eSlug, *Limit)
		}
	} else {
		if Since != nil {
			rows, _ = self.pool.Query(qSelectUsersSince, eSlug, *Since, *Limit)
		} else {
			rows, _ = self.pool.Query(qSelectUsers, eSlug, *Limit)
		}
	}

	existingUsers := models.Users{}
	for rows.Next() {
		t := models.User{}
		rows.Scan(&t.Nickname, &t.Fullname, &t.About, &t.Email)
		existingUsers = append(existingUsers, &t)
	}

	return existingUsers, nil
}

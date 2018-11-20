package mapper

import (
	"errors"
	"fmt"
	"github.com/Silvman/tech-db-forum/models"
	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"log"
)

func (self *HandlerDB) ForumCreate(Forum *models.Forum) (*models.Forum, error) {
	if forumExisting, err := self.ForumGetOne(Forum.Slug); err == nil {
		return forumExisting, errors.New("already exists")
	}

	if err := self.pool.QueryRow(self.psqSelectUsersNickname.Name, Forum.User).Scan(&Forum.User); err != nil {
		return nil, errors.New(fmt.Sprintf("Can't find user with nickname: %s", Forum.User))
	}

	if _, err := self.pool.Exec(self.psqInsertForum.Name, Forum.Slug, Forum.Title, Forum.User); err != nil {
		log.Println(err)
	}

	return Forum, nil
}

func (self *HandlerDB) ForumGetOne(Slug string) (*models.Forum, error) {
	forumExisting := models.Forum{}
	if err := self.pool.QueryRow(self.psqSelectForumBySlug.Name, Slug).
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

func (self *HandlerDB) CheckForumNotExists(Slug string) error {
	var temp int
	if err := self.pool.QueryRow(self.psqCheckForum.Name, Slug).Scan(&temp); err != nil {
		return errors.New(fmt.Sprintf("Can't find forum with slug: %s", Slug))
	}

	return nil
}

func (self HandlerDB) ForumGetThreads(Slug string, Desc *bool, Since *string, Limit *int) (*models.Threads, error) {
	var rows *pgx.Rows
	if Desc != nil && *Desc {
		if Since != nil {
			rows, _ = self.pool.Query(self.psqSelectThreadsCreatedDesc.Name, Slug, *Since, *Limit)
		} else {
			rows, _ = self.pool.Query(self.psqSelectThreadsDesc.Name, Slug, *Limit)
		}
	} else {
		if Since != nil {
			rows, _ = self.pool.Query(self.psqSelectThreadsCreated.Name, Slug, *Since, *Limit)
		} else {
			rows, _ = self.pool.Query(self.psqSelectThreads.Name, Slug, *Limit)
		}
	}

	// todo eu
	existingThreads := make(models.Threads, 0, *Limit)
	pgSlug := pgtype.Text{}
	for rows.Next() {
		thread := models.Thread{}
		rows.Scan(&thread.ID, &thread.Title, &thread.Message, &thread.Votes, &pgSlug, &thread.Created, &thread.Forum, &thread.Author)
		if pgSlug.Status != pgtype.Null {
			thread.Slug = pgSlug.String
		}

		existingThreads = append(existingThreads, &thread)
	}

	if len(existingThreads) == 0 {
		if err := self.CheckForumNotExists(Slug); err != nil {
			return nil, err
		}
	}

	return &existingThreads, nil
}

func (self HandlerDB) ForumGetUsers(Slug string, Desc *bool, Since *string, Limit *int) (*models.Users, error) {
	var rows *pgx.Rows
	if Desc != nil && *Desc {
		if Since != nil {
			rows, _ = self.pool.Query(self.psqSelectUsersSinceDesc.Name, Slug, *Since, *Limit)
		} else {
			rows, _ = self.pool.Query(self.psqSelectUsersDesc.Name, Slug, *Limit)
		}
	} else {
		if Since != nil {
			rows, _ = self.pool.Query(self.psqSelectUsersSince.Name, Slug, *Since, *Limit)
		} else {
			rows, _ = self.pool.Query(self.psqSelectUsers.Name, Slug, *Limit)
		}
	}

	existingUsers := make(models.Users, 0, *Limit)
	for rows.Next() {
		t := models.User{}
		rows.Scan(&t.Nickname, &t.Fullname, &t.About, &t.Email)
		existingUsers = append(existingUsers, &t)
	}

	if len(existingUsers) == 0 {
		if err := self.CheckForumNotExists(Slug); err != nil {
			return nil, err
		}
	}

	return &existingUsers, nil
}

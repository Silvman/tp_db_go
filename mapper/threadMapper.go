package mapper

import (
	"errors"
	"fmt"
	"github.com/Silvman/tech-db-forum/models"
	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"log"
	"strconv"
	"strings"
)

func (self HandlerDB) ThreadCreate(Slug string, Thread *models.Thread) (*models.Thread, error) {
	var author string
	if err := self.pool.QueryRow(self.psqSelectUsersNickname.Name, Thread.Author).Scan(&author); err != nil {
		return nil, errors.New(fmt.Sprintf("Can't find thread author by nickname: %s", Thread.Author))
	}

	// ругается на регистр
	if err := self.pool.QueryRow(self.psqSelectSlug.Name, Slug).Scan(&Slug); err != nil {
		return nil, errors.New(fmt.Sprintf("Can't find forum by slug: %s", Slug))
	}

	var err error
	pgSlug := pgtype.Text{}
	eThread := models.Thread{}
	if Thread.Slug == "" {
		err = self.pool.QueryRow(self.psqSelectThreadsForumTitle.Name, Thread.Title, Slug, Thread.Message).Scan(&eThread.ID, &eThread.Title, &eThread.Message, &eThread.Votes, &pgSlug, &eThread.Created, &eThread.Forum, &eThread.Author)
	} else {
		err = self.pool.QueryRow(self.psqSelectThreadsForumSlug.Name, Thread.Slug).Scan(&eThread.ID, &eThread.Title, &eThread.Message, &eThread.Votes, &pgSlug, &eThread.Created, &eThread.Forum, &eThread.Author)
	}

	if err == nil {
		if pgSlug.Status != pgtype.Null {
			eThread.Slug = pgSlug.String
		}

		return &eThread, errors.New("already exists")
	}

	Thread.Forum = Slug
	if Thread.Created.IsZero() {
		if Thread.Slug != "" {
			err = self.pool.QueryRow(self.psqInsertThreadSlug.Name, Thread.Title, Thread.Message, Thread.Author, Slug, Thread.Slug).Scan(&Thread.ID, &Thread.Created)
		} else {
			err = self.pool.QueryRow(self.psqInsertThread.Name, Thread.Title, Thread.Message, Thread.Author, Slug).Scan(&Thread.ID, &Thread.Created)
		}
	} else {
		if Thread.Slug != "" {
			err = self.pool.QueryRow(self.psqInsertThreadCreatedSlug.Name, Thread.Title, Thread.Message, Thread.Author, Slug, Thread.Created, Thread.Slug).Scan(&Thread.ID, &Thread.Created)
		} else {
			err = self.pool.QueryRow(self.psqInsertThreadCreated.Name, Thread.Title, Thread.Message, Thread.Author, Slug, Thread.Created).Scan(&Thread.ID, &Thread.Created)
		}
	}

	if err != nil {
		log.Println(err)
	}

	return Thread, nil
}

func (self HandlerDB) ThreadGetOne(SlugOrID string) (*models.Thread, error) {
	pgSlug := pgtype.Text{}
	eThread := models.Thread{}

	if _, err := strconv.Atoi(SlugOrID); err != nil {
		if err := self.pool.QueryRow(self.psqSelectThreadBySlug.Name, SlugOrID).Scan(&eThread.ID, &eThread.Title, &eThread.Message, &eThread.Votes, &pgSlug, &eThread.Created, &eThread.Forum, &eThread.Author); err != nil {
			return nil, errors.New(fmt.Sprintf("Can't find thread by slug: %s", SlugOrID))
		}
	} else {
		if err := self.pool.QueryRow(self.psqSelectThreadById.Name, SlugOrID).Scan(&eThread.ID, &eThread.Title, &eThread.Message, &eThread.Votes, &pgSlug, &eThread.Created, &eThread.Forum, &eThread.Author); err != nil {
			return nil, errors.New(fmt.Sprintf("Can't find thread by id: %s", SlugOrID))
		}
	}

	if pgSlug.Status != pgtype.Null {
		eThread.Slug = pgSlug.String
	}

	return &eThread, nil
}

func (self HandlerDB) ThreadGetPosts(SlugOrID string, Sort *string, Since *int, Desc *bool, Limit *int) (*models.Posts, error) {
	var tId int32
	if _, err := strconv.Atoi(SlugOrID); err != nil {
		if err := self.pool.QueryRow(self.psqSelectIdFromThreadsSlug.Name, SlugOrID).Scan(&tId); err != nil {
			return nil, errors.New(fmt.Sprintf("Can't find thread by slug: %s", SlugOrID))
		}
	} else {
		if err := self.pool.QueryRow(self.psqSelectIdFromThreadsId.Name, SlugOrID).Scan(&tId); err != nil {
			return nil, errors.New(fmt.Sprintf("Can't find thread by id: %s", SlugOrID))
		}
	}

	var err error
	var rows *pgx.Rows

	switch *Sort {
	default:
		fallthrough

	case "flat":
		{
			if Desc != nil && *Desc {
				if Since != nil {
					rows, err = self.pool.Query(self.psqSelectPostsFSinceDesc.Name, tId, *Since, *Limit)
				} else {
					rows, err = self.pool.Query(self.psqSelectPostsFDesc.Name, tId, *Limit)
				}
			} else {
				if Since != nil {
					rows, err = self.pool.Query(self.psqSelectPostsFSince.Name, tId, *Since, *Limit)
				} else {
					rows, err = self.pool.Query(self.psqSelectPostsF.Name, tId, *Limit)
				}
			}
		}

	case "tree":
		{
			if Desc != nil && *Desc {
				if Since != nil {
					rows, err = self.pool.Query(self.psqSelectPostsTSinceDesc.Name, tId, *Since, *Limit)
				} else {
					rows, err = self.pool.Query(self.psqSelectPostsTDesc.Name, tId, *Limit)
				}
			} else {
				if Since != nil {
					rows, err = self.pool.Query(self.psqSelectPostsTSince.Name, tId, *Since, *Limit)
				} else {
					rows, err = self.pool.Query(self.psqSelectPostsT.Name, tId, *Limit)
				}
			}
		}

	case "parent_tree":
		{
			if Desc != nil && *Desc {
				if Since != nil {
					rows, err = self.pool.Query(self.psqSelectPostsPTSinceDesc.Name, tId, *Since, *Limit)
				} else {
					rows, err = self.pool.Query(self.psqSelectPostsPTDesc.Name, tId, *Limit)
				}
			} else {
				if Since != nil {
					rows, err = self.pool.Query(self.psqSelectPostsPTSince.Name, tId, *Since, *Limit)
				} else {
					rows, err = self.pool.Query(self.psqSelectPostsPT.Name, tId, *Limit)
				}
			}
		}
	}

	if err != nil {
		log.Println(err)
	}

	fetchPosts := make(models.Posts, 0, *Limit)
	pgSlug := pgtype.Text{}

	for rows.Next() {
		post := models.Post{}
		err := rows.Scan(&post.ID, &post.Parent, &post.Message, &post.IsEdited, &pgSlug, &post.Created, &post.Thread, &post.Author)
		if err != nil {
			//log.Println(err)
		}

		if pgSlug.Status != pgtype.Null {
			post.Forum = pgSlug.String
		}

		fetchPosts = append(fetchPosts, &post)
	}

	return &fetchPosts, nil
}

func (self HandlerDB) ThreadUpdate(SlugOrID string, Thread *models.ThreadUpdate) (*models.Thread, error) {
	var tId int32
	if _, err := strconv.Atoi(SlugOrID); err != nil {
		if err := self.pool.QueryRow(self.psqSelectIdFromThreadsSlug.Name, SlugOrID).Scan(&tId); err != nil {
			return nil, errors.New(fmt.Sprintf("Can't find thread by slug: %s", SlugOrID))
		}
	} else {
		if err := self.pool.QueryRow(self.psqSelectIdFromThreadsId.Name, SlugOrID).Scan(&tId); err != nil {
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
	if err := self.pool.QueryRow(query, args...).Scan(
		&updThread.ID,
		&updThread.Title,
		&updThread.Message,
		&updThread.Votes,
		&pgSlug,
		&updThread.Created,
		&updThread.Forum,
		&updThread.Author,
	); err != nil {
		//log.Println(err)
	}

	if pgSlug.Status != pgtype.Null {
		updThread.Slug = pgSlug.String
	}

	return &updThread, nil
}

func (self HandlerDB) ThreadVote(SlugOrID string, Vote *models.Vote) (*models.Thread, error) {
	var tId int32
	if _, err := strconv.Atoi(SlugOrID); err != nil {
		if err := self.pool.QueryRow(self.psqSelectIdFromThreadsSlug.Name, SlugOrID).Scan(&tId); err != nil {
			return nil, errors.New(fmt.Sprintf("Can't find thread by slug: %s", SlugOrID))
		}
	} else {
		if err := self.pool.QueryRow(self.psqSelectIdFromThreadsId.Name, SlugOrID).Scan(&tId); err != nil {
			return nil, errors.New(fmt.Sprintf("Can't find thread by id: %s", SlugOrID))
		}
	}

	if _, err := self.pool.Exec(self.psqInsertVote.Name, Vote.Nickname, tId, Vote.Voice); err != nil {
		return nil, errors.New(fmt.Sprintf("Can't find user by nickname: %s", Vote.Nickname))
	}

	pgSlug := pgtype.Text{}
	updThread := models.Thread{}
	if err := self.pool.QueryRow(self.psqSelectThreadById.Name, tId).
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
		//log.Println(err)
	}

	if pgSlug.Status != pgtype.Null {
		updThread.Slug = pgSlug.String
	}

	return &updThread, nil
}

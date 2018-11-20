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
	"sync/atomic"
)

var totalPosts int32

func (self *HandlerDB) GetPostData(ID int) (*models.Post, error) {
	// todo id не нужен
	post := &models.Post{}
	if err := self.pool.QueryRow(qSelectPostById, ID).
		Scan(
			&post.ID,
			&post.Parent,
			&post.Message,
			&post.IsEdited,
			&post.Forum,
			&post.Created,
			&post.Thread,
			&post.Author,
		); err != nil {
		return nil, errors.New(fmt.Sprintf("Can't find post with id: %s", ID))
	}

	return post, nil
}

func (self *HandlerDB) PostGetOne(ID int, Related []string) (*models.PostFull, error) {
	ePostFull := models.PostFull{}

	var err error
	if ePostFull.Post, err = self.GetPostData(ID); err != nil {
		return nil, err
	}

	if Related != nil {
		for _, value := range Related {
			switch value {
			case "user":
				{
					ePostFull.Author = &models.User{}
					if err := self.pool.QueryRow(qSelectUserByNick, ePostFull.Post.Author).
						Scan(&ePostFull.Author.Nickname, &ePostFull.Author.Fullname, &ePostFull.Author.About, &ePostFull.Author.Email); err != nil {
						//log.Println(err)
					}
				}

			case "forum":
				{
					ePostFull.Forum = &models.Forum{}
					if err := self.pool.QueryRow(qSelectForumBySlug, ePostFull.Post.Forum).
						Scan(
							&ePostFull.Forum.Slug,
							&ePostFull.Forum.Title,
							&ePostFull.Forum.Posts,
							&ePostFull.Forum.Threads,
							&ePostFull.Forum.User,
						); err != nil {
						//log.Println(err)
					}
				}

			case "thread":
				{
					pgSlug := pgtype.Text{}
					ePostFull.Thread = &models.Thread{}
					if err := self.pool.QueryRow(qSelectThreadById, ePostFull.Post.Thread).
						Scan(
							&ePostFull.Thread.ID,
							&ePostFull.Thread.Title,
							&ePostFull.Thread.Message,
							&ePostFull.Thread.Votes,
							&pgSlug,
							&ePostFull.Thread.Created,
							&ePostFull.Thread.Forum,
							&ePostFull.Thread.Author,
						); err != nil {
						//log.Println(err)
					}

					if pgSlug.Status != pgtype.Null {
						ePostFull.Thread.Slug = pgSlug.String
					}
				}
			}
		}
	}

	return &ePostFull, nil
}

func (self HandlerDB) PostUpdate(ID int, Post *models.PostUpdate) (ePost *models.Post, err error) {
	if ePost, err = self.GetPostData(ID); err != nil {
		return nil, err
	}

	if Post.Message != "" && Post.Message != ePost.Message {
		if err := self.pool.QueryRow(qUpdatePost, Post.Message, ID).Scan(
			&ePost.ID,
			&ePost.Parent,
			&ePost.Message,
			&ePost.IsEdited,
			&ePost.Forum,
			&ePost.Created,
			&ePost.Thread,
			&ePost.Author,
		); err != nil {
			log.Println(err)
		}
	}

	return
}

func (self HandlerDB) PostsCreate(SlugOrID string, Posts models.Posts) (models.Posts, error) {
	tx, err := self.pool.Begin()
	if err != nil {
		log.Println(err)
	}
	defer tx.Rollback()

	var tIdCurrent int32
	var tForumCurrent string
	if _, err := strconv.Atoi(SlugOrID); err != nil {
		if err := tx.QueryRow(qSelectIdForumFromThreadsSlug, SlugOrID).Scan(&tIdCurrent, &tForumCurrent); err != nil {
			return nil, errors.New(fmt.Sprintf("Can't find thread by slug: %s", SlugOrID))
		}
	} else {
		if err := tx.QueryRow(qSelectIdForumFromThreadsId, SlugOrID).Scan(&tIdCurrent, &tForumCurrent); err != nil {
			return nil, errors.New(fmt.Sprintf("Can't find thread by id: %s", SlugOrID))
		}
	}

	if len(Posts) == 0 {
		return nil, nil
	}

	query := `insert into posts (parent, message, thread, author, forum) values `
	queryEnd := " returning id, isEdit, created"
	var queryValues []string

	args := make([]interface{}, 0, len(Posts)*5)
	parents := make([]string, 0, len(Posts))

	if len(Posts) == 100 {
		for _, value := range Posts {
			if value.Parent != 0 {
				parents = append(parents, strconv.Itoa(int(value.Parent)))
			}

			args = append(args, value.Parent, value.Message, tIdCurrent, value.Author, tForumCurrent)
		}
	} else {
		for _, value := range Posts {
			if value.Parent != 0 {
				parents = append(parents, strconv.Itoa(int(value.Parent)))
			}

			queryValues = append(queryValues, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d)", len(args)+1, len(args)+2, len(args)+3, len(args)+4, len(args)+5))
			args = append(args, value.Parent, value.Message, tIdCurrent, value.Author, tForumCurrent)
		}
	}

	if len(parents) != 0 {
		rows, err := tx.Query(fmt.Sprint(`select thread from posts where id in (`, strings.Join(parents, ","), ")"))
		hasP := false

		for rows.Next() {
			hasP = true

			var tId int32
			err = rows.Scan(&tId)
			if err != nil {
				//log.Println(err)
				//log.Println(tId)
			}

			if tId != tIdCurrent {
				return nil, errors.New(fmt.Sprintf("Parent post was created in another thread"))
			}
		}

		if !hasP {
			return nil, errors.New(fmt.Sprintf("Parent post was created in another thread"))
		}
	}

	query += strings.Join(queryValues, ",") + queryEnd

	if len(Posts) == 100 {
		query = "bigInsert"
	}

	rows, err := tx.Query(query, args...)

	var par []string
	var nopar []string
	auth := make([]interface{}, 0, len(Posts))
	querries := make([]string, 0, len(Posts))

	auth = append(auth, tForumCurrent)
	for _, value := range Posts {
		if rows.Next() {

			err = rows.Scan(&value.ID, &value.IsEdited, &value.Created)
			if err != nil {
				//log.Println(err)
			}

			if value.Parent != 0 {
				par = append(par, strconv.Itoa(int(value.ID)))
			} else {
				nopar = append(nopar, strconv.Itoa(int(value.ID)))
			}

			auth = append(auth, value.Author)
			querries = append(querries, fmt.Sprintf(`($1, (select id from users where nickname = $%d))`, len(auth)))

			value.Forum = tForumCurrent
			value.Thread = tIdCurrent

		}
	}
	rows.Close()

	if err := rows.Err(); err != nil {
		if err.(pgx.PgError).Code == "23503" {
			return nil, errors.New(fmt.Sprintf("Can't find post author by nickname"))
		}

		log.Println("error on main query")
		log.Println(err)
		log.Println(err.(pgx.PgError).Error())
		return nil, errors.New(fmt.Sprintf("Parent post was created in another thread"))
	}

	if len(par) != 0 {
		tx.Exec(`update posts p set mPath = (select mPath from posts where id = p.parent) || id,
                 rootParent = (select rootParent from posts where id = p.parent)
where id in (` + strings.Join(par, ",") + ")")
	}

	if len(nopar) != 0 {
		tx.Exec("update posts set mPath[1] = id, rootParent = id where id in (" + strings.Join(nopar, ",") + ")")
	}

	tx.Exec(qUpdateForumPosts, len(Posts), tForumCurrent)

	tx.Exec(`insert into forums_users (forum, uid) values `+strings.Join(querries, ",")+` on conflict do nothing`, auth...)

	err = tx.Commit()
	if err != nil {
		log.Println(err)
	}

	atomic.AddInt32(&totalPosts, int32(len(Posts)))

	if atomic.LoadInt32(&totalPosts) >= 1500000 {
		//log.Println("VACUUM ANALYZE")
		self.pool.Exec("cluster")
		self.pool.Exec("VACUUM ANALYZE")
		//log.Println("VACUUM ANALYZE end")
	}

	return Posts, nil
}

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

const qSelectPostById = `select id, parent, message, isEdit, forum, created, thread, author from posts where id = $1`
const qSelectUserByNick = `select nickname, fullname, about, email from users where nickname = $1`
const qSelectThreadById = `select id, title, message, votes, slug, created, forum, author from threads where id = $1`
const qUpdatePost = `update posts set isEdit = true, message = $1 where id = $2 returning id, parent, message, isEdit, forum, created, thread, author`
const qUpdateForumPosts = `update forums set posts = posts + $1 where slug = $2`
const qSelectIdForumFromThreadsId = `select id, forum from threads where id = $1::bigint`
const qSelectIdForumFromThreadsSlug = `select id, forum from threads where slug = $1`

func (self HandlerDB) PostGetOne(ID int, Related []string) (*models.PostFull, error) {
	tx, err := self.pool.Begin()
	if err != nil {
		check(err)
	}
	defer tx.Rollback()

	ePostFull := models.PostFull{}
	ePostFull.Post = &models.Post{}

	// todo а нам нужны все эти поля?
	if err := tx.QueryRow(qSelectPostById, ID).
		Scan(
			&ePostFull.Post.ID,
			&ePostFull.Post.Parent,
			&ePostFull.Post.Message,
			&ePostFull.Post.IsEdited,
			&ePostFull.Post.Forum,
			&ePostFull.Post.Created,
			&ePostFull.Post.Thread,
			&ePostFull.Post.Author,
		); err != nil {
		return nil, errors.New(fmt.Sprintf("Can't find post with id: %s", ID))
	}

	if Related != nil {
		for _, value := range Related {
			switch value {
			case "user":
				{
					ePostFull.Author = &models.User{}
					if err := tx.QueryRow(qSelectUserByNick, ePostFull.Post.Author).
						Scan(&ePostFull.Author.Nickname, &ePostFull.Author.Fullname, &ePostFull.Author.About, &ePostFull.Author.Email); err != nil {
						check(err)
					}
				}

			case "forum":
				{
					ePostFull.Forum = &models.Forum{}
					if err := tx.QueryRow(qSelectForumBySlug, ePostFull.Post.Forum).
						Scan(
							&ePostFull.Forum.Slug,
							&ePostFull.Forum.Title,
							&ePostFull.Forum.Posts,
							&ePostFull.Forum.Threads,
							&ePostFull.Forum.User,
						); err != nil {
						check(err)
					}
				}

			case "thread":
				{
					pgSlug := pgtype.Text{}
					ePostFull.Thread = &models.Thread{}
					if err := tx.QueryRow(qSelectThreadById, ePostFull.Post.Thread).
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
						check(err)
					}

					if pgSlug.Status != pgtype.Null {
						ePostFull.Thread.Slug = pgSlug.String
					}
				}
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		check(err)
	}

	return &ePostFull, nil
}

func (self HandlerDB) PostUpdate(ID int, Post *models.PostUpdate) (*models.Post, error) {
	tx, err := self.pool.Begin()
	if err != nil {
		check(err)
	}
	defer tx.Rollback()

	ePost := models.Post{}
	check("post_update")

	if err := tx.QueryRow(qSelectPostById, ID).Scan(
		&ePost.ID,
		&ePost.Parent,
		&ePost.Message,
		&ePost.IsEdited,
		&ePost.Forum,
		&ePost.Created,
		&ePost.Thread,
		&ePost.Author,
	); err != nil {
		return nil, errors.New(fmt.Sprintf("Can't find post with id: %d", ID))
	}

	if Post.Message != "" && Post.Message != ePost.Message {
		if err := tx.QueryRow(qUpdatePost, Post.Message, ID).Scan(
			&ePost.ID,
			&ePost.Parent,
			&ePost.Message,
			&ePost.IsEdited,
			&ePost.Forum,
			&ePost.Created,
			&ePost.Thread,
			&ePost.Author,
		); err != nil {
			check(err)
		}
	}

	err = tx.Commit()
	if err != nil {
		check(err)
	}

	return &ePost, nil
}

func (self HandlerDB) PostsCreate(SlugOrID string, Posts models.Posts) (models.Posts, error) {
	tx, err := self.pool.Begin()
	if err != nil {
		check(err)
	}
	defer tx.Rollback()

	check("posts_create")

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
		check("nilPosts")
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
		check(err)

		check(fmt.Sprint(`select thread from posts where id in (`, strings.Join(parents, ","), ")"))
		hasP := false

		for rows.Next() {
			hasP = true
			check("select parent thread")

			var tId int32
			err = rows.Scan(&tId)
			check(err)
			check(tId)

			if tId != tIdCurrent {
				return nil, errors.New(fmt.Sprintf("Parent post was created in another thread"))
			}
		}

		if !hasP {
			return nil, errors.New(fmt.Sprintf("Parent post was created in another thread"))
		}
	}

	query += strings.Join(queryValues, ",") + queryEnd

	check(query)
	//log.Printf("%#v\n", args)

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
			check("insert post")

			err = rows.Scan(&value.ID, &value.IsEdited, &value.Created)
			check(err)

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

		check("error on main query")
		check(err)
		check(err.(pgx.PgError).Error())
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
		check(err)
	}

	return Posts, nil
}

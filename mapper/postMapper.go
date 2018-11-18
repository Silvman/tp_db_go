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

func (self HandlerDB) PostGetOne(ID int, Related []string) (*models.PostFull, error) {
	tx, err := self.pool.Begin()
	if err != nil {
		check(err)
	}
	defer tx.Rollback()

	ePostFull := models.PostFull{}
	ePostFull.Post = &models.Post{}

	pgTime := pgtype.Timestamptz{}

	if err := tx.QueryRow(`select id, parent, message, isEdit, forum, created, thread, author from posts where id = $1`, ID).
		Scan(
			&ePostFull.Post.ID,
			&ePostFull.Post.Parent,
			&ePostFull.Post.Message,
			&ePostFull.Post.IsEdited,
			&ePostFull.Post.Forum,
			&pgTime,
			&ePostFull.Post.Thread,
			&ePostFull.Post.Author,
		); err != nil {
		return nil, errors.New(fmt.Sprintf("Can't find post with id: %s", ID))
	}

	tgtimeToString(&pgTime, &ePostFull.Post.Created)

	if Related != nil {
		for _, value := range Related {
			switch value {
			case "user":
				{
					ePostFull.Author = &models.User{}
					if err := tx.QueryRow(`select nickname, fullname, about, email from users where nickname = $1`, ePostFull.Post.Author).
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
					if err := tx.QueryRow(`select id, title, message, votes, slug, created, forum, author from threads where id = $1`, ePostFull.Post.Thread).
						Scan(
							&ePostFull.Thread.ID,
							&ePostFull.Thread.Title,
							&ePostFull.Thread.Message,
							&ePostFull.Thread.Votes,
							&pgSlug,
							&pgTime,
							&ePostFull.Thread.Forum,
							&ePostFull.Thread.Author,
						); err != nil {
						check(err)
					}

					if pgSlug.Status != pgtype.Null {
						ePostFull.Thread.Slug = pgSlug.String
					}

					tgtimeToString(&pgTime, &ePostFull.Thread.Created)
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

	pgTime := pgtype.Timestamptz{}
	ePost := models.Post{}

	check("post_update")

	if err := tx.QueryRow(`select id, parent, message, isEdit, forum, created, thread, author from posts where id = $1`, ID).Scan(
		&ePost.ID,
		&ePost.Parent,
		&ePost.Message,
		&ePost.IsEdited,
		&ePost.Forum,
		&pgTime,
		&ePost.Thread,
		&ePost.Author,
	); err != nil {
		return nil, errors.New(fmt.Sprintf("Can't find post with id: %d", ID))
	}

	tgtimeToString(&pgTime, &ePost.Created)

	if Post.Message != "" && Post.Message != ePost.Message {
		if err := tx.QueryRow(`update posts set isEdit = true, message = $1 where id = $2 returning id, parent, message, isEdit, forum, created, thread, author`, Post.Message, ID).Scan(
			&ePost.ID,
			&ePost.Parent,
			&ePost.Message,
			&ePost.IsEdited,
			&ePost.Forum,
			&pgTime,
			&ePost.Thread,
			&ePost.Author,
		); err != nil {
			check(err)
		}

		tgtimeToString(&pgTime, &ePost.Created)
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

	queryCheck := "select id, forum from threads where"
	var currentErr string

	if _, err := strconv.Atoi(SlugOrID); err != nil {
		currentErr = fmt.Sprintf("Can't find thread by slug: %s", SlugOrID)
		queryCheck += " slug = $1"
	} else {
		currentErr = fmt.Sprintf("Can't find thread by id: %s", SlugOrID)
		queryCheck += " id = $1::bigint"
	}

	var tIdCurrent int32
	var tForumCurrent string
	if err := tx.QueryRow(queryCheck, SlugOrID).Scan(&tIdCurrent, &tForumCurrent); err != nil {
		return nil, errors.New(currentErr)
	}

	if len(Posts) == 0 {
		check("nilPosts")
		return Posts, nil
	}

	query := `insert into posts (parent, message, thread, author, forum) values `
	queryEnd := " returning id, isEdit, created"
	var queryValues []string

	args := make([]interface{}, 0, len(Posts)*5)
	//var args []interface{}
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
	var auth []interface{}
	var querries []string
	auth = append(auth, tForumCurrent)
	for _, value := range Posts {
		if rows.Next() {
			check("insert post")

			pgTime := pgtype.Timestamptz{}
			err = rows.Scan(&value.ID, &value.IsEdited, &pgTime)
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

			tgtimeToString(&pgTime, &value.Created)
		}
	}
	rows.Close()

	if len(par) != 0 {
		tx.Exec(`update posts p set mPath = (select mPath from posts where id = p.parent) || id,
                 rootParent = (select rootParent from posts where id = p.parent)
where id in (` + strings.Join(par, ",") + ")")
	}

	if len(nopar) != 0 {
		tx.Exec("update posts set mPath[1] = id, rootParent = id where id in (" + strings.Join(nopar, ",") + ")")
	}

	tx.Exec("update forums set posts = posts + $1 where slug = $2", len(Posts), tForumCurrent)

	tx.Exec(`insert into forums_users (forum, uid) values `+strings.Join(querries, ",")+`on conflict do nothing`, auth...)

	if err := rows.Err(); err != nil {
		if err.(pgx.PgError).Code == "23503" {
			//log.Println("f")
			return nil, errors.New(fmt.Sprintf("Can't find post author by nickname"))
		}

		check("error on main query")
		//log.Println(err)
		//log.Println(err.(pgx.PgError))
		check(err)
		check(err.(pgx.PgError).Error())
		return nil, errors.New(fmt.Sprintf("Parent post was created in another thread"))
	}

	err = tx.Commit()
	if err != nil {
		check(err)
	}

	return Posts, nil
}

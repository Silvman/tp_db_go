package service

import (
	"fmt"
	"github.com/go-openapi/runtime/middleware"
	"log"
	"github.com/Silvman/tech-db-forum/models"
	"github.com/Silvman/tech-db-forum/restapi/operations"
)

func (self HandlerDB) UserCreate(params operations.UserCreateParams) middleware.Responder {
	tx, err := self.pool.Begin()
	if err != nil {
		log.Fatalln(err)
	}
	defer tx.Rollback()

	rows, _ := tx.Query(`select nickname, fullname, about, email from users where nickname = $1 or email = $2`, params.Nickname, params.Profile.Email)
	eUsers := models.Users{}
	for rows.Next() {
		eUser := models.User{}
		rows.Scan(&eUser.Nickname, &eUser.Fullname, &eUser.About, &eUser.Email)
		eUsers = append(eUsers, &eUser)
	}

	if len(eUsers) != 0 {
		return operations.NewUserCreateConflict().WithPayload(eUsers)
	}

	tx.Exec(`insert into users values ($1, $2, $3, $4)`, params.Nickname, params.Profile.Fullname, params.Profile.About, params.Profile.Email)
	err = tx.Commit()
	if err != nil {
		log.Println(err)
	}

	params.Profile.Nickname = params.Nickname

	return operations.NewUserCreateCreated().WithPayload(params.Profile)
}

func (self HandlerDB) UserGetOne(params operations.UserGetOneParams) middleware.Responder {
	tx, err := self.pool.Begin()
	if err != nil {
		log.Fatalln(err)
	}
	defer tx.Rollback()

	eUser := models.User{}
	if err := tx.QueryRow(`select nickname, fullname, about, email from users where nickname = $1`, params.Nickname).
		Scan(&eUser.Nickname, &eUser.Fullname, &eUser.About, &eUser.Email); err != nil {
		currentErr := models.Error{Message: fmt.Sprintf("Can't find user by nickname: %s", params.Nickname)}
		return operations.NewUserGetOneNotFound().WithPayload(&currentErr)
	}

	err = tx.Commit()
	if err != nil {
		log.Println(err)
	}

	return operations.NewUserGetOneOK().WithPayload(&eUser)
}

func (self HandlerDB) UserUpdate(params operations.UserUpdateParams) middleware.Responder {
	tx, err := self.pool.Begin()
	if err != nil {
		log.Fatalln(err)
	}
	defer tx.Rollback()

	eUser := models.User{}
	if err := tx.QueryRow(`select nickname, fullname, about, email from users where nickname = $1`, params.Nickname).
		Scan(&eUser.Nickname, &eUser.Fullname, &eUser.About, &eUser.Email); err != nil {
		currentErr := models.Error{Message: fmt.Sprintf("Can't find user by nickname: %s", params.Nickname)}
		return operations.NewUserUpdateNotFound().WithPayload(&currentErr)
	}

	query := `update users set`

	args := []interface{}{}
	if params.Profile.Fullname != "" || params.Profile.Email != "" || params.Profile.About != "" {
		if params.Profile.Fullname != "" {
			eUser.Fullname = params.Profile.Fullname
			args = append(args, params.Profile.Fullname)
			query += fmt.Sprintf(" fullname = $%d", len(args))
		}

		if params.Profile.Email != "" {
			if rows, _ := tx.Query(`select nickname from users where email = $1`, params.Profile.Email); rows.Next() == true {
				var nickname string
				rows.Scan(&nickname)
				currentErr := models.Error{Message: fmt.Sprintf("This email is already registered by user: %s", nickname)}
				return operations.NewUserUpdateConflict().WithPayload(&currentErr)
			}

			eUser.Email = params.Profile.Email
			args = append(args, params.Profile.Email)
			query += fmt.Sprintf(" email = $%d", len(args))
		}

		if params.Profile.About != "" {
			eUser.About = params.Profile.About
			args = append(args, params.Profile.About)
			query += fmt.Sprintf(" about = $%d", len(args))
		}
	}

	args = append(args, params.Nickname)
	query += fmt.Sprintf(" where nickname = $%d", len(args))

	if _, err := tx.Exec(query, args...); err != nil {
		log.Println(err)
	}

	err = tx.Commit()
	if err != nil {
		log.Println(err)
	}

	return operations.NewUserCreateCreated().WithPayload(&eUser)
}

package service

import (
	"fmt"
	"github.com/Silvman/tech-db-forum/models"
	"github.com/Silvman/tech-db-forum/restapi/operations"
	"github.com/go-openapi/runtime/middleware"
	"log"
	"strings"
)

func (self HandlerDB) UserCreate(params operations.UserCreateParams) middleware.Responder {
	tx, err := self.pool.Begin()
	if err != nil {
		log.Fatalln(err)
	}
	defer tx.Rollback()

	log.Println("user_create")

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

	_, err = tx.Exec(`insert into users values ($1, $2, $3, $4)`, params.Nickname, params.Profile.Fullname, params.Profile.About, params.Profile.Email)
	if err != nil {
		log.Println(err)
	}

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

	log.Println("user_update")

	tempUser := models.User{}
	if err := tx.QueryRow(`select nickname, fullname, about, email from users where nickname = $1`, params.Nickname).
		Scan(&tempUser.Nickname, &tempUser.Fullname, &tempUser.About, &tempUser.Email); err != nil {
		currentErr := models.Error{Message: fmt.Sprintf("Can't find user by nickname: %s", params.Nickname)}
		return operations.NewUserUpdateNotFound().WithPayload(&currentErr)
	}

	query := `update users set `
	var queryParams []string

	args := []interface{}{}
	if params.Profile.Fullname != "" || params.Profile.Email != "" || params.Profile.About != "" {
		if params.Profile.Fullname != "" {
			tempUser.Fullname = params.Profile.Fullname
			args = append(args, params.Profile.Fullname)
			queryParams = append(queryParams, fmt.Sprintf("fullname = $%d", len(args)))
		}

		if params.Profile.Email != "" {
			var nickname string
			if err := tx.QueryRow(`select nickname from users where email = $1`, params.Profile.Email.String()).Scan(&nickname); err == nil {
				currentErr := models.Error{Message: fmt.Sprintf("This email is already registered by user: %s", nickname)}
				return operations.NewUserUpdateConflict().WithPayload(&currentErr)
			} else {
				log.Println(err)
			}

			tempUser.Email = params.Profile.Email
			args = append(args, params.Profile.Email)
			queryParams = append(queryParams, fmt.Sprintf("email = $%d", len(args)))

		}

		if params.Profile.About != "" {
			tempUser.About = params.Profile.About
			args = append(args, params.Profile.About)
			queryParams = append(queryParams, fmt.Sprintf("about = $%d", len(args)))
		}
	} else {
		return operations.NewUserUpdateOK().WithPayload(&tempUser)
	}

	args = append(args, params.Nickname)
	query += strings.Join(queryParams, ",") + fmt.Sprintf(" where nickname = $%d", len(args)) + " returning nickname, fullname, about, email"

	if err := tx.QueryRow(query, args...).Scan(&tempUser.Nickname, &tempUser.Fullname, &tempUser.About, &tempUser.Email); err != nil {
		log.Println(err)
	}

	err = tx.Commit()
	if err != nil {
		log.Println(err)
	}

	return operations.NewUserUpdateOK().WithPayload(&tempUser)
}

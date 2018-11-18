package mapper

import (
	"errors"
	"fmt"
	"github.com/Silvman/tech-db-forum/models"
	"strings"
)

func (self HandlerDB) UserCreate(Nickname string, Profile *models.User) (*models.Users, error) {
	tx, err := self.pool.Begin()
	if err != nil {
		check(err)
	}
	defer tx.Rollback()

	check("user_create")

	rows, _ := tx.Query(`select nickname, fullname, about, email from users where nickname = $1 or email = $2`, Nickname, Profile.Email)
	eUsers := models.Users{}
	for rows.Next() {
		eUser := models.User{}
		rows.Scan(&eUser.Nickname, &eUser.Fullname, &eUser.About, &eUser.Email)
		eUsers = append(eUsers, &eUser)
	}

	if len(eUsers) != 0 {
		return &eUsers, errors.New("already exists")
	}

	_, err = tx.Exec(`insert into users values ($1, $2, $3, $4)`, Nickname, Profile.Fullname, Profile.About, Profile.Email)
	if err != nil {
		check(err)
	}

	err = tx.Commit()
	if err != nil {
		check(err)
	}

	Profile.Nickname = Nickname

	// !! данные юзера доступны по указателю
	return nil, nil
}

func (self HandlerDB) UserGetOne(Nickname string) (*models.User, error) {
	tx, err := self.pool.Begin()
	if err != nil {
		check(err)
	}
	defer tx.Rollback()

	eUser := models.User{}
	if err := tx.QueryRow(`select nickname, fullname, about, email from users where nickname = $1`, Nickname).
		Scan(&eUser.Nickname, &eUser.Fullname, &eUser.About, &eUser.Email); err != nil {
		return nil, errors.New(fmt.Sprintf("Can't find user by nickname: %s", Nickname))
	}

	err = tx.Commit()
	if err != nil {
		check(err)
	}

	return &eUser, nil
}

func (self HandlerDB) UserUpdate(Nickname string, Profile *models.UserUpdate) (*models.User, error) {
	tx, err := self.pool.Begin()
	if err != nil {
		check(err)
	}
	defer tx.Rollback()

	check("user_update")

	tempUser := models.User{}
	if err := tx.QueryRow(`select nickname, fullname, about, email from users where nickname = $1`, Nickname).
		Scan(&tempUser.Nickname, &tempUser.Fullname, &tempUser.About, &tempUser.Email); err != nil {
		return nil, errors.New(fmt.Sprintf("Can't find user by nickname"))

	}

	query := `update users set `
	var queryParams []string

	args := []interface{}{}
	if Profile.Fullname != "" || Profile.Email != "" || Profile.About != "" {
		if Profile.Fullname != "" {
			tempUser.Fullname = Profile.Fullname
			args = append(args, Profile.Fullname)
			queryParams = append(queryParams, fmt.Sprintf("fullname = $%d", len(args)))
		}

		if Profile.Email != "" {
			var nickname string
			if err := tx.QueryRow(`select nickname from users where email = $1`, Profile.Email).Scan(&nickname); err == nil {
				return nil, errors.New(fmt.Sprintf("This email is already registered by user: %s", nickname))
			}

			tempUser.Email = Profile.Email
			args = append(args, Profile.Email)
			queryParams = append(queryParams, fmt.Sprintf("email = $%d", len(args)))

		}

		if Profile.About != "" {
			tempUser.About = Profile.About
			args = append(args, Profile.About)
			queryParams = append(queryParams, fmt.Sprintf("about = $%d", len(args)))
		}
	} else {
		return &tempUser, nil
	}

	args = append(args, Nickname)
	query += strings.Join(queryParams, ",") + fmt.Sprintf(" where nickname = $%d", len(args)) + " returning nickname, fullname, about, email"

	if err := tx.QueryRow(query, args...).Scan(&tempUser.Nickname, &tempUser.Fullname, &tempUser.About, &tempUser.Email); err != nil {
		check(err)
	}

	err = tx.Commit()
	if err != nil {
		check(err)
	}

	return &tempUser, nil
}

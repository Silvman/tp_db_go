package mapper

import (
	"errors"
	"fmt"
	"github.com/Silvman/tech-db-forum/models"
)

const qInsertUser = `insert into users values ($1, $2, $3, $4)`
const qSelectUserByNickEmail = `select nickname, fullname, about, email from users where nickname = $1 or email = $2`

const qUpdateUserFullname = `update users set fullname = $1 where nickname = $2`
const qUpdateUserFullnameAbout = `update users set fullname = $1,about = $2 where nickname = $3`
const qUpdateUserFullnameEmail = `update users set fullname = $1,email = $2 where nickname = $3`
const qUpdateUserFullnameEmailAbout = `update users set fullname = $1,email = $2,about = $3 where nickname = $4`
const qUpdateUserAbout = `update users set about = $1 where nickname = $2`
const qUpdateUserEmail = `update users set email = $1 where nickname = $2`
const qUpdateUserEmailAbout = `update users set email = $1,about = $2 where nickname = $3`

//returning nickname, fullname, about, email

func (self HandlerDB) UserCreate(Nickname string, Profile *models.User) (*models.Users, error) {
	rows, _ := self.pool.Query(qSelectUserByNickEmail, Nickname, Profile.Email)
	eUsers := models.Users{}
	for rows.Next() {
		eUser := models.User{}
		rows.Scan(&eUser.Nickname, &eUser.Fullname, &eUser.About, &eUser.Email)
		eUsers = append(eUsers, &eUser)
	}

	if len(eUsers) != 0 {
		return &eUsers, errors.New("already exists")
	}

	_, err := self.pool.Exec(qInsertUser, Nickname, Profile.Fullname, Profile.About, Profile.Email)
	if err != nil {
		//log.Println(err)
	}

	Profile.Nickname = Nickname

	// !! данные юзера доступны по указателю
	return nil, nil
}

func (self HandlerDB) UserGetOne(Nickname string) (*models.User, error) {
	eUser := models.User{}
	if err := self.pool.QueryRow(qSelectUserByNick, Nickname).
		Scan(&eUser.Nickname, &eUser.Fullname, &eUser.About, &eUser.Email); err != nil {
		return nil, errors.New(fmt.Sprintf("Can't find user by nickname: %s", Nickname))
	}

	return &eUser, nil
}

func (self HandlerDB) UserUpdate(Nickname string, Profile *models.UserUpdate) (*models.User, error) {
	tempUser := models.User{}
	if err := self.pool.QueryRow(qSelectUserByNick, Nickname).Scan(&tempUser.Nickname, &tempUser.Fullname, &tempUser.About, &tempUser.Email); err != nil {
		return nil, errors.New(fmt.Sprintf("Can't find user by nickname"))
	}

	var err error
	if Profile.Email != "" {
		if Profile.About != "" {
			if Profile.Fullname != "" {
				_, err = self.pool.Exec(qUpdateUserFullnameEmailAbout, Profile.Fullname, Profile.Email, Profile.About, Nickname)
				tempUser.Fullname = Profile.Fullname
				tempUser.About = Profile.About
				tempUser.Email = Profile.Email
			} else {
				_, err = self.pool.Exec(qUpdateUserEmailAbout, Profile.Email, Profile.About, Nickname)
				tempUser.About = Profile.About
				tempUser.Email = Profile.Email
			}
		} else {
			if Profile.Fullname != "" {
				_, err = self.pool.Exec(qUpdateUserFullnameEmail, Profile.Fullname, Profile.Email, Nickname)
				tempUser.Fullname = Profile.Fullname
				tempUser.Email = Profile.Email
			} else {
				_, err = self.pool.Exec(qUpdateUserEmail, Profile.Email, Nickname)
				tempUser.Email = Profile.Email
			}
		}
	} else {
		if Profile.About != "" {
			if Profile.Fullname != "" {
				_, err = self.pool.Exec(qUpdateUserFullnameAbout, Profile.Fullname, Profile.About, Nickname)
				tempUser.Fullname = Profile.Fullname
				tempUser.About = Profile.About
			} else {
				_, err = self.pool.Exec(qUpdateUserAbout, Profile.About, Nickname)
				tempUser.About = Profile.About
			}
		} else {
			if Profile.Fullname != "" {
				_, err = self.pool.Exec(qUpdateUserFullname, Profile.Fullname, Nickname)
				tempUser.Fullname = Profile.Fullname
			}
		}
	}

	if err != nil {
		return nil, errors.New(fmt.Sprintf("This email is already registered"))
	}

	return &tempUser, nil
}

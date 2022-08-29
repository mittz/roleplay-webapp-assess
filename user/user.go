package user

import (
	"context"
	"log"

	"github.com/jackc/pgx"
	"github.com/mittz/roleplay-webapp-assess/database"
)

type User struct {
	Userkey   string
	LDAP      string
	Team      string
	Region    string
	SubRegion string
	Role      string
}

func GetUser(userkey string) User {
	dbPool := database.GetDatabaseConnection()

	user := User{}
	if err := dbPool.QueryRow(context.Background(), "select * from users where userkey=$1", userkey).Scan(
		&user.Userkey,
		&user.LDAP,
		&user.Team,
		&user.Region,
		&user.SubRegion,
		&user.Role,
	); err != nil && err != pgx.ErrNoRows {
		log.Printf("QueryRow failed: %v\n", err)
	}

	return user
}

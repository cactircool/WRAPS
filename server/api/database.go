package api

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

type User struct {
	Username string `json:"username"`
	Email string `json:"email"`
	HashedPassword string `json:"hashedPassword"`
}

var db *sql.DB = nil
var env map[string]string = initEnvironment()

func Connect() error {
	source := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?interpolateParams=true", env["SQL_USER"], env["SQL_PASSWORD"], env["SQL_IP_ADDR"], env["SQL_PORT"], env["SQL_DB_NAME"])
	myDB, err := sql.Open("mysql", source)
	if err != nil {
		return err
	}
	db = myDB
	return nil
}

func Close() {
	db.Close()
	db = nil
}

// This server will only manage authorization, since all users must have been already created for
// them to even be at this stage, maybe later you can add a sign up here folks or wtvr the fuck
func AuthorizeUser(username string, password string) (User, error) {
	var user User
	query, err := db.Query("SELECT * FROM users WHERE username = ?", username)
	if err != nil {
		fmt.Println(err)
		return user, err
	}
	defer query.Close()

	// Will only iterate once since username is a primary key
	hasher := sha256.New()
	for query.Next() { 
		var salt string
		err = query.Scan(&user.Username, &user.Email, &user.HashedPassword, &salt)
		if err != nil {
			return user, err
		}
		if hasher.Write([]byte(password + salt)); hex.EncodeToString(hasher.Sum(nil)) != user.HashedPassword {
			return user, errors.New("invalid password")
		}
		return user, nil
	}
	return user, errors.New("invalid username")
}

func GetDB() *sql.DB { return db; }

func initEnvironment() map[string]string {
	a, err := godotenv.Read(".env")
	if err != nil {
		log.Fatal(err)
	}
	return a
}
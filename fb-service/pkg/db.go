package fb_service

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"os"
)

var DB *sql.DB

func ConnectDatabase() (dbHandler *sql.DB) {
	log.Println("Connecting to database...")

	mysqlPwd := os.Getenv("MYSQL_PWD")
	// get a handle to the database
	db, err := sql.Open("mysql", fmt.Sprintf("root:%s@tcp(localhost:3306)/users", mysqlPwd))
	DB = db
	if err != nil {
		log.Fatalf("Unable to connect to database: %s", err.Error())
	}
	// defer DB.Close()
	defer func(DB *sql.DB) {
		err := DB.Close()
		if err != nil {
			return
		}
	}(DB)

	// sql.Open() doesn't open a connection. validate using DB.Ping()
	s, err := PingDatabase(DB)
	if err != nil {
		log.Fatalf("Unable to ping the db: %s", err)
	} else {
		log.Println(s)
		return DB
	}
	return
}

func PingDatabase(db *sql.DB) (message string, string error) {
	err := db.Ping()
	if err != nil {
		return "", errors.New(fmt.Sprintf("Unable to ping the DB: %s", err.Error()))
	}

	return "Pinged db successfully", nil
}

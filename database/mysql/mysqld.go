package mysqld

import (
	"database/sql"
)

var DB *sql.DB


func SetDB(db *sql.DB) {
	DB = db
}

func GetDB() *sql.DB {
	return DB
}
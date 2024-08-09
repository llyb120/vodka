package tests

import (
	"database/sql"
	"fmt"
	"testing"
	database "vodka/database"
)

var _db *sql.DB

func TestConnectMySQL(t *testing.T) {
	t.Run("连接MySQL", func(t *testing.T) {
		ConnectMySQL(t)
	})

	t.Run("查询Map", func(t *testing.T) {
		ConnectMySQL(t)
		users, err := database.QueryMap(_db, "select * from user where id = ?", 1)
		if err != nil {
			t.Errorf("查询失败: %v", err)
		}
		fmt.Println(users)
	})

	t.Run("查询Struct", func(t *testing.T) {
		ConnectMySQL(t)
		var users []*User
		user := &User{}
		err := database.QueryStruct(_db, "select * from user where id = ?", []interface{}{1}, []interface{}{&users, &user})
		if err != nil {
			t.Errorf("查询失败: %v", err)
		}
	})

	t.Run("查询Struct为空", func(t *testing.T) {
		ConnectMySQL(t)
		user := &User{Id: 10}
		err := database.QueryStruct(_db, "select * from user where id = ?", []interface{}{10086}, []interface{}{&user})
		if err != nil {
			t.Errorf("查询失败: %v", err)
		}
		if user != nil {
			t.Errorf("查询失败: %v", user)
		}
		fmt.Println(user)
	})
}

func ConnectMySQL(t *testing.T) *sql.DB {
	dsn := "root:root@tcp(127.0.0.1:3306)/test?charset=utf8mb4&parseTime=True&loc=Local"
	var err error
	_db, err = database.ConnectMySQL(dsn)
	if err != nil {
		t.Fatalf("连接MySQL失败: %v", err)
	}
	database.SetDB(_db)
	return _db
}

package tests

import (
	"testing"
	"vodka"
	"vodka/database"
)

// type User struct {
// 	Id   int
// 	Name string
// }

// type UserMapper struct {
// 	GetUserById func(id int) *User
// 	GetUsers    func() []*User
// }

var userMapper *UserMapper = &UserMapper{}

func TestVodka(t *testing.T) {
	// tests := []struct {
	// 	name string
	// 	want string
	// }{
	// 	{name: "test1", want: "test1"},
	// }
	// for _, tt := range tests {
	// 	t.Run(tt.name, func(t *testing.T) {
	// 		if got := tt.want; got != tt.want {
	// 			t.Errorf("%q. Testvodka() = %v, want %v", tt.name, got, tt.want)
	// 		}
	// 	})
	// }

	t.Run("test1", func(t *testing.T) {
		Prepare(t)
		users, user, err := userMapper.GetUsers()
		if err != nil {
			t.Fatalf("GetUsers() error = %v", err)
		}
		println(users)
		println(user)
	})

	t.Run("test2", func(t *testing.T) {
		Prepare(t)
		user := userMapper.GetUserById(1)
		if user == nil {
			t.Fatalf("GetUserById() error = %v", user)
		}
		user = userMapper.GetUserById(10086)
		if user != nil {
			t.Fatalf("GetUserById() error = %v", user)
		}
		println("test1")
		println(user)
	})
}

func Prepare(t *testing.T) {
	ConnectMySQL(t)
	_db := ConnectMySQL(t)
	// 设置数据库
	database.SetDB(_db)
	// 扫描mapper
	vodka.ScanMapper("./mapper")

	err := vodka.InitMapper(userMapper)
	if err != nil {
		t.Fatalf("InitMapper() error = %v", err)
	}
}

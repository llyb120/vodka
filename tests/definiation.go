package tests

type User struct {
	Id   int64  `vo:"id"`
	Name string `vo:"name"`
	Age  int    `vo:"age"`
}

type UserMapper struct {
	GetUser          func(ids []int) (*User, error) `params:"ids"`
	GetUsers         func() ([]*User, *User, error)
	GetUserById      func(id int) *User                `params:"id"`
	GetUsersInIds    func(ids []int) ([]*User, error)  `params:"ids"`
	// 使用结构体作为参数进行查询
	GetUsersByStruct func(user User) ([]*User, error)  `params:"user"`
	// 复用xml中的语句
	GetUsersByPtr    func(user *User) ([]*User, error) `params:"user" xml:"GetUsersByStruct"`
	// 自定义sql
	GetUsersByCustomSql func(user *User) ([]*User, error) `params:"user" sql:"select * from user where id = #{id}"`

	// 分别为影响的行数，最后插入的id，错误
	Insert      func(user *User) (int64, int64, error)    `params:"user"`
	InsertBatch func(users []*User) (int64, int64, error) `params:"users"`

	// 分别为影响的行数，最后插入的id，错误
	UpdateById func(user *User) (int64, int64, error) `params:"user"`
	// UpdateBatch func(users []*User) (int64, int64, error) `params:"users"`

	TestInclude func() error 

}

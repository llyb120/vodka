package tests

import (
	"fmt"
	"testing"
	mapper "vodka/mapper"
)

func TestMapper(t *testing.T) {

	t.Run("测试mapper缓存", func(t *testing.T) {
		prepare(t)
		var user UserMapper
		err := mapper.InitMapper(&user)
		if err != nil {
			t.Errorf("第一次获取mapper失败: %v", err)
		}
		var user2 UserMapper
		err = mapper.InitMapper(&user2)
		if err != nil {
			t.Errorf("第二次获取mapper失败: %v", err)
		}

		fmt.Printf("%p\n", &user)
		fmt.Printf("%p\n", &user2)

		// if reflect.ValueOf(user).Addr().Pointer() != reflect.ValueOf(user2).Addr().Pointer() {
		// 	t.Errorf("两次获取的mapper不是同一个对象")
		// }
		// t.Log(mapper)

		// userMapper.GetUserById(10)
	})

	t.Run("测试mapper查询", func(t *testing.T) {
		prepare(t)
		users, err := userMapper.GetUsersInIds([]int{1, 2, 3})
		if err != nil {
			t.Errorf("获取用户失败: %v", err)
		}
		t.Log(users)
	})

	t.Run("测试mapper结构体查询", func(t *testing.T) {
		prepare(t)
		user, err := userMapper.GetUsersByStruct(User{Id: 1, Name: "test", Age: 18})
		if err != nil {
			t.Errorf("获取用户失败: %v", err)
		}
		t.Log(user)
	})

	t.Run("测试mapper指针查询", func(t *testing.T) {
		prepare(t)
		user, err := userMapper.GetUsersByPtr(&User{Id: 1, Name: "test", Age: 18})
		if err != nil {
			t.Errorf("获取用户失败: %v", err)
		}
		t.Log(user)
	})

	t.Run("测试mapper自定义sql查询", func(t *testing.T) {
		prepare(t)
		user, err := userMapper.GetUsersByCustomSql(&User{Id: 1})
		if err != nil {
			t.Errorf("获取用户失败: %v", err)
		}
		t.Log(user)
	})

	t.Run("测试mapper插入", func(t *testing.T) {
		prepare(t)
		rows, id, err := userMapper.Insert(&User{Name: "test", Age: 18})
		if err != nil {
			t.Errorf("插入用户失败: %v", err)
		}
		t.Log(rows, id)
	})

	t.Run("测试mapper更新", func(t *testing.T) {
		prepare(t)
		rows, id, err := userMapper.UpdateById(&User{Id: 1, Name: "test", Age: 18})
		if err != nil {
			t.Errorf("插入用户失败: %v", err)
		}
		t.Log(rows, id)
	})

	t.Run("测试mapper include", func(t *testing.T) {
		prepare(t)
		err := userMapper.TestInclude()
		if err != nil {
			t.Errorf("测试include失败: %v", err)
		}
	})

}

func prepare(t *testing.T) {
	Prepare(t)
}

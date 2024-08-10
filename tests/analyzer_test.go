package tests

import (
	"testing"
	analyzer "vodka/analyzer"
)

const xmlContent = `
<mapper namespace="UserRepo">
    <select id="GetUser" resultType="User">
        SELECT id, name FROM user
        <where>
            and id in (
                <foreach collection="ids" item="id" separator=",">
                    #{id}
                </foreach>
            )
            <if test="name != null">
                and name = #{name}
            </if>
        </where>
    </select>

    <insert id="InsertUser" parameterType="User">
        INSERT INTO user (id, name) VALUES (
			#{id == 0 ? 'DEFAULT' : id},
			#{name}
		)
    </insert>


    <insert id="InsertUserBatch" parameterType="User">
        INSERT INTO user (id, name) VALUES
        <foreach collection="users" item="user" separator=",">
            <if test="user.Age >= 10" >
                (#{user.Id}, #{user.Name})
            </if>
        </foreach>
    </insert>

    <select id="GetUsersByIds" resultType="User">
        SELECT id, name FROM user WHERE id in (
            <foreach collection="ids" item="id" separator=",">
                <if test="id < 2">
                    #{id}
                </if>
            </foreach>
        )
    </select>
</mapper>
`

type User struct {
	Id   int    `vo:"id"`
	Name string `vo:"name"`
	Age  int    `vo:"age"`
}

type UserMapper struct {
	GetUser       func(ids []int) (*User, error) `params:"ids"`
	GetUsers      func() ([]*User, *User, error)
	GetUserById   func(id int) *User               `params:"id"`
	GetUsersInIds func(ids []int) ([]*User, error) `params:"ids"`

	// 分别为影响的行数，最后插入的id，错误
	Insert      func(user *User) error                    `params:"user"`
	InsertBatch func(users []*User) (int64, int64, error) `params:"users"`

	// 分别为影响的行数，最后插入的id，错误
	Update func(user *User) (int64, int64, error) `params:"user"`
	// UpdateBatch func(users []*User) (int64, int64, error) `params:"users"`
}

func TestAnalyzer(t *testing.T) {
	tests := []struct {
		name     string
		funcName string
		params   map[string]interface{}
		expected func(result interface{}) bool
	}{
		{
			name:     "测试where",
			funcName: "GetUser",
			params:   map[string]interface{}{"ids": []int{1, 2, 3}, "name": "张三"},
			//expected: func(result interface{}) bool {
			//	// 这里暂时result是string
			//	// 查到第一个where单词，where后第一个单词不是and
			//	sql := result.(string)
			//	whereIndex := strings.Index(strings.ToLower(sql), "where")
			//	if whereIndex == -1 {
			//		return false
			//	}
			//
			//	// 跳过"where"后的空白字符
			//	afterWhere := strings.TrimSpace(sql[whereIndex+5:])
			//
			//	// 检查"where"后的第一个单词是否为"and"
			//	firstWord := strings.Fields(afterWhere)[0]
			//	return strings.ToLower(firstWord) != "and"
			//},
		},
		{
			name:     "测试foreach",
			funcName: "GetUsersByIds",
			params:   map[string]interface{}{"ids": []int{1, 2, 3}},
		},
		{
			name:     "测试if",
			funcName: "InsertUserBatch",
			params:   map[string]interface{}{"users": []User{{Id: 0, Name: "张三", Age: 20}, {Id: 2, Name: "李四", Age: 9}}},
		},
		{
			name:     "测试三元表达式",
			funcName: "InsertUser",
			params:   map[string]interface{}{"user": User{Id: 0, Name: "张三", Age: 20}},
		},
	}

	ConnectMySQL(t)
	newAnalyzer := analyzer.NewAnalyzer(xmlContent)
	newAnalyzer.Parse()
	// newAnalyzer.BindMapper(&userMapper)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var user *User = &User{}
			err := newAnalyzer.Call(test.funcName, test.params, []interface{}{&user})
			if err != nil {
				t.Errorf("测试失败: %v", err)
			}
			if test.expected == nil {
				return
			}
			if !test.expected(user) {
				t.Errorf("测试失败")
			}
		})
	}

}

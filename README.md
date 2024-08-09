# Vodka

Vodka是Go的一个轻量级的半自动化ORM框架，灵感来自MyBatis。

希望可以和go中的gin框架，与柯南中的Gin同Vodka一样，成为形影不离的好伙伴。

## 特性

- 简单易用 
- 天生防注入
- SQL和代码分离，方便管理
- 支持用于复杂查询的动态SQL
- 支持缓存 (开发中)
- 支持插件 (开发中)
- 支持定义切面，插入自己的权限语句（开发中）
- 对基础查询语句直接自动装配，无需再书写xml文件（开发中）

## 快速上手

### 定义你的model
```go
type User struct {
    Id   int    `gb:"id"` //gb标签表示对应数据库中的字段 
    Name string `gb:"name"`
    Age  int    `gb:"age"`
}
```

### 定义你的mapper接口 (无需实现，vodka会自动装配这些方法)
```go
type UserMapper struct {
    Select(params interface{}) ([]*User, error) `param:"params"` //params为在xml中映射的名字
    Insert(user *User) (int64, error) `param:"user"`
    InsertBatch(users []*User) (int64, error) `param:"users"`
    Update(user *User) (int64, error) `param:"user"`
    Delete(id int) (int64, error) `param:"id"`
}
```

### 定义你的xml映射文件
- 可直接使用mybatis的工具生成，无需繁琐的书写步骤
- 针对复杂查询，在xml中和原生sql书写并无二至，只需要附加你的条件即可
```xml
<mapper namespace="UserMapper">
    <!-- 复杂查询 -->
    <select id="Select" resultType="User">
        SELECT * FROM users 
        <where>
            <if test="id != null">
                and id = #{id}
            </if>

            <if test="in_ids != null">
                and id in (
                    <foreach collection="in_ids" item="id" separator=",">
                        #{id}
                    </foreach>
                )
            </if>

            <!-- 其他条件 -->
            <!-- 具体可参照mybatis文档 -->
        </where>
    </select>

    <insert id="Insert" >
        INSERT INTO users (name, age) VALUES (#{name}, #{age})
    </insert>
    
    <insert id="insertBatch">
        INSERT INTO users (name, age) VALUES (
            <foreach collection="users" item="user" separator=",">
                (#{user.name}, #{user.age})
            </foreach>
        )
    </insert>

    <update id="Update" >
        UPDATE users SET name = #{name}, age = #{age} WHERE id = #{id}
    </update>

    <delete id="Delete" >
        DELETE FROM users WHERE id = #{id}
    </delete>

</mapper>
```

### 初始化
- 在系统初始化时，调用`vodka.ScanMapper("你的xml路径文件夹")`方法进行初始化

- 获取上面定义的mapper
```go
var userMapper UserMapper
vodka.InitMapper(&userMapper)

// 使用
// 使用map作为参数进行复杂查询
userMapper.Select(map[string]interface{}{"id": 1})
// 使用struct作为参数
userMapper.Select(User{Id: 1})

// 插入
userMapper.Insert(&User{Name: "张三", Age: 18})
// 更新
userMapper.Update(&User{Id: 1, Name: "李四", Age: 20})
// 删除
userMapper.Delete(1)
```


### 标签说明
 
- mapper：定义命名空间，每个xml根节点都要有，相同的命名空间会合并成一个
- select: 定义查询语句
- insert：定义插入语句
- update：定义更新语句
- delete：定义删除语句
- where：定义查询条件，使用该标签，可直接使用and进行条件拼装，无需判断在第一个条件上不使用and
- if：定义表达式判断，符合test的表达式才会生效

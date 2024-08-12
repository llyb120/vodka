# Vodka

Vodka是Go的一个轻量级的半自动化ORM框架，灵感来自MyBatis。

希望可以和go中的gin框架，与柯南中的Gin同Vodka一样，成为形影不离的好伙伴。

![](img/banner.png)

## 特性

- 简单，开箱即用 
- 天生防注入
- SQL和代码分离，方便管理
- 支持用于复杂查询的动态SQL
- 对基础查询语句直接自动装配，无需再书写xml文件
- 支持缓存 (开发中)
- 支持插件 (开发中)
- 支持定义切面，插入自己的权限语句（开发中）


## 快速上手

### 定义你的model
```go
type User struct {
    Id   int    `vo:"id"` //vo标签表示对应数据库中的字段 
    Name string `vo:"name"`
    Age  int    `vo:"age"`
}
```



### 定义你的mapper接口 (无需实现，vodka会自动装配这些方法)
```go
type UserMapper struct {
    // 基础查询
    Select(params interface{}) ([]*User, error) `param:"params"` //params为在xml中映射的名字
    // 部分无需xml的情况，可以直接通过tag自定义sql
    SelectByCustomSql(params interface{}) ([]*User, error) `param:"params" sql:"select * from user where id = #{id}"`
    // 插入
    // insert语句最多支持3个返回值，分别为影响的行数、自增主键、错误
    Insert(user *User) (int64, int64, error) `param:"user"`
    InsertBatch(users []*User) (int64, error) `param:"users"`
    // 更新
    // 更新语句最多支持2个返回值，分别为影响的行数、错误
    Update(user *User) (int64, error) `param:"user"`
    // 删除
    // 删除语句最多支持2个返回值，分别为影响的行数、错误
    Delete(id int) (int64, error) `param:"id"`
}
```

### 定义你的xml映射文件
- 可直接使用mybatis的工具生成，无需繁琐的书写步骤
- 针对复杂查询，在xml中和原生sql书写并无二致，只需要附加你的条件即可
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
- foreach 循环语句，分为collection/item/separator/open/close 五个属性
- where：定义查询条件，使用该标签，可直接使用and进行条件拼装，无需判断在第一个条件上不使用and
- if：定义表达式判断，符合test的表达式才会生效
- set: 定义更新语句中的set部分，使用该标签，可直接在每条语句后拼装逗号，无需检查最后一个是否拼装
- sql: 定义sql语句，抽象出公共的模块，可以供include引用
- include: 引用sql语句，可以简单理解为文本替换


### 通用Mapper
- 直接继承mapper.VodkaMapper，即可拥有通用Mapper的所有功能
- 以下基础语句会自动装配，无需再书写xml文件
```go
type VodkaMapper[T any, ID any] struct {
	InsertOne            func(params *T) (int64, int64, error)                                                      `params:"params"`
	InsertBatch          func(params []*T) (int64, int64, error)                                                    `params:"params"`
	UpdateById           func(params *T) (int64, error)                                                             `params:"params"`
	UpdateSelectiveById  func(params *T) (int64, error)                                                             `params:"params"`
	UpdateByCondition    func(condition *T, action *T) (int64, error)                                               `params:"condition,action"`
	UpdateByConditionMap func(condition map[string]interface{}, action map[string]interface{}) (int64, error)       `params:"condition,action"`
	DeleteById           func(id ID) (int64, error)                                                                 `params:"id"`
	SelectById           func(id ID) (*T, error)                                                                    `params:"id"`
	SelectAll            func(params *T, order string, offset int64, limit int64) ([]*T, error)                     `params:"params,order,offset,limit"`
	CountAll             func(params *T) (int64, error)                                                             `params:"params"`
	SelectAllByMap       func(params map[string]interface{}, order string, offset int64, limit int64) ([]*T, error) `params:"params,order,offset,limit"`
	CountAllByMap        func(params map[string]interface{}) (int64, error)                                         `params:"params"`
}

// 示例
type UserMapper struct {
    mapper.VodkaMapper[User, int64]
    // 需要额外书写表名和主键定义
    _ any `table:"user" pk:"id"`
}

//test
var userMapper UserMapper
vodka.InitMapper(&userMapper)

userMapper.InsertOne(&User{Name:"张三"})

// ByMap系列方法可以使用多种策略参数，例如GT_EQ、LT_EQ、GT、LT、EQ、NE、LIKE、IN、NOT_IN、BETWEEN、NOT_BETWEEN等
userMapper.SelectAllByMap(map[string]interface{}{"GTE_age": 18, "name": "张三"}, "", 0, 10)
```


### 其余说明
- GO中在insert语句中，无法直接使用nil，所以如果你需要在insert语句中使用自增主键，可以这么写，假如主键为int64，以下写法同时可以满足自增主键和非自增主键，当然，如果你只使用自增主键，最好的方法是不对主键写插入
```xml
<insert id="Insert">
    insert into user (id, name, age) values (
        #{id == 0 ? $AUTO : id},
        #{name}, #{age}
    )
</insert>
```
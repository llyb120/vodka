package tests

import (
	"fmt"
	"testing"
	xml "vodka/xml"
)

func TestLexer(t *testing.T) {
	xmlData := `
<mapper namespace="UserRepo">
    <select id="GetUser" resultType="User">
        SELECT id, name FROM users 
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
        INSERT INTO users (id, name) VALUES (#{id}, #{name})
    </insert>
    <insert id="InsertUserBatch" parameterType="User">
        INSERT INTO users (id, name) VALUES
        <foreach collection="users" item="user" separator=",">
            <if test="user.Age >= 10" >
                (#{user.Id}, #{user.Name})
            </if>
        </foreach>
    </insert>

    <select id="GetUsersByIds" resultType="User">
        SELECT id, name FROM users WHERE id in (
            <foreach collection="ids" item="id" separator=",">
                <if test="<![CDATA[ id < 2 ]]>">
                    #{id}
                </if>
            </foreach>
        )
    </select>

    <sql id="UserColumns" >
        id, name
    </sql>

    <include refid="UserColumns" />
</mapper>
`

	parser := xml.NewParser(xmlData)
	root, err := parser.Parse()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Print(root)

	//parser.PrintNode(root, "")
}

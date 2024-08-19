package tests

import (
	"testing"
	"vodka/plugin/page"
)

func TestPage(t *testing.T) {
	// plugin.NewPlugin()
	//t.Run("test plugin", func(t *testing.T) {
	//	plugin.RegisterNode("PERMISSION", func(builder *strings.Builder, node *xml.Node, params map[string]any, resultParams *[]any, root *xml.Node) {
	//		builder.WriteString("PERMISSION")
	//	})
	//})

	t.Run("测试分页查询", func(t *testing.T) {
		Prepare(t)

		var pg page.Page
		pg.PageNum = 1
		pg.PageSize = 10
		pg.Sort = "id desc"
		err := page.DoPage(&pg, func() {
			userMapper.GetUsers()
		})
		if err != nil {
			t.Fatal(err)
		}

		//if err != nil {
		//	t.Fatal(err)
		//}
		//log.Println(page.List)
		// var page plugin.Page
		// page.PageNum = 2
		// page.PageSize = 10
		// plugin.StartPage(&page)
		// defer plugin.EndPage()
		// userMapper.GetUsers()
		// log.Println(page.List)
	})

	// pm := StartPage()
	// defer EndPage()

}

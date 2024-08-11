package tests

import (
	"fmt"
	"testing"
	"vodka"
	"vodka/database"
	mapper "vodka/mapper"
)

type DepMapper struct {
	mapper.VodkaMapper[Dep, int64]
	_ struct{} `table:"dep" pk:"id"`
}

type Dep struct {
	Id    int64  `vo:"id"`
	Name  string `vo:"name"`
	Descr string `vo:"descr"`
}

func TestBaseMapper(t *testing.T) {

	t.Run("测试通用Mapper", func(t *testing.T) {
		baseMapperPrepare(t)
		//var dep Dep
		//result, err := dep.BuildTags()
		//if err != nil {
		//	t.Fatal(err)
		//}

		rows, _, err := depMapper.InsertOne(&Dep{Name: "heihei"})
		if err != nil {
			t.Fatal(err)
		}
		if rows < 0 {
			t.Fatal("insert failed")
		}
		rows, _, err = depMapper.InsertBatch([]*Dep{
			{Name: "dep1"},
			{Name: "dep2"},
		})
		if err != nil {
			t.Fatal(err)
		}
		if rows < 0 {
			t.Fatal("insert failed")
		}
		depMapper.UpdateSelectiveById(&Dep{Id: 1, Name: "dep111"})

		deps, err := depMapper.SelectAll(&Dep{Name: "dep1"}, "id desc", 0, 10)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println(deps)
	})

	t.Run("selectAll测试", func(t *testing.T) {
		baseMapperPrepare(t)
		deps, err := depMapper.SelectAll(&Dep{Name: "dep1"}, "id desc", 0, 10)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println(deps)
	})

	t.Run("countAll测试", func(t *testing.T) {
		baseMapperPrepare(t)
		count, err := depMapper.CountAll(&Dep{Name: "dep1"})
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println(count)
	})

	t.Run("selectAllByMap测试", func(t *testing.T) {
		baseMapperPrepare(t)
		deps, err := depMapper.SelectAllByMap(map[string]interface{}{"IN_name": []string{"dep1", "dep2"}}, "id desc", 0, 10)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println(deps)
	})

	t.Run("updateByCondition测试", func(t *testing.T) {
		baseMapperPrepare(t)
		rows, err := depMapper.UpdateByCondition(&Dep{Name: "dep111"}, &Dep{Name: "dep1"})
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println(rows)
	})

	t.Run("updateByConditionMap测试", func(t *testing.T) {
		baseMapperPrepare(t)
		rows, err := depMapper.UpdateByConditionMap(map[string]interface{}{"IN_name": []string{"dep1", "dep2"}}, map[string]interface{}{"name": "dep111"})
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println(rows)
	})

}

var baseMapperInited = false
var depMapper *DepMapper

func baseMapperPrepare(t *testing.T) {
	if baseMapperInited {
		return
	}
	ConnectMySQL(t)
	_db := ConnectMySQL(t)
	// 设置数据库
	database.SetDB(_db)
	// 扫描mapper
	vodka.ScanMapper("./mapper")

	depMapper = &DepMapper{}
	err := vodka.InitMapper(depMapper)
	if err != nil {
		t.Fatal(err)
	}

	baseMapperInited = true
}

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
		depMapper := &DepMapper{}
		err := vodka.InitMapper(depMapper)
		if err != nil {
			t.Fatal(err)
		}
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

		fmt.Println(depMapper)
	})

}

var baseMapperInited = false

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

	baseMapperInited = true
}

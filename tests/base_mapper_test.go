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
	Id   int64  `vo:"id"`
	Name string `vo:"name"`
}

func TestBaseMapper(t *testing.T) {

	t.Run("TestBaseMapper", func(t *testing.T) {
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
		//depMapper.InsertOne(&Dep{Name: "heihei"})
		depMapper.InsertBatch([]*Dep{
			{Name: "dep1"},
			{Name: "dep2"},
		})

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

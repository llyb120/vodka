package page

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"reflect"
	"vodka/database"
	"vodka/util"
)

type Page[T any] struct {
	PageNum    int64
	PageSize   int64
	TotalRows  int64
	TotalPages int64
	Sort       string
	List       []*T
}

var ThreadLocal = util.NewThreadLocal(false)

// func (p *Page[T]) GetTotalPages() int64 {
// 	return (p.TotalRows + int64(p.PageSize-1)) / int64(p.PageSize)
// }

// func (p *Page[T]) HasNext() bool {
// 	return p.PageNum < p.GetTotalPages()
// }

// func (p *Page[T]) HasPrevious() bool {
// 	return p.PageNum > 1
// }

// func (p *Page[T]) GetNextPageNum() int64 {
// 	return p.PageNum + 1
// }

// func (p *Page[T]) GetPreviousPageNum() int64 {
// 	return p.PageNum - 1
// }

// func (p *Page[T]) GetOffset() int64 {
// 	return (p.PageNum - 1) * p.PageSize
// }

func DoPage[T any](page *Page[T], fun func()) error {
	ThreadLocal.Set(page)
	defer ThreadLocal.Remove()
	fun()
	return nil
}

func GetPageContext() interface{} {
	value, _ := ThreadLocal.Get()
	if value == nil {
		return nil
	}
	return value
	//page, ok := value.(*Page[T])
	//if !ok {
	//	return nil
	//}
	//return page
}

func SelectTotal(db *sql.DB, sql string, args ...interface{}) (int64, error) {
	sql = "select count(*) from (" + sql + ") t"
	rows, err := db.Query(sql, args...)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	var total int64
	rows.Next()
	rows.Scan(&total)
	return total, nil
}

func QueryPage(db *sql.DB, query string, args []interface{}, dest []interface{}, _pg interface{}) error {
	// page := GetPageContext[T]()
	// 使用反射获取泛型类型
	pgValue := reflect.ValueOf(_pg)
	pgType := pgValue.Type()

	if pgType.Kind() != reflect.Ptr || pgType.Elem().Kind() != reflect.Struct {
		return errors.New("_pg must be a pointer to a Page struct")
	}

	// 获取 Page 结构体的字段
	pageNum := pgValue.Elem().FieldByName("PageNum")
	pageSize := pgValue.Elem().FieldByName("PageSize")
	totalRows := pgValue.Elem().FieldByName("TotalRows")
	list := pgValue.Elem().FieldByName("List")
	sort := pgValue.Elem().FieldByName("Sort")

	// pageNum最小值为1
	if pageNum.Int() < 1 {
		pageNum.SetInt(1)
	}
	// 计算总行数
	total, err := SelectTotal(db, query, args...)
	if err != nil {
		return err
	}
	// 每页行数
	if pageSize.Int() <= 0 {
		pageSize.SetInt(10)
	}
	totalRows.SetInt(total)
	// 计算总页数
	totalPages := (total + pageSize.Int() - 1) / pageSize.Int()
	pgValue.Elem().FieldByName("TotalPages").SetInt(totalPages)
	// 重新拼装sql语句
	sql := fmt.Sprintf("select * from ("+query+") as t order by %s limit ?,?", sort.String())
	log.Println("分页sql:", sql)
	offset := (pageNum.Int() - 1) * pageSize.Int()
	args = append(args, offset, pageSize.Int())
	resultErr := database.QueryStruct(db, sql, args, dest)
	// 拼装到page对象中
	// dest里必定有一个结构体指针
	// 拼装到 page 对象中
	for _, v := range dest {
		destValue := reflect.ValueOf(v)
		if destValue.Kind() == reflect.Ptr && destValue.Elem().Kind() == reflect.Slice {
			list.Set(destValue.Elem())
			break
		}
	}
	return resultErr
}

// func EndPage(){

// }

//func InitPagePlugin() {
//	RegisterHook(HOOK_BEFORE_EXECUTE, func(ctx *HookContext) error {
//		StartPage(new(Page))
//		return nil
//	})
//}

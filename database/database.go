package database

import (
	"database/sql"
	"encoding/binary"
	"errors"
	"fmt"
	mysqld "github.com/llyb120/vodka/database/mysql"
	"reflect"
	"strconv"

	_ "github.com/go-sql-driver/mysql" // 添加这行
)

type MockSlice struct {
	Data *[]interface{}
	Type *reflect.Type
}

// 连接SQLite数据库
func ConnectMySQL(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	// 测试连接
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

// 正常的查询map接口
func QueryMap(db *sql.DB, query string, args ...interface{}) ([]map[string]interface{}, error) {
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var maps []map[string]interface{}

	for rows.Next() {
		values := make([]interface{}, len(columns))
		scanArgs := make([]interface{}, len(columns))
		for i := range values {
			scanArgs[i] = &values[i]
		}

		err := rows.Scan(scanArgs...)
		if err != nil {
			return nil, err
		}

		rowMap := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			if a, ok := val.([]uint8); ok {
				var num uint64
				switch len(a) {
				case 1:
					num = uint64(a[0])
					rowMap[col] = num
				case 2:
					num = uint64(binary.BigEndian.Uint16(a))
					rowMap[col] = num
				case 4:
					num = uint64(binary.BigEndian.Uint32(a))
					rowMap[col] = num
				case 8:
					num = binary.BigEndian.Uint64(a)
					rowMap[col] = num
				default:
					// 处理其他长度的情况
					rowMap[col] = string(a)
				}
			} else if b, ok := val.([]byte); ok {
				rowMap[col] = string(b)
			} else {
				rowMap[col] = val
			}
		}
		maps = append(maps, rowMap)
	}

	return maps, nil
}

func Execute(db *sql.DB, query string, args ...interface{}) (sql.Result, error) {
	return db.Exec(query, args...)
}

func ExecuteInt64(db *sql.DB, query string, args []interface{}, dest []interface{}) error {
	sqlResult, err := db.Exec(query, args...)
	if err != nil {
		return err
	}

	affected, err := sqlResult.RowsAffected()
	if err != nil {
		return err
	}

	lastInsertId, err := sqlResult.LastInsertId()
	if err != nil {
		return err
	}

	//按照影响的行数、最后插入的id以此进行赋值，如果int64的个数超过了2，后续的都忽略
	useAffected := false
	for _, _dest := range dest {
		destValue := reflect.ValueOf(_dest)
		if destValue.Kind() == reflect.Ptr && destValue.Elem().Kind() == reflect.Int64 {
			if !useAffected {
				destValue.Elem().Set(reflect.ValueOf(affected))
				useAffected = true
			} else {
				destValue.Elem().Set(reflect.ValueOf(lastInsertId))
			}
		}
	}

	return nil
}

// dest中目前只允许有以下几种可能：
// 1. 指向切片的指针, 如：&[]User{}, 表示直接查出一个列表，最常用的用法
// 2. 指向结构体的指针, 如：&User{}, 如果列表个数唯一，则直接赋值，如果列表个数大于1，则返回报错
// 3. 因为返回值已经有了error，所以不允许这里再出现error
// todo: 指向map的后续再加
func QueryStruct(db *sql.DB, query string, args []interface{}, dest []interface{}) error {
	// 先查个map出来
	maps, err := QueryMap(db, query, args...)
	if err != nil {
		return err
	}

	for i, _dest := range dest {
		// 特殊处理，go 1.16没有泛型的情况
		if mockSlice, ok := _dest.(*MockSlice); ok {
			// 将map中的数据填充到mockSlice中
			// for _, m := range maps {
			// 	mockSlice.Data = append(mockSlice.Data, m)
			// }
			result := []interface{}{}
			for _, m := range maps {
				// 创建新的结构体实例
				newElemPtr := reflect.New(*mockSlice.Type)
				newElem := newElemPtr.Elem()

				// 遍历结构体的字段
				for i := 0; i < newElem.NumField(); i++ {
					field := newElem.Type().Field(i)
					// 获取字段名，优先使用 vo 标签
					fieldName := field.Name
					if voTag := field.Tag.Get("vo"); voTag != "" {
						fieldName = voTag
					} else if jsonTag := field.Tag.Get("json"); jsonTag != "" {
						fieldName = jsonTag
					}

					// 如果map中存在对应的键，则设置字段值
					if value, ok := m[fieldName]; ok {
						// 特殊处理一下，如果是string类型，是不能设置为nil的
						if value == nil {
							if field.Type.Kind() == reflect.String {
								value = ""
							}
						}
						// 将interface{}转换为字段类型
						fieldValue := reflect.ValueOf(value)
						if fieldValue.Type().ConvertibleTo(field.Type) {
							newElem.Field(i).Set(fieldValue.Convert(field.Type))
						} else {
							// 数字 -> bool
							if field.Type.Kind() == reflect.Bool {
								switch value := value.(type) {
								case int:
									newElem.Field(i).SetBool(value != 0)
								case int8:
									newElem.Field(i).SetBool(value != 0)
								case int16:
									newElem.Field(i).SetBool(value != 0)
								case int32:
									newElem.Field(i).SetBool(value != 0)
								case int64:
									newElem.Field(i).SetBool(value != 0)
								case uint:
									newElem.Field(i).SetBool(value != 0)
								case uint8:
									newElem.Field(i).SetBool(value != 0)
								case uint16:
									newElem.Field(i).SetBool(value != 0)
								case uint32:
									newElem.Field(i).SetBool(value != 0)
								case uint64:
									newElem.Field(i).SetBool(value != 0)
								}
							}
						}
					}
				}

				// 将新的结构体添加到切片中
				result = append(result, newElemPtr.Interface())
			}
			dest[i] = &result
			continue
		}

		destValue := reflect.ValueOf(_dest)
		// 找到指向切片的指针
		if destValue.Kind() == reflect.Ptr && destValue.Elem().Kind() == reflect.Slice {
			// 获取切片的元素类型
			sliceElemType := destValue.Elem().Type().Elem()

			// 创建新的切片
			newSlice := reflect.MakeSlice(reflect.SliceOf(sliceElemType), 0, len(maps))

			// 遍历maps，将每个map转换为结构体并添加到切片中
			for _, m := range maps {
				// 创建新的结构体实例
				newElemPtr := reflect.New(sliceElemType.Elem())
				newElem := newElemPtr.Elem()

				// 遍历结构体的字段
				for i := 0; i < newElem.NumField(); i++ {
					field := newElem.Type().Field(i)
					// 获取字段名，优先使用 vo 标签
					fieldName := field.Name
					if voTag := field.Tag.Get("vo"); voTag != "" {
						fieldName = voTag
					} else if jsonTag := field.Tag.Get("json"); jsonTag != "" {
						fieldName = jsonTag
					}

					// 如果map中存在对应的键，则设置字段值
					if value, ok := m[fieldName]; ok {
						// 特殊处理一下，如果是string类型，是不能设置为nil的
						if value == nil {
							if field.Type.Kind() == reflect.String {
								value = ""
							}
						}
						// 将interface{}转换为字段类型
						fieldValue := reflect.ValueOf(value)
						if fieldValue.Type().ConvertibleTo(field.Type) {
							newElem.Field(i).Set(fieldValue.Convert(field.Type))
						} else {
							// 数字 -> bool
							if field.Type.Kind() == reflect.Bool {
								switch value := value.(type) {
								case int:
									newElem.Field(i).SetBool(value != 0)
								case int8:
									newElem.Field(i).SetBool(value != 0)
								case int16:
									newElem.Field(i).SetBool(value != 0)
								case int32:
									newElem.Field(i).SetBool(value != 0)
								case int64:
									newElem.Field(i).SetBool(value != 0)
								case uint:
									newElem.Field(i).SetBool(value != 0)
								case uint8:
									newElem.Field(i).SetBool(value != 0)
								case uint16:
									newElem.Field(i).SetBool(value != 0)
								case uint32:
									newElem.Field(i).SetBool(value != 0)
								case uint64:
									newElem.Field(i).SetBool(value != 0)
								}
							}
						}
					}
				}

				// 将新的结构体添加到切片中
				newSlice = reflect.Append(newSlice, newElemPtr)
			}

			// 将新的切片赋值给目标变量
			destValue.Elem().Set(newSlice)
			// sliceContainers = append(sliceContainers, i)
		} else if destValue.Kind() == reflect.Ptr && destValue.Elem().Kind() == reflect.Int64 {
			// 如果是int64类型，则必然是count查询，返回一个int64
			if len(maps) == 1 {
				// 取出map的第一个value，转为int64
				m := maps[0]
				var ret int64
				for _, v := range m {
					switch v := v.(type) {
					case int, int8, int16, int32, int64:
						ret = reflect.ValueOf(v).Int()
					case uint, uint8, uint16, uint32, uint64:
						ret = int64(reflect.ValueOf(v).Uint())
					case float32, float64:
						ret = int64(reflect.ValueOf(v).Float())
					case string:
						ret, _ = strconv.ParseInt(v, 10, 64)
					default:
						return errors.New(fmt.Sprintf("无法转换类型 %T 为 int64", v))
					}
					break
				}
				destValue.Elem().Set(reflect.ValueOf(ret))
			} else {
				destValue.Elem().Set(reflect.Zero(destValue.Elem().Type()))
			}
		} else if destValue.Kind() == reflect.Ptr && destValue.Elem().Elem().Kind() == reflect.Struct {
			// 这里必须用指针的指针进行判断
			// 如果maps只有一个，则映射到结构体
			if len(maps) == 0 {
				// 设置为nil
				// dest[0] = nil
				destValue.Elem().Set(reflect.Zero(destValue.Elem().Type()))
			} else {
				m := maps[0]
				// 遍历结构体的字段
				// 指针的指针进行变形
				destValue = destValue.Elem()
				for i := 0; i < destValue.Elem().NumField(); i++ {
					field := destValue.Elem().Type().Field(i)
					// 获取字段名，优先使用 vo 标签
					fieldName := field.Name
					if voTag := field.Tag.Get("vo"); voTag != "" {
						fieldName = voTag
					} else if jsonTag := field.Tag.Get("json"); jsonTag != "" {
						fieldName = jsonTag
					}

					// 如果map中存在对应的键，则设置字段值
					if value, ok := m[fieldName]; ok {
						// 将interface{}转换为字段类型
						fieldValue := reflect.ValueOf(value)
						if fieldValue.Type().ConvertibleTo(field.Type) {
							destValue.Elem().Field(i).Set(fieldValue.Convert(field.Type))
						}
					}
				}
			}
		}
	}

	return nil
	// // 检查 dest 是否为指向切片的指针
	// destValue := reflect.ValueOf(dest[0])
	// if destValue.Kind() != reflect.Ptr && destValue.Kind() != reflect.Slice {
	// 	return errors.New("dest must be a pointer to a slice")
	// }

	// sliceValue := destValue
	// // elementType必定为一个指针，不支持直接使用结构体
	// elementType := sliceValue.Type().Elem()

	// // 执行查询
	// rows, err := DB.Query(query, args...)
	// if err != nil {
	// 	return err
	// }
	// defer rows.Close()

	// for rows.Next() {
	// 	// 创建新的元素
	// 	newElement := reflect.New(elementType).Elem()

	// 	// 准备扫描目标
	// 	scanFields := make([]interface{}, elementType.Elem().NumField())
	// 	for i := 0; i < elementType.NumField(); i++ {
	// 		scanFields[i] = newElement.Field(i).Addr().Interface()
	// 	}

	// 	// 扫描行数据到结构体字段
	// 	if err := rows.Scan(scanFields...); err != nil {
	// 		return err
	// 	}

	// 	// 将新元素添加到切片中
	// 	sliceValue.Set(reflect.Append(sliceValue, newElement))
	// }

	// if err = rows.Err(); err != nil {
	// 	return err
	// }

	// return nil
}

func SetDB(db *sql.DB) {
	// 默认使用mysql
	mysqld.SetDB(db)
}

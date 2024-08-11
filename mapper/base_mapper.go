package mapper

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"strings"
)

type VodkaMapper[T any, ID any] struct {
	InsertOne           func(params *T) (int64, int64, error)             `params:"params"`
	InsertBatch         func(params []*T) (int64, int64, error)           `params:"params"`
	UpdateById          func(params *T) (int64, error)                    `params:"params"`
	UpdateSelectiveById func(params *T) (int64, error)                    `params:"params"`
	DeleteById          func(id ID) (int64, error)                        `params:"id"`
	SelectById          func(id ID) (*T, error)                           `params:"id"`
	SelectAll           func(params *T) ([]*T, error)                     `params:"params"`
	SelectAllByMap      func(params map[string]interface{}) ([]*T, error) `params:"params"`
}

type Tag struct {
	// Params []string
	// Xml    string
	Sql string
}

func (m *VodkaMapper[T, ID]) BuildTags(metadata *MetaData) (map[string]string, error) {
	if metadata == nil {
		return nil, errors.New("没有分析出 _ 字段附加的表信息，无法使用BaseMapper")
	}
	if metadata.TableName == "" {
		return nil, errors.New("没有分析出 _ 字段附加的表信息，无法使用BaseMapper")
	}
	if len(metadata.PKNames) == 0 {
		return nil, errors.New("没有分析出 _ 字段附加的表信息，无法使用BaseMapper")
	}
	// 反射T类型
	baseMapperType := reflect.TypeOf(m).Elem()
	log.Println("Mapper类型", baseMapperType)

	// 获取泛型参数T的类型
	insertOneField, _ := baseMapperType.FieldByName("InsertOne")
	insertOneFuncType := insertOneField.Type

	// 获取InsertOne函数的第一个参数类型，即*T
	tPtrType := insertOneFuncType.In(0)

	// 获取T的类型（去掉指针）
	tType := tPtrType.Elem()
	log.Println("T的类型:", tType)

	// 获取tType中的空字段
	// 获取tType中的_字段
	var fields []reflect.StructField
	var tags []string

	for i := 0; i < tType.NumField(); i++ {
		field := tType.Field(i)

		// if field.Name == "_" {
		// 	// 拿到table的tag
		// 	tableName = field.Tag.Get("table")
		// 	if tableName == "" {
		// 		return nil, errors.New("table tag is required")
		// 	}
		// 	// 拿到pk的tag
		// 	pkTag := field.Tag.Get("pk")
		// 	if pkTag != "" {
		// 		// 支持多个主键，用逗号分隔
		// 		for _, pk := range strings.Split(pkTag, ",") {
		// 			pkNames[strings.TrimSpace(pk)] = 1
		// 		}
		// 	}
		// } else {
		if voTag := field.Tag.Get("vo"); voTag != "" {
			tags = append(tags, voTag)
			fields = append(fields, field)
		}
		// }

	}

	// 拼装sql
	var insertOneBuilder strings.Builder
	insertOneBuilder.WriteString("insert into " + metadata.TableName + " (")
	var insertBatchBuilder strings.Builder
	insertBatchBuilder.WriteString("insert into " + metadata.TableName + " (")
	var updateByIdBuilder strings.Builder
	updateByIdBuilder.WriteString("update " + metadata.TableName + " <set>")
	var updateSelectiveByIdBuilder strings.Builder
	updateSelectiveByIdBuilder.WriteString("update " + metadata.TableName + " <set>")
	var deleteByIdBuilder strings.Builder
	deleteByIdBuilder.WriteString("delete from " + metadata.TableName + " <where> ")
	var selectByIdBuilder strings.Builder
	selectByIdBuilder.WriteString("select * from " + metadata.TableName + " <where> ")
	//var selectAllBuilder strings.Builder
	//var selectAllByMapBuilder strings.Builder

	// 处理字段
	isNumberTypeArr := make([]bool, len(fields))
	isStringTypeArr := make([]bool, len(fields))
	for i := 0; i < len(fields); i++ {
		// fieldType := field.Type
		// fieldTag := field.Tag
		// fieldValue := field.Value
		// 只处理有tag的
		insertOneBuilder.WriteString(tags[i])
		insertBatchBuilder.WriteString(tags[i])
		isNumberType := fields[i].Type.Kind() == reflect.Int || fields[i].Type.Kind() == reflect.Int64 || fields[i].Type.Kind() == reflect.Float64
		isNumberTypeArr[i] = isNumberType
		isStringType := fields[i].Type.Kind() == reflect.String
		isStringTypeArr[i] = isStringType
		// 如果是主键
		if _, ok := metadata.PKNames[tags[i]]; ok {
			// updateByIdBuilder.WriteString(tags[i] + " = #{" + tags[i] + "}")
			selectByIdBuilder.WriteString(" and " + tags[i] + " = #{" + tags[i] + "}")
		} else {
			updateByIdBuilder.WriteString(tags[i] + " = #{" + tags[i] + "}")
			// 处理selective的类型，如果是int int64 float64 这些，不能判断==null
			updateSelectiveByIdBuilder.WriteString(fmt.Sprintf(`<if test="%s != 0 && %s != null && %s != ''">%s = #{%s},</if>`, tags[i], tags[i], tags[i], tags[i], tags[i]))
			// if isNumberType {
			// } else if(is{
			// 	updateSelectiveByIdBuilder.WriteString(fmt.Sprintf(`<if test="%s != null">%s = #{%s},</if>`, tags[i], tags[i], tags[i]))
			// }
			deleteByIdBuilder.WriteString(" and " + tags[i] + " = #{" + tags[i] + "}")
			if i != len(fields)-1 {
				updateByIdBuilder.WriteString(",")
			}
		}
		if i != len(fields)-1 {
			insertOneBuilder.WriteString(",")
			insertBatchBuilder.WriteString(",")
		}
	}
	insertOneBuilder.WriteString(") values (")
	insertBatchBuilder.WriteString(") values <foreach collection='params' item='item' separator=','>(")
	updateByIdBuilder.WriteString("</set> <where>")
	updateSelectiveByIdBuilder.WriteString("</set> <where>")
	// 处理值
	for i := 0; i < len(fields); i++ {
		// 如果是主键
		if _, ok := metadata.PKNames[tags[i]]; ok && (fields[i].Type.Kind() == reflect.Int64) {
			insertOneBuilder.WriteString("#{" + tags[i] + " == 0 ? $AUTO : " + tags[i] + "}")
			insertBatchBuilder.WriteString("#{item." + tags[i] + " == 0 ? $AUTO : item." + tags[i] + "}")
			updateByIdBuilder.WriteString(" and " + tags[i] + " = #{" + tags[i] + "}")
			updateSelectiveByIdBuilder.WriteString(fmt.Sprintf(" and %s = #{%s}", tags[i], tags[i]))
		} else {
			insertOneBuilder.WriteString("#{" + tags[i] + "}")
			insertBatchBuilder.WriteString("#{item." + tags[i] + "}")
		}
		if i != len(fields)-1 {
			insertOneBuilder.WriteString(",")
			insertBatchBuilder.WriteString(",")
		}
	}
	insertOneBuilder.WriteString(")")
	insertBatchBuilder.WriteString(")</foreach>")
	updateByIdBuilder.WriteString("</where>")
	updateSelectiveByIdBuilder.WriteString("</where>")
	deleteByIdBuilder.WriteString("</where>")
	selectByIdBuilder.WriteString("</where>")

	resultMap := make(map[string]string)
	resultMap["InsertOne"] = insertOneBuilder.String()
	resultMap["InsertBatch"] = insertBatchBuilder.String()
	resultMap["UpdateById"] = updateByIdBuilder.String()
	resultMap["UpdateSelectiveById"] = updateSelectiveByIdBuilder.String()
	resultMap["DeleteById"] = deleteByIdBuilder.String()

	return resultMap, nil
}

type User struct {
	VodkaMapper[User, int64]
}

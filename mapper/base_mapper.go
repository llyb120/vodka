package mapper

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"strings"
	"vodka/analyzer"
)

type VodkaMapper struct {
	InsertOne            func(params interface{}) (int64, int64, error)                                                      `params:"params"`
	InsertBatch          func(params interface{}) (int64, int64, error)                                                      `params:"params"`
	UpdateById           func(params interface{}) (int64, error)                                                             `params:"params"`
	UpdateSelectiveById  func(params interface{}) (int64, error)                                                             `params:"params"`
	UpdateByCondition    func(condition interface{}, action interface{}) (int64, error)                                      `params:"condition,action"`
	UpdateByConditionMap func(condition map[string]interface{}, action map[string]interface{}) (int64, error)                `params:"condition,action"`
	DeleteById           func(id interface{}) (int64, error)                                                                 `params:"id"`
	SelectById           func(id interface{}) (interface{}, error)                                                           `params:"id"`
	SelectAll            func(params interface{}, order string, offset int64, limit int64) ([]interface{}, error)            `params:"...params,order,offset,limit"`
	CountAll             func(params interface{}) (int64, error)                                                             `params:"params"`
	SelectAllByMap       func(params map[string]interface{}, order string, offset int64, limit int64) ([]interface{}, error) `params:"...params,order,offset,limit"`
	CountAllByMap        func(params map[string]interface{}) (int64, error)                                                  `params:"params"`
}

type Tag struct {
	// Params []string
	// Xml    string
	Sql string
}

func (m *VodkaMapper) BuildTags(metadata *MetaData) ([]*analyzer.Function, error) {
	if metadata == nil {
		return nil, errors.New("没有分析出 _ 字段附加的表信息，无法使用BaseMapper")
	}
	if metadata.TableName == "" {
		return nil, errors.New("没有分析出 _ 字段附加的表信息，无法使用BaseMapper")
	}
	if len(metadata.PKNames) == 0 {
		return nil, errors.New("没有分析出 _ 字段附加的表信息，无法使用BaseMapper")
	}
	if metadata.ModelType == nil || metadata.PkType == nil {
		return nil, errors.New("没有分析出 _model 和 _pk 字段附加的表信息，无法使用BaseMapper")
	}
	// go 1.16不支持泛型，所以必须定义 _model 和 _pk 类型
	// 反射T类型
	baseMapperType := reflect.TypeOf(m).Elem()
	log.Println("Mapper类型", baseMapperType)

	// 获取泛型参数T的类型
	// insertOneField, _ := baseMapperType.FieldByName("InsertOne")
	// insertOneFuncType := insertOneField.Type

	// 获取InsertOne函数的第一个参数类型，即*T
	// tPtrType := insertOneFuncType.In(0)

	// 获取T的类型（去掉指针）
	tType := *metadata.ModelType
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
	insertOneBuilder.WriteString("<insert id=\"InsertOne\">insert into " + metadata.TableName + " (")
	var insertBatchBuilder strings.Builder
	insertBatchBuilder.WriteString("<insert id=\"InsertBatch\">insert into " + metadata.TableName + " (")
	var updateByIdBuilder strings.Builder
	updateByIdBuilder.WriteString("<update id=\"UpdateById\">update " + metadata.TableName + " <set>")
	var updateSelectiveByIdBuilder strings.Builder
	updateSelectiveByIdBuilder.WriteString("<update id=\"UpdateSelectiveById\">update " + metadata.TableName + " <set>")
	var deleteByIdBuilder strings.Builder
	deleteByIdBuilder.WriteString("<delete id=\"DeleteById\">delete from " + metadata.TableName + " <where> ")
	var selectByIdBuilder strings.Builder
	selectByIdBuilder.WriteString("<select id=\"SelectById\">select * from " + metadata.TableName + " <where> ")
	var selectAllBuilder strings.Builder
	var selectAllWhereBuilder strings.Builder
	selectAllBuilder.WriteString("<select id=\"SelectAll\">select * from " + metadata.TableName + " <where> ")
	var selectAllByMapBuilder strings.Builder
	selectAllByMapBuilder.WriteString("<select id=\"SelectAllByMap\">select * from " + metadata.TableName + " <where> ")
	var selectAllByMapWhereBuilder strings.Builder
	// update的condition
	var updateByConditionBuilder strings.Builder
	updateByConditionBuilder.WriteString(`<update id="UpdateByCondition">update ` + metadata.TableName + ` <set> `)
	var updateByConditionMapBuilder strings.Builder
	updateByConditionMapBuilder.WriteString(`<update id="UpdateByConditionMap">update ` + metadata.TableName + ` <set> `)
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
			updateByIdBuilder.WriteString(tags[i] + " = #{" + tags[i] + "},")
			// 处理selective的类型，如果是int int64 float64 这些，不能判断==null
			updateSelectiveByIdBuilder.WriteString(fmt.Sprintf(`<if test="%s != 0 && %s != null && %s != ''">%s = #{%s},</if>`, tags[i], tags[i], tags[i], tags[i], tags[i]))
			updateSetStatement := fmt.Sprintf(`<if test="action.%s != 0 && action.%s != null && action.%s != ''">%s = #{action.%s},</if>`, tags[i], tags[i], tags[i], tags[i], tags[i])
			updateByConditionBuilder.WriteString(updateSetStatement)
			updateByConditionMapBuilder.WriteString(updateSetStatement)
			// if isNumberType {
			// } else if(is{
			// 	updateSelectiveByIdBuilder.WriteString(fmt.Sprintf(`<if test="%s != null">%s = #{%s},</if>`, tags[i], tags[i], tags[i]))
			// }
			deleteByIdBuilder.WriteString(" and " + tags[i] + " = #{" + tags[i] + "}")
			// if i != len(fields)-1 {
			// 	updateByIdBuilder.WriteString(",")
			// }
		}
		// 查询条件
		selectAllWhereBuilder.WriteString(fmt.Sprintf(` <if test="%s != null && %s != '' && %s != 0"> and %s = #{%s} </if>`, tags[i], tags[i], tags[i], tags[i], tags[i]))
		// 针对map的查询条件
		buildMapCondition(&selectAllByMapWhereBuilder, tags[i], tags[i])
		// selectAllByMapWhereBuilder.WriteString(fmt.Sprintf(` <if test="EQ_%s != null && EQ_%s != '' && EQ_%s != 0"> and %s = #{%s} </if>`, tags[i], tags[i], tags[i], tags[i], tags[i]))
		// selectAllByMapWhereBuilder.WriteString(fmt.Sprintf(` <if test="GT_%s != null && GT_%s != '' && GT_%s != 0"> and %s > #{%s} </if>`, tags[i], tags[i], tags[i], tags[i], tags[i]))
		// selectAllByMapWhereBuilder.WriteString(fmt.Sprintf(` <if test="LT_%s != null && LT_%s != '' && LT_%s != 0"> and %s < #{%s} </if>`, tags[i], tags[i], tags[i], tags[i], tags[i]))
		// selectAllByMapWhereBuilder.WriteString(fmt.Sprintf(` <if test="GTE_%s != null && GTE_%s != '' && GTE_%s != 0"> and %s >= #{%s} </if>`, tags[i], tags[i], tags[i], tags[i], tags[i]))
		// selectAllByMapWhereBuilder.WriteString(fmt.Sprintf(` <if test="LTE_%s != null && LTE_%s != '' && LTE_%s != 0"> and %s <= #{%s} </if>`, tags[i], tags[i], tags[i], tags[i], tags[i]))
		// selectAllByMapWhereBuilder.WriteString(fmt.Sprintf(` <if test="LIKE_%s != null && LIKE_%s != '' && LIKE_%s != 0"> and %s like concat('%%',#{%s},'%%') </if>`, tags[i], tags[i], tags[i], tags[i], tags[i]))
		// selectAllByMapWhereBuilder.WriteString(fmt.Sprintf(` <if test="IN_%s != null && IN_%s != '' && IN_%s != 0"> and %s in <foreach collection='%s' item='item' separator=',' open='(' close=')'>#{item}</foreach> </if>`, tags[i], tags[i], tags[i], tags[i], tags[i]))
		if i != len(fields)-1 {
			insertOneBuilder.WriteString(",")
			insertBatchBuilder.WriteString(",")
		}
	}
	insertOneBuilder.WriteString(") values (")
	insertBatchBuilder.WriteString(") values <foreach collection='params' item='item' separator=','>(")
	updateByIdBuilder.WriteString("</set> <where>")
	updateSelectiveByIdBuilder.WriteString("</set> <where>")
	updateByConditionBuilder.WriteString("</set> <where>")
	updateByConditionMapBuilder.WriteString("</set> <where>")
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
		// 更新语句
		updateByConditionBuilder.WriteString(fmt.Sprintf(`<if test="condition.%s != 0 && condition.%s != null && condition.%s != ''"> and %s = #{condition.%s}</if>`, tags[i], tags[i], tags[i], tags[i], tags[i]))
		buildMapCondition(&updateByConditionMapBuilder, "condition."+tags[i], tags[i])
		if i != len(fields)-1 {
			insertOneBuilder.WriteString(",")
			insertBatchBuilder.WriteString(",")
		}
	}
	insertOneBuilder.WriteString(")</insert>")
	insertBatchBuilder.WriteString(")</foreach></insert>")
	updateByIdBuilder.WriteString("</where></update>")
	updateSelectiveByIdBuilder.WriteString("</where></update>")
	deleteByIdBuilder.WriteString("</where></delete>")
	selectByIdBuilder.WriteString("</where></select>")
	selectAllBuilder.WriteString(selectAllWhereBuilder.String())
	selectAllBuilder.WriteString(`</where> <if test="order != ''"> order by ${order} </if> limit #{offset},#{limit}</select>`)
	selectAllByMapBuilder.WriteString(selectAllByMapWhereBuilder.String())
	selectAllByMapBuilder.WriteString(`</where> <if test="order != ''"> order by ${order} </if> limit #{offset},#{limit}</select>`)
	updateByConditionBuilder.WriteString("</where></update>")
	updateByConditionMapBuilder.WriteString("</where></update>")

	// 针对map类参数的处理
	var builder strings.Builder
	builder.WriteString("<mapper>")
	builder.WriteString(insertOneBuilder.String())
	builder.WriteString(insertBatchBuilder.String())
	builder.WriteString(updateByIdBuilder.String())
	builder.WriteString(updateSelectiveByIdBuilder.String())
	builder.WriteString(deleteByIdBuilder.String())
	builder.WriteString(selectByIdBuilder.String())
	builder.WriteString(selectAllBuilder.String())
	builder.WriteString(fmt.Sprintf(`<select id="CountAll">select count(*) from %s <where> %s </where></select>`, metadata.TableName, selectAllWhereBuilder.String()))
	builder.WriteString(selectAllByMapBuilder.String())
	builder.WriteString(fmt.Sprintf(`<select id="CountAllByMap">select count(*) from %s <where> %s </where></select>`, metadata.TableName, selectAllByMapWhereBuilder.String()))
	builder.WriteString(updateByConditionBuilder.String())
	builder.WriteString(updateByConditionMapBuilder.String())
	builder.WriteString("</mapper>")

	return analyzer.ParseXml(metadata.Namespace, builder.String())
	// return resultMap, nil
}

// 构造conditionMap
func buildMapCondition(builder *strings.Builder, action, condition string) {
	builder.WriteString(fmt.Sprintf(` <if test="EQ_%s != null && EQ_%s != '' && EQ_%s != 0"> and %s = #{EQ_%s} </if>`, condition, condition, condition, condition, condition))
	builder.WriteString(fmt.Sprintf(` <if test="NE_%s != null && NE_%s != '' && NE_%s != 0"> and %s <> #{NE_%s} </if>`, condition, condition, condition, condition, condition))
	builder.WriteString(fmt.Sprintf(` <if test="GT_%s != null && GT_%s != '' && GT_%s != 0"> and %s > #{GT_%s} </if>`, condition, condition, condition, condition, condition))
	builder.WriteString(fmt.Sprintf(` <if test="LT_%s != null && LT_%s != '' && LT_%s != 0"> and %s < #{LT_%s} </if>`, condition, condition, condition, condition, condition))
	builder.WriteString(fmt.Sprintf(` <if test="GTE_%s != null && GTE_%s != '' && GTE_%s != 0"> and %s >= #{GTE_%s} </if>`, condition, condition, condition, condition, condition))
	builder.WriteString(fmt.Sprintf(` <if test="LTE_%s != null && LTE_%s != '' && LTE_%s != 0"> and %s <= #{LTE_%s} </if>`, condition, condition, condition, condition, condition))
	builder.WriteString(fmt.Sprintf(` <if test="LIKE_%s != null && LIKE_%s != '' && LIKE_%s != 0"> and %s like concat('%%',#{LIKE_%s},'%%') </if>`, condition, condition, condition, condition, condition))
	builder.WriteString(fmt.Sprintf(` <if test="IN_%s != null && IN_%s != '' && IN_%s != 0"> and %s in <foreach collection='IN_%s' item='item' separator=',' open='(' close=')'>#{item}</foreach> </if>`, condition, condition, condition, condition, condition))
	builder.WriteString(fmt.Sprintf(` <if test="NOT_IN_%s != null && NOT_IN_%s != '' && NOT_IN_%s != 0"> and %s not in <foreach collection='NOT_IN_%s' item='item' separator=',' open='(' close=')'>#{item}</foreach> </if>`, condition, condition, condition, condition, condition))
	// 暂时不要between
	// builder.WriteString(fmt.Sprintf(` <if test="BETWEEN_%s != null && BETWEEN_%s != '' && BETWEEN_%s != 0"> and %s between #{%s} and #{%s} </if>`, condition, condition, condition, condition, condition, condition))
	// builder.WriteString(fmt.Sprintf(` <if test="NOT_BETWEEN_%s != null && NOT_BETWEEN_%s != '' && NOT_BETWEEN_%s != 0"> and %s not between #{%s} and #{%s} </if>`, condition, condition, condition, condition, condition, condition))
}

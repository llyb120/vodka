package mapper

import (
	"reflect"
	"strings"
)

type MetaData struct {
	TableName    string
	PKNames      map[string]byte
	CustomSqlMap map[string]string
	// Fields    []reflect.StructField
	// Tags      []string
}

func NewMetaData(mapperValue reflect.Value) *MetaData {
	metadataField, ok := mapperValue.Elem().Type().FieldByName("_")
	if !ok {
		return nil
	}
	metadata := &MetaData{
		TableName:    "",
		PKNames:      make(map[string]byte),
		CustomSqlMap: make(map[string]string),
	}
	tableName := metadataField.Tag.Get("table")
	if tableName == "" {
		return metadata
	}
	metadata.TableName = tableName
	pkTag := metadataField.Tag.Get("pk")
	if pkTag != "" {
		// 支持多个主键，用逗号分隔
		for _, pk := range strings.Split(pkTag, ",") {
			metadata.PKNames[strings.TrimSpace(pk)] = 1
		}
	}

	if method := mapperValue.MethodByName("BuildTags"); method.IsValid() {
		// refErr := reflect.Zero(reflect.TypeOf(errors.New("")))
		ret := method.Call([]reflect.Value{
			reflect.ValueOf(metadata),
			//mapperValue,
			// reflect.ValueOf(customSqlMap),
			// refErr,
		})
		err, ok := ret[1].Interface().(error)
		if !(ok && err != nil) {
			if customSqlMap, ok := ret[0].Interface().(map[string]string); ok {
				metadata.CustomSqlMap = customSqlMap
			}
		}
	}
	return metadata
}

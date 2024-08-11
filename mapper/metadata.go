package mapper

import (
	"reflect"
	"strings"
	"vodka/analyzer"
)

type MetaData struct {
	Namespace    string
	TableName    string
	PKNames      map[string]byte
	Functions    []*analyzer.Function
	// CustomSqlMap map[string]string
	// Fields    []reflect.StructField
	// Tags      []string
}

func NewMetaData(namespace string, mapperValue reflect.Value) *MetaData {
	metadataField, ok := mapperValue.Elem().Type().FieldByName("_")
	if !ok {
		return nil
	}
	metadata := &MetaData{
		Namespace:    namespace,
		TableName:    "",
		PKNames:      make(map[string]byte),
		Functions:    make([]*analyzer.Function, 0),
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
			if functions, ok := ret[0].Interface().([]*analyzer.Function); ok {
				metadata.Functions = functions
			}
		}
	}
	return metadata
}

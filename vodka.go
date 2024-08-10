package vodka

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"vodka/analyzer"
	"vodka/mapper"
)

func ScanMapper(dir string) error {
	var wg sync.WaitGroup
	var analyzers []*analyzer.Analyzer
	rwMutex := sync.RWMutex{}

	// 遍历指定目录
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 检查文件是否为XML
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".xml") {
			// 读取XML文件内容
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			wg.Add(1)
			go func(path string) {
				// 输出找到的XML文件路径
				println("找到XML文件:", path)
				defer wg.Done()
				parser := analyzer.NewAnalyzer(string(content))
				parser.Parse()
				rwMutex.Lock()
				defer rwMutex.Unlock()
				analyzers = append(analyzers, parser)
			}(path)
		}
		return nil
	})
	wg.Wait()
	// 关闭analyzersChan，释放
	if err != nil {
		return err
	}
	// 整理所有的analyzer，将相同命名空间的mapper集合到一起
	return mapper.InitMappers(analyzers)
}

func InitMapper(source interface{}) error {
	return mapper.BindMapper(source)
}


// func SetDB(_db *sql.DB) {
// 	database.SetDB(_db)
// }

// // 查询并转换为指定结构体
// func QueryStruct(results []interface{}, query string, args ...interface{}) error {
// 	return database.QueryStruct(query, args, results)
// }

// // 查询并转换为map
// func QueryMap(query string, args ...interface{}) ([]map[string]interface{}, error) {
// 	return database.QueryMap(query, args)
// }

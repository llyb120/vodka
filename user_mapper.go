package vodka

import (
	"runtime"
)

type User struct {
	Id   int
	Name string
}

// 定义一个结构体来匹配XML的结构
type UserMapper struct {
	GetUserById   func(id int) *User
	GetUserByName func(name string) *User
	GetUsers      func() []*User
}

func (Mapper *UserMapper) GetCurrentFunctionName() string {
	pc, _, _, _ := runtime.Caller(1)
	return runtime.FuncForPC(pc).Name()
}

func getCurrentFileName() string {
	_, file, _, _ := runtime.Caller(1)
	return file
}

func main() {
	//fileName := getCurrentFileName()
	//fmt.Println(fileName)
	//// 查找当前目录对应的mapper
	//if !strings.HasSuffix(fileName, ".go") {
	//	return
	//}
	//fileName = fileName[strings.LastIndex(fileName, "/")+1 : len(fileName)-3]
	//xmlFileName := fileName + ".xml"
	//_, err := os.Stat(xmlFileName)
	//if os.IsNotExist(err) {
	//	// fmt.Printf("XML文件 %s 不存在\n", xmlFileName)
	//	return
	//}
	//
	//// 读取XML文件
	//xmlContent, err := os.ReadFile(xmlFileName)
	//if err != nil {
	//	// fmt.Printf("读取XML文件失败: %v\n", err)
	//	return
	//}
	//analyzer.NewAnalyzer(string(xmlContent))
	//
	//mapper := new(UserMapper)
	//fmt.Println("当前函数名:", mapper.GetCurrentFunctionName())
	//
	//fmt.Println(fileName)
	//fmt.Println(getCurrentFileName())
}

package main

type TaskParamData struct {
	Name       string
	Type       string
	Range      []string
	IntRange   *[]int
	FloatRange *[]float64
	Dimensions *[]int
	TotalCount int
}

type Task struct {
	Id      int
	Subject int
	Work    int
	Variant int
	Number  int
	Name    string
	Desc    string
	Input   []TaskParamData
	Output  string
	Path    string
}

type Token struct {
	Token   string
	UserId  int
	Subject int
	Work    int
	Variant int
}

type Solution struct {
	Task      int
	Source    string
	Path      string
	TestCases string
	Token     *Token
}

type UserData struct {
	Token       string
	Name        string
	Group       string
	Teacher     string
	GenerateDoc bool
}

type Config struct {
	IsHTTP       bool
	DBName       string
	EntryPoint   string
	Host         string
	Port         int
	CertFilename string
	KeyFilename  string
	Verbose      bool
}

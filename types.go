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

type Subject struct {
	Id   int
	Name string
}

type Work struct {
	Id     int
	NextId *int
	Name   string
}

type Task struct {
	Id       int
	Subject  int
	Work     int
	Variant  int
	Number   int
	Name     string
	Desc     string
	Input    []TaskParamData
	Output   string
	Path     string
	IsPassed bool
}

type Token struct {
	Id      int
	UserId  int
	Subject int
	Variant int
}

type Solution struct {
	Task      *Task
	Source    string
	Path      string
	TestCases string
	Token     *Token
}

type UserData struct {
	Token string
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

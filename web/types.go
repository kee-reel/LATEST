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
	Id         int
	Name       string
	FolderName string
}

type Work struct {
	Id         int
	NextId     *int
	Name       string
	FolderName string
}

type Task struct {
	Id         int
	Subject    *Subject
	Work       *Work
	Position   int
	FolderName string
	Extention  string
	Name       string
	Desc       string
	Input      []TaskParamData
	Output     string
	Path       string
	IsPassed   bool
}

type Token struct {
	Id      int
	UserId  int
	Subject int
}

type Solution struct {
	Task                 *Task
	Source               string
	Path                 string
	ExecFilename         string
	CompleteExecFilename string
	TestCases            string
	Token                *Token
	IsVerbose            bool
}

type UserData struct {
	Token string
}

type ConfigDB struct {
	Host string
	Port string
	User string
	Pass string
	Name string
}

type ConfigServer struct {
	IsHTTP       bool
	EntryPoint   string
	Host         string
	Port         int
	CertFilename string
	KeyFilename  string
}

type Config struct {
	DB      ConfigDB
	Server  ConfigServer
	Verbose bool
}

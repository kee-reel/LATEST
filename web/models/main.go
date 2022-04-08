package models

type Project struct {
	Id         int    `json:"id" example:"666"`
	Name       string `json:"name" example:"Sample project"`
	FolderName string `json:"folder_name" example:"sample_project"`
}

type Unit struct {
	Id         int    `json:"id" example:"333"`
	ProjectId  int    `json:"id" example:"666"`
	Name       string `json:"name" example:"Sample unit"`
	FolderName string `json:"folder_name" example:"unit-1"`
}

type TaskParamData struct {
	Name       string     `example:"task-1"`
	Type       string     `example:"int"`
	Dimensions []int      `example:"5,4"`
	Range      []string   `example:"-1000,1000"`
	IntRange   *[]int     `json:"-"`
	FloatRange *[]float64 `json:"-"`
	TotalCount int        `json:"-"`
}

type Task struct {
	Id          int             `json:"id" example:"111"`
	Number      int             `json:"number" example:"0"`
	UnitId      int             `json:"unit_id" example:"333"`
	ProjectId   int             `json:"project_id" example:"666"`
	Name        string          `json:"name" example:"Sample task"`
	Desc        string          `json:"desc" example:"Sample description"`
	FolderName  string          `json:"folder_name" example:"task-1"`
	Input       []TaskParamData `json:"input"`
	Output      string          `json:"output" example:"Program must output sum of two integers on a newline"`
	Score       int             `json:"score" json:"15"`
	IsCompleted bool            `json:"is_passed" example:"true"`
	Project     *Project        `json:"-"`
	Unit        *Unit           `json:"-"`
	Extention   string          `json:"-"`
	Path        string          `json:"-"`
}

type Token struct {
	Token  string `json:"token"`
	Id     int    `json:"-"`
	UserId int    `json:"-"`
	IP     string `json:"-"`
}

type User struct {
	Id    int    `json:"-"`
	Email string `json:"-"`
	Name  string `json:"name"`
}

type Solution struct {
	Task                 *Task
	Source               string
	Path                 string
	Extention            string
	ExecFilename         string
	CompleteExecFilename string
	TestCases            string
	Token                *Token
	IsVerbose            bool
}

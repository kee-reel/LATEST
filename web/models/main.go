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
	Name       string     `json:"name" example:"task-1"`
	Type       string     `json:"type" example:"int"`
	Dimensions []int      `json:"dimensions" example:"5,4"`
	Range      []string   `json:"range" example:"-1000,1000"`
	IntRange   *[]int     `json:"-"`
	FloatRange *[]float64 `json:"-"`
	TotalCount int        `json:"-"`
}

type Task struct {
	Id         int             `json:"id" example:"111"`
	Number     int             `json:"number" example:"0"`
	UnitId     int             `json:"unit_id" example:"333"`
	ProjectId  int             `json:"project_id" example:"666"`
	Name       string          `json:"name" example:"Sample task"`
	Desc       string          `json:"desc" example:"Sample description"`
	FolderName string          `json:"folder_name" example:"task-1"`
	Input      []TaskParamData `json:"input"`
	Output     string          `json:"output,omitempty" example:"Program must output sum of two integers on a newline"`
	Score      int             `json:"score" json:"15"`
	Completion float32         `json:"completion" example:"0.58"`
	Project    *Project        `json:"-"`
	Unit       *Unit           `json:"-"`
	Extention  string          `json:"-"`
	Path       string          `json:"-"`
}

type Token struct {
	Token  string
	Email  string
	IP     string
	UserId int
}

type User struct {
	Id    int     `json:"-"`
	Email string  `json:"email"`
	Name  string  `json:"name"`
	Score float32 `json:"score"`
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
	UserId               int
	IsVerbose            bool
}

type Leaderboard map[string]float32

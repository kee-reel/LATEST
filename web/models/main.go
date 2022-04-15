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
	Token  string `json:"token" example:"9rzNUDp8bP6VOnGIqOO011f5EB4jk0eN0osZt0KFZHTtWIpiwqzVj2vof5sOq80QIJbne5dHiH5vEUe7uJ42X5X39tHGpt0LTreFOjMkfdn4sB6gzouUHc4tGubhikoKuK05P06W1x0QK0zJzbPaZYG4mfBpfU1u8xbqSPVo8ZI9zumiJUiHC8MbJxMPYsGJjZMChQBtA0NvKuAReS3v1704QBX5zZCAyyNP47VZ51E9MMqVGoZBxFmJ4mCHRBy7"`
	Id     int    `json:"-"`
	UserId int    `json:"-"`
	IP     string `json:"-"`
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
	IsVerbose            bool
}

type Leaderboard map[string]float32

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
	LanguageId int             `json:"-"`
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
	Id                   int64
	Task                 *Task
	Source               string
	Path                 string
	LanguageId           int
	ExecFilename         string
	CompleteExecFilename string
	Token                *Token
	UserId               int
}

type Leaderboard map[string]float32

type SolutionBuildError struct {
	Msg string `json:"msg,omitempty" example:"Build fail message"`
}
type SolutionTimeoutError struct {
	Params string  `json:"params,omitempty" example:"2;1;7'"`
	Time   float32 `json:"time,omitempty" example:"1.5"`
}
type SolutionRuntimeError struct {
	Params string `json:"params,omitempty" example:"2;1;7'"`
	Msg    string `json:"msg,omitempty" example:"Build fail message"`
}
type SolutionTestError struct {
	Params   string `json:"params,omitempty" example:"2;1;7'"`
	Expected string `json:"expected,omitempty" example:"8"`
	Result   string `json:"result,omitempty" example:"1"`
}

type SolutionErrorData struct {
	Build       *SolutionBuildError   `json:"build,omitempty"`
	Timeout     *SolutionTimeoutError `json:"timeout,omitempty"`
	Runtime     *SolutionRuntimeError `json:"runtime,omitempty"`
	Test        *SolutionTestError    `json:"test,omitempty"`
	TestsPassed int                   `json:"tests_passed" example:"7"`
	TestsTotal  int                   `json:"tests_total" example:"10"`
}

type TestResult struct {
	Id            int64              `json:"id,omitempty"`
	ErrorData     *SolutionErrorData `json:"error_data,omitempty"`
	InternalError *string            `json:"internal_error,omitempty"`
}

type SolutionData struct {
	Text      string `json:"text"`
	Extention string `json:"extention"`
}

type SolutionTests struct {
	Fixed  string `json:"fixed,omitempty"`
	Random string `json:"random,omitempty"`
}

type RunnerData struct {
	Id               int64         `json:"id,omitempty"`
	UserSolution     SolutionData  `json:"user_solution"`
	CompleteSolution SolutionData  `json:"complete_solution"`
	Tests            SolutionTests `json:"tests"`
}

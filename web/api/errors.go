package api

// @Description Empty - success response
type APINoError struct {
}

// @Description Error code from https://github.com/kee-reel/LATE/blob/main/web/errors.go
type APIError struct {
    Error WebError `json:"error" example:"300"`
}

// @Description Error code from https://github.com/kee-reel/LATE/blob/main/web/errors.go
type APIInternalError struct {
	Error WebError `json:"error" example:"1"`
}

type WebError int

const (
	NoError            WebError = 0 // Everything is OK
	Internal                    = 1 // Internal server error, must be a bug
	MethodNotSupported          = 2 // Called unsupported web method

	EmailNotProvided = 100 // "email" field not provided
	EmailInvalid     = 101 // "email" field is invalid, incorrect email address format
	EmailUnknown     = 102 // Provided "email" is not registered
	EmailTaken       = 103 // Provided "email" is already taken

	PasswordNotProvided = 200 // "pass" field not provided
	PasswordInvalid     = 201 // "pass" field is invalid, must be at least 6 symbols
	PasswordWrong       = 202 // "pass" doesn't match

	TokenNotProvided    = 300 // "token" field not provided
	TokenInvalid        = 301 // "token" is invalid, len isn't 256 symbols
	TokenUnknown        = 302 // "token" like this doesn't exist
	TokenNotVerified    = 303 // "token" is not verified, it must be verified by link sent on email
	TokenBoundToOtherIP = 304 // "token" is bound to other IP, need to get new token for this IP

	TaskIdNotProvided = 400 // "task_id" field not provided
	TaskIdInvalid     = 401 // "task_id" field is invalid, must be a number
	TaskNotFound      = 402 // "task_id" doesn't match to any task

	SolutionTextNotProvided = 500 // Neither "source_text" nor "source_file" fields provided
	SolutionTextTooLong     = 501 // Solution text is too long, more than 50000 symbols
	SolutionTestsTooLong    = 502 // Tests text is too long, more than 50000 symbols
	SolutionTestsInvalid    = 503 // Tests doesn't match required format
	SolutionBuildFail       = 504 // Solution compiltion failed, error_data field will be like `{"msg":"Compilation error"}`
	SolutionTestFail        = 505 // Test failed - expected result doesn't match with actual result, error_data field will be like `{"params":"2;1;3;", "expected":"4", "result":"3"}`
	SolutionTimeoutFail     = 506 // Reached timeout while executing solution, error_data field will be like `{"params":"2;1;3;", result:"Last outputed string"}`
	SolutionRuntimeFail     = 507 // Runtime error while solution execution, error_data field will be like `{"params":"2;1;3;", "msg:"Runtime error"}`

	LanguageNotProvided  = 600 // "lang" field not provided
	LanguageNotSupported = 601 // Provided "lang" is not supported

	NameNotProvided = 700 // "name" field not provided
	NameInvalid     = 701 // "name" field is invalid, must be less than 128 symbols

	TasksFoldersInvalid        = 800 // "task_folders" contains more than 3 items
	TasksProjectFolderNotFound = 801 // Project specified in "task_folders" not found, when
	TasksUnitFolderNotFound    = 802 // Unit specified in "task_folders" not found
	TasksTaskFolderNotFound    = 803 // Task specified in "task_folders" not found

	LeaderboardProjectIdNotProvided     = 900
	LeaderboardProjectFolderNotProvided = 901
)

package main

type WebError int

const (
	NoError            WebError = 0 // Everything is OK
	Internal                    = 1 // Internal server error, must be bug
	MethodNotSupported          = 2 // Called unsupported web method

	EmailNotProvided = 100 // "email" field not provided
	EmailInvalid     = 101 // "email" field is invalid, incorrect email address format

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

	SolutionTextNotProvided       = 500 // Neither "source_text" nor "source_file" fields provided
	SolutionTextTooLong           = 501 // Solution text is too long, more than 50000 symbols
	SolutionTestsTooLong          = 502 // Tests text is too long, more than 50000 symbols
	SolutionTestsInvalid          = 503 // Tests doesn't match required format
	SolutionTaskFoldersInvalid    = 504 // "task_folders" contains more than 3 items
	SolutionProjectFolderNotFound = 505 // Project specified in "task_folders" not found, when
	SolutionUnitFolderNotFound    = 506 // Unit specified in "task_folders" not found
	SolutionTaskFolderNotFound    = 507 // Task specified in "task_folders" not found
	SolutionBuildFail             = 508 // Fail during solution building, error message could be found in "error_data"
	SolutionTestFail              = 509 // Fail during solution testing, error message could be found in "error_data"

	LanguageNotProvided  = 600 // "lang" field not provided
	LanguageNotSupported = 601 // Provided "lang" is not supported
)

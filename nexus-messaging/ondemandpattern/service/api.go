package service

// Language represents a spoken language for greetings.
type Language int

const (
	Arabic     Language = iota
	Chinese    Language = iota
	English    Language = iota
	French     Language = iota
	Hindi      Language = iota
	Portuguese Language = iota
	Spanish    Language = iota
)

func (l Language) String() string {
	switch l {
	case Arabic:
		return "Arabic"
	case Chinese:
		return "Chinese"
	case English:
		return "English"
	case French:
		return "French"
	case Hindi:
		return "Hindi"
	case Portuguese:
		return "Portuguese"
	case Spanish:
		return "Spanish"
	}
	return "Unknown"
}

const ServiceName = "NexusRemoteGreetingService"

const (
	RunFromRemoteOperationName = "runFromRemote"
	GetLanguagesOperationName  = "getLanguages"
	GetLanguageOperationName   = "getLanguage"
	SetLanguageOperationName   = "setLanguage"
	ApproveOperationName       = "approve"
)

type RunFromRemoteInput struct {
	UserID string
}

type GetLanguagesInput struct {
	IncludeUnsupported bool
	UserID             string
}

type GetLanguagesOutput struct {
	Languages []Language
}

type GetLanguageInput struct {
	UserID string
}

type SetLanguageInput struct {
	Language Language
	UserID   string
}

type ApproveInput struct {
	Name   string
	UserID string
}

type ApproveOutput struct{}

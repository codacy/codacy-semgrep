package docgen

type Category string

const (
	Security      Category = "Security"
	Performance   Category = "Performance"
	Compatibility Category = "Compatibility"
	ErrorProne    Category = "ErrorProne"
	BestPractice  Category = "BestPractice"
	CodeStyle     Category = "CodeStyle"
)

type Level string

const (
	Critical Level = "Error"
	Medium   Level = "Warning"
	Low      Level = "Info"
)

type SubCategory string

const (
	InsecureStorage          SubCategory = "InsecureStorage"
	Cryptography             SubCategory = "Cryptography"
	InputValidation          SubCategory = "InputValidation"
	Other                    SubCategory = "Other"
	Visibility               SubCategory = "Visibility"
	InsecureModulesLibraries SubCategory = "InsecureModulesLibraries"
	Auth                     SubCategory = "Auth"
	UnexpectedBehaviour      SubCategory = "UnexpectedBehaviour"
)

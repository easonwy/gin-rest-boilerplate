package user

// RegisterUserInput represents the data required to register a new user.
type RegisterUserInput struct {
	Email     string
	Password  string
	FirstName string
	LastName  string
}

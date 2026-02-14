package protocol

const (
	// returns 's' if username is free, 'x' if user exists, anything else on server error
	AuthCheckUser = 1
	// returns 's' if success, 'd' if username doesn't exist, 'p' if password is wrong, any other number for anything else (including rate limiting)
	Auth           = 2
	AuthCreateUser = 3
)

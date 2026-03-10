package protocol

const (
	// returns 's' if success, 'x' if username exists, anything else on fail
	AuthCreateUser = 1
	// returns 's' if success, 'd' if username doesn't exist, 'p' if password is wrong, any other number for anything else (including rate limiting)
	Auth = 2
)

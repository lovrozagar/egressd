package proxy

// Authenticator validates proxy credentials. Anonymous access is never allowed.
type Authenticator interface {
	Authenticate(username, password string) bool
}

// AuthFunc adapts a function to Authenticator.
type AuthFunc func(username, password string) bool

// Authenticate implements Authenticator.
func (f AuthFunc) Authenticate(username, password string) bool {
	return f(username, password)
}

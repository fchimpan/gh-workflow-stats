package github

type RateLimitError struct {
	Err error
}

func (e RateLimitError) Error() string {
	return e.Err.Error()
}

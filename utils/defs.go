package utils

var (
	ErrExtractAuthentication          = &Error{Text: "cannot extract auth data"}
	ErrExtractAuthenticationMalformed = ErrExtractAuthentication.Extend("malformed header")
	ErrExtractAuthenticationPrefix    = ErrExtractAuthentication.Extend("incorrect authorization prefix")
	ErrExtractAuthenticationDelimiter = ErrExtractAuthentication.Extend("cannot find a delimiter in user/passsword pair")
)

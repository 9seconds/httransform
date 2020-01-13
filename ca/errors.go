package ca

import "github.com/9seconds/httransform/v2/utils"

var (
	ErrCA = &utils.Error{Text: "certificate error"}

	ErrCAInvalidCertificates = ErrCA.Extend("invalid certificates")
	ErrCAContextClosed       = ErrCA.Extend("context is closed")
)

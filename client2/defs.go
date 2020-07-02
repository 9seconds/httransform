package client2

import "github.com/9seconds/httransform/v2/utils"

var (
	ErrClient           = &utils.Error{Text: "cannot perform HTTP request"}
	ErrClientCannotDial = ErrClient.Extend("cannot dial to the address")
)

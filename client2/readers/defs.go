package readers

import "github.com/9seconds/httransform/v2/utils"

var (
	ErrReader                    = &utils.Error{Text: "cannot read"}
	ErrReaderTooLargeHexNum      = ErrReader.Extend("too large hex number")
	ErrReaderNoNumber            = ErrReader.Extend("cannot find a number")
	ErrReaderCannotUnread       = ErrReader.Extend("cannot unread byte")
	ErrReaderCannotConsumeCRLF  = ErrReader.Extend("cannot consume CRLF")
	ErrReaderUnexpectedCharacter = ErrReader.Extend("unexpected character")
)

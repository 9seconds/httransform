package readers

import "github.com/PumpkinSeed/errors"

var (
	ErrReader                    = errors.New("cannot read")
	ErrReaderNoNumber            = errors.Wrap(errors.New("no number in transfer-encoding"), ErrReader)
	ErrReaderTooLargeHexNum      = errors.Wrap(errors.New("chunk size is to large"), ErrReader)
	ErrReaderCannotConsumeCRLF   = errors.Wrap(errors.New("cannot consume CRLF"), ErrReader)
	ErrReaderUnexpectedCharacter = errors.Wrap(errors.New("unexpected character"), ErrReader)
)

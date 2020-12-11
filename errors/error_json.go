package errors

import (
	"encoding/json"
	"io"
)

type errorJSON struct {
	Error struct {
		Code    string                `json:"code"`
		Message string                `json:"message"`
		Stack   []errorJSONStackEntry `json:"stack"`
	} `json:"error"`
}

type errorJSONStackEntry struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	StatusCode int    `json:"status_code"`
}

func writeErrorAsJSON(err *Error, writer io.Writer) {
	obj := errorJSON{}
    obj.Error.Code = err.GetChainCode()
    obj.Error.Message = err.GetMessage()
	lastMessage := ""
	foundLastMessage := false

	for current := err; current != nil; current = unwrapError(current) {
		obj.Error.Stack = append(obj.Error.Stack, errorJSONStackEntry{
			Message:    current.Message,
			Code:       current.Code,
			StatusCode: current.StatusCode,
		})

		if current.Err != nil {
			if _, ok := current.Err.(*Error); !ok { // nolint: errorlint
				foundLastMessage = true
				lastMessage = current.Err.Error()
			}
		}
	}

	if foundLastMessage {
		obj.Error.Stack = append(obj.Error.Stack, errorJSONStackEntry{
			Message: lastMessage,
		})
	}

	encoder := json.NewEncoder(writer)

	encoder.SetEscapeHTML(false) // omg, why true is default!?

	if err2 := encoder.Encode(&obj); err2 != nil {
		panic(err2)
	}
}

package errors_test

import (
	"fmt"
	"io"
	"testing"

	"github.com/9seconds/httransform/v2/errors"
	"github.com/stretchr/testify/suite"
	"github.com/valyala/fasthttp"
	"github.com/xeipuuv/gojsonschema"
)

const errorJSONSchema = `
{
  "type": "object",
  "required": ["error"],

  "properties": {
    "error": {
      "type": "object",
      "required": ["code", "message", "stack"],

      "properties": {
        "code": {"type": "string"},
        "message": {"type": "string"},
        "stack": {
          "type": "array",

          "items": {
            "type": "object",
            "required": ["code", "message", "status_code"],

            "properties": {
              "code": {"type": "string"},
              "message": {"type": "string"},
              "status_code": {"type": "integer", "minimum": 0}
            }
          }
        }
      }
    }
  }
}
`

var errorJSONValidator = func() *gojsonschema.Schema {
	loader := gojsonschema.NewStringLoader(errorJSONSchema)
	rv, err := gojsonschema.NewSchema(loader)

	if err != nil {
		panic(err)
	}

	return rv
}()

type ErrorTestSuite struct {
	suite.Suite

	e *errors.Error
}

func (suite *ErrorTestSuite) SetupTest() {
	suite.e = nil
}

func (suite *ErrorTestSuite) TestErrorMessage() {
	suite.Equal("", suite.e.Error())

	suite.e = &errors.Error{
		Err: errors.New("EOF"),
	}

	suite.Equal("EOF", suite.e.Error())

	suite.e.Message = "MyMessage"

	suite.Contains(suite.e.Error(), "MyMessage")
	suite.Contains(suite.e.Error(), "EOF")

	suite.e = &errors.Error{
		Message: "Another",
		Err:     suite.e,
	}

	suite.Contains(suite.e.Error(), "MyMessage")
	suite.Contains(suite.e.Error(), "EOF")
	suite.Contains(suite.e.Error(), "Another")

	suite.e = &errors.Error{
		Message: "QQ",
	}

	suite.Equal("QQ", suite.e.Error())
}

func (suite *ErrorTestSuite) TestUnwrap() {
	suite.Nil(suite.e.Unwrap())

	suite.e = &errors.Error{
		Err: io.EOF,
	}

	suite.True(errors.Is(suite.e.Unwrap(), io.EOF))
}

func (suite *ErrorTestSuite) TestGetStatusCode() {
	suite.Equal(0, suite.e.GetStatusCode())

	suite.e = &errors.Error{
		StatusCode: 1,
	}

	suite.Equal(1, suite.e.GetStatusCode())
}

func (suite *ErrorTestSuite) TestGetChainStatusCode() {
	suite.Equal(errors.DefaultChainStatusCode, suite.e.GetChainStatusCode())

	suite.e = &errors.Error{
		StatusCode: 1,
	}

	suite.Equal(1, suite.e.GetChainStatusCode())

	suite.e = &errors.Error{
		Err: &errors.Error{
			StatusCode: 2,
			Err: &errors.Error{
				StatusCode: 3,
			},
		},
	}

	suite.Equal(2, suite.e.GetChainStatusCode())

	suite.e = &errors.Error{
		Err: &errors.Error{
			StatusCode: 0,
			Err: fmt.Errorf("another error: %w", &errors.Error{
				StatusCode: 2,
			}),
		},
	}

	suite.Equal(errors.DefaultChainStatusCode, suite.e.GetChainStatusCode())
}

func (suite *ErrorTestSuite) TestGetMessage() {
	suite.Equal("", suite.e.GetMessage())

	suite.e = &errors.Error{
		Message: "111",
		Err:     io.EOF,
	}

	suite.Equal("111", suite.e.GetMessage())
}

func (suite *ErrorTestSuite) TestGetCode() {
	suite.Equal("", suite.e.GetCode())

	suite.e = &errors.Error{
		Code: "111",
		Err: &errors.Error{
			Code: "222",
		},
	}

	suite.Equal("111", suite.e.GetCode())
}

func (suite *ErrorTestSuite) TestGetChainCode() {
	suite.Equal(errors.DefaultChainErrorCode, suite.e.GetChainCode())

	suite.e = &errors.Error{
		Code: "111",
		Err: &errors.Error{
			Code: "222",
		},
	}

	suite.Equal("111", suite.e.GetChainCode())

	suite.e.Code = ""

	suite.Equal("222", suite.e.GetChainCode())

	suite.e = &errors.Error{
		Err: &errors.Error{
			Err: fmt.Errorf("???: %w", &errors.Error{
				Code: "333",
			}),
		},
	}

	suite.Equal(errors.DefaultChainErrorCode, suite.e.GetChainCode())
}

func (suite *ErrorTestSuite) TestErrorJSON() {
	js := gojsonschema.NewStringLoader(suite.e.ErrorJSON())
	result, err := errorJSONValidator.Validate(js)

	suite.NoError(err)
	suite.True(result.Valid())

	suite.e = &errors.Error{
		Code:    "lala",
		Message: "blabla",
	}

	js = gojsonschema.NewStringLoader(suite.e.ErrorJSON())
	result, err = errorJSONValidator.Validate(js)

	suite.NoError(err)
	suite.True(result.Valid())

	suite.e = &errors.Error{
		Code:    "lala",
		Message: "blabla",
		Err: &errors.Error{
			StatusCode: 100,
			Err:        io.EOF,
		},
	}

	js = gojsonschema.NewStringLoader(suite.e.ErrorJSON())
	result, err = errorJSONValidator.Validate(js)

	suite.NoError(err)
	suite.True(result.Valid())
}

func (suite *ErrorTestSuite) TestWriteTo() {
	ctx := &fasthttp.RequestCtx{}

	suite.e = &errors.Error{
		Code:    "lala",
		Message: "blabla",
		Err: &errors.Error{
			StatusCode: 100,
			Err:        io.EOF,
		},
	}

	suite.e.WriteTo(ctx)

	suite.True(ctx.Response.ConnectionClose())
	suite.Equal("application/json", string(ctx.Response.Header.ContentType()))

	js := gojsonschema.NewStringLoader(string(ctx.Response.Body()))
	result, err := errorJSONValidator.Validate(js)

	suite.NoError(err)
	suite.True(result.Valid())
}

func TestError(t *testing.T) {
	suite.Run(t, &ErrorTestSuite{})
}

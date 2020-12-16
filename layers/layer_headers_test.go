package layers_test

import (
	"io"
	"testing"

	"github.com/9seconds/httransform/v2/headers"
	"github.com/9seconds/httransform/v2/layers"
	"github.com/stretchr/testify/suite"
)

type LayerHeadersTestSuite struct {
	BaseLayerTestSuite
}

func (suite *LayerHeadersTestSuite) SetupTest() {
	suite.BaseLayerTestSuite.SetupTest()

	suite.l = &layers.HeadersLayer{
		RequestSet: []headers.Header{
			headers.NewHeader("RequestSet", "1"),
		},
		RequestSetExact: []headers.Header{
			headers.NewHeader("RequestSetExact", "1"),
		},
		RequestRemove: []string{
			"RequestRemove",
		},
		RequestRemoveExact: []string{
			"RequestRemoveExact",
		},

		ResponseOkSet: []headers.Header{
			headers.NewHeader("ResponseOkSet", "1"),
		},
		ResponseOkSetExact: []headers.Header{
			headers.NewHeader("ResponseOkSetExact", "1"),
		},
		ResponseOkRemove: []string{
			"ResponseOkRemove",
		},
		ResponseOkRemoveExact: []string{
			"ResponseOkRemoveExact",
		},

		ResponseErrSet: []headers.Header{
			headers.NewHeader("ResponseErrSet", "1"),
		},
		ResponseErrSetExact: []headers.Header{
			headers.NewHeader("ResponseErrSetExact", "1"),
		},
		ResponseErrRemove: []string{
			"ResponseErrRemove",
		},
		ResponseErrRemoveExact: []string{
			"ResponseErrRemoveExact",
		},
	}

	suite.ctx.RequestHeaders.Append("requestSet", "2")
	suite.ctx.RequestHeaders.Append("requestSetExact", "2")
	suite.ctx.RequestHeaders.Append("RequestSetExact", "3")
	suite.ctx.RequestHeaders.Append("requestRemove", "3")
	suite.ctx.RequestHeaders.Append("RequestRemoveExact", "3")
	suite.ctx.RequestHeaders.Append("requestRemoveExact", "4")

	suite.ctx.ResponseHeaders.Append("responseOkSet", "2")
	suite.ctx.ResponseHeaders.Append("responseOkSetExact", "2")
	suite.ctx.ResponseHeaders.Append("ResponseOkSetExact", "3")
	suite.ctx.ResponseHeaders.Append("responseOkRemove", "2")
	suite.ctx.ResponseHeaders.Append("ResponseOkRemoveExact", "2")
	suite.ctx.ResponseHeaders.Append("responseOkRemoveExact", "2")

	suite.ctx.ResponseHeaders.Append("responseErrSet", "2")
	suite.ctx.ResponseHeaders.Append("responseErrSetExact", "2")
	suite.ctx.ResponseHeaders.Append("ResponseErrSetExact", "3")
	suite.ctx.ResponseHeaders.Append("responseErrRemove", "2")
	suite.ctx.ResponseHeaders.Append("ResponseErrRemoveExact", "2")
	suite.ctx.ResponseHeaders.Append("responseErrRemoveExact", "2")
}

func (suite *LayerHeadersTestSuite) TestOk() {
	suite.NoError(suite.l.OnRequest(suite.ctx))
	suite.NoError(suite.l.OnResponse(suite.ctx, nil))

	suite.Equal("1", suite.ctx.RequestHeaders.GetFirst("RequestSet").Value())
	suite.Equal("1", suite.ctx.RequestHeaders.GetLastExact("requestSet").Value())
	suite.Equal("2", suite.ctx.RequestHeaders.GetFirst("requestSetExact").Value())
	suite.Equal("1", suite.ctx.RequestHeaders.GetLast("requestSetExact").Value())
	suite.Equal("2", suite.ctx.RequestHeaders.GetFirstExact("requestSetExact").Value())
	suite.Nil(suite.ctx.RequestHeaders.GetFirst("requestRemove"))
	suite.Equal("4", suite.ctx.RequestHeaders.GetFirst("requestRemoveExact").Value())

	suite.Equal("1", suite.ctx.ResponseHeaders.GetFirst("responseOkSet").Value())
	suite.Equal("2", suite.ctx.ResponseHeaders.GetFirst("responseOkSetExact").Value())
	suite.Equal("1", suite.ctx.ResponseHeaders.GetFirstExact("ResponseOkSetExact").Value())
	suite.Nil(suite.ctx.RequestHeaders.GetFirst("responseOkremove"))
	suite.Nil(suite.ctx.ResponseHeaders.GetFirstExact("ResponseOkRemoveExact"))
	suite.NotNil(suite.ctx.ResponseHeaders.GetFirst("ResponseOkRemoveExact"))

	suite.Equal("2", suite.ctx.ResponseHeaders.GetFirstExact("responseErrSet").Value())
	suite.Equal("2", suite.ctx.ResponseHeaders.GetFirstExact("responseErrSetExact").Value())
	suite.Equal("3", suite.ctx.ResponseHeaders.GetFirstExact("ResponseErrSetExact").Value())
	suite.Equal("2", suite.ctx.ResponseHeaders.GetFirstExact("responseErrRemove").Value())
	suite.Equal("2", suite.ctx.ResponseHeaders.GetFirstExact("ResponseErrRemoveExact").Value())
	suite.Equal("2", suite.ctx.ResponseHeaders.GetFirstExact("responseErrRemoveExact").Value())
}

func (suite *LayerHeadersTestSuite) TestErr() {
	suite.NoError(suite.l.OnRequest(suite.ctx))
	suite.Equal(io.EOF, suite.l.OnResponse(suite.ctx, io.EOF))

	suite.Equal("2", suite.ctx.ResponseHeaders.GetFirstExact("responseOkSet").Value())
	suite.Equal("2", suite.ctx.ResponseHeaders.GetFirstExact("responseOkSetExact").Value())
	suite.Equal("3", suite.ctx.ResponseHeaders.GetFirstExact("ResponseOkSetExact").Value())
	suite.NotNil(suite.ctx.ResponseHeaders.GetFirst("ResponseOkRemove"))
	suite.NotNil(suite.ctx.ResponseHeaders.GetFirst("ResponseOkRemoveExact"))

	suite.Equal("1", suite.ctx.ResponseHeaders.GetFirstExact("responseErrSet").Value())
	suite.Equal("2", suite.ctx.ResponseHeaders.GetFirstExact("responseErrSetExact").Value())
	suite.Equal("1", suite.ctx.ResponseHeaders.GetFirstExact("ResponseErrSetExact").Value())
	suite.Nil(suite.ctx.ResponseHeaders.GetFirstExact("responseErrRemove"))
	suite.Nil(suite.ctx.ResponseHeaders.GetFirstExact("ResponseErrRemoveExact"))
	suite.Equal("2", suite.ctx.ResponseHeaders.GetFirstExact("responseErrRemoveExact").Value())
}

func TestLayerHeaders(t *testing.T) {
	suite.Run(t, &LayerHeadersTestSuite{})
}

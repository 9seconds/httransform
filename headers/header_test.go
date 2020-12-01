package headers_test

import (
	"testing"

	"github.com/9seconds/httransform/v2/headers"
	"github.com/stretchr/testify/suite"
)

type HeaderTestSuite struct {
	suite.Suite
}

func (suite *HeaderTestSuite) TestNil() {
	var h *headers.Header

	suite.Empty(h.ID())
	suite.Equal("", h.Name())
	suite.Equal("", h.Value())
	suite.Equal("", h.CanonicalName())
	suite.Empty(h.Values())
	suite.Equal("{no-key}:{no-value}", h.String())
}

func (suite *HeaderTestSuite) TestOk() {
	h := headers.NewHeader("accept-Encoding", "deflate,gzip , br")

	suite.Equal("accept-Encoding", h.Name())
	suite.Equal("Accept-Encoding", h.CanonicalName())
	suite.Equal("deflate,gzip , br", h.Value())
	suite.Equal([]string{"deflate", "gzip ", "br"}, h.Values())
	suite.Equal("accept-Encoding:deflate,gzip , br", h.String())
}

func (suite *HeaderTestSuite) TestSingleValue() {
	h := headers.NewHeader("accept-Encoding", "gzip")

	suite.ElementsMatch([]string{"gzip"}, h.Values())
}

func (suite *HeaderTestSuite) TestNoValue() {
	h := headers.NewHeader("accept-Encoding", "")

	suite.Empty(h.Values())
}

func (suite *HeaderTestSuite) TestID() {
	h1 := headers.NewHeader("accept-Encoding", "")
	h2 := headers.NewHeader("Accept-Encoding", "")
	h3 := headers.NewHeader("accept-encoding", "")
	h4 := headers.NewHeader("ACCEPT-ENCODING", "")

	h11 := headers.NewHeader("ACCEPTENCODING", "")
	h12 := headers.NewHeader("acceptencoding", "")

	suite.Equal(h1.ID(), h2.ID())
	suite.Equal(h1.ID(), h3.ID())
	suite.Equal(h1.ID(), h4.ID())
	suite.Equal(h11.ID(), h12.ID())
	suite.NotEqual(h11.ID(), h1.ID())
	suite.NotEmpty(h1.ID())
	suite.NotEmpty(h11.ID())
}

func (suite *HeaderTestSuite) TestWithName() {
	h1 := headers.NewHeader("accept-Encoding", "")
	h2 := h1.WithName("Accept")

	suite.Equal(h1.Value(), h2.Value())
	suite.NotEqual(h1.ID(), h2.ID())
	suite.NotEqual(h1.Name(), h2.Name())
	suite.Equal("Accept", h2.Name())
}

func (suite *HeaderTestSuite) TestWithValue() {
	h1 := headers.NewHeader("accept-Encoding", "")
	h2 := h1.WithValue("Accept")

	suite.NotEqual(h1.Value(), h2.Value())
	suite.Equal(h1.ID(), h2.ID())
	suite.Equal(h1.Name(), h2.Name())
	suite.Equal("Accept", h2.Value())
}

func TestHeader(t *testing.T) {
	suite.Run(t, &HeaderTestSuite{})
}

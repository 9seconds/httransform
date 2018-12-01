package httransform

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type HeaderSetTestSuite struct {
	suite.Suite

	set      *HeaderSet
	bytesKey []byte
	strKey   string
}

func (suite *HeaderSetTestSuite) SetupTest() {
	suite.set = getHeaderSet()
	suite.bytesKey = []byte("BytesKey")
	suite.strKey = "StrKey"
}

func (suite *HeaderSetTestSuite) TearDownTest() {
	releaseHeaderSet(suite.set)
}

func (suite *HeaderSetTestSuite) TestEmptyGet() {
	_, bytesOk := suite.set.GetBytes(suite.bytesKey)
	_, strOk := suite.set.GetString(suite.strKey)

	suite.False(bytesOk)
	suite.False(strOk)
}

func (suite *HeaderSetTestSuite) TestEmptyDeleteGet() {
	suite.set.DeleteBytes(suite.bytesKey)
	suite.set.DeleteString(suite.strKey)

	_, bytesOk := suite.set.GetBytes(suite.bytesKey)
	_, strOk := suite.set.GetString(suite.strKey)

	suite.False(bytesOk)
	suite.False(strOk)
}

func (suite *HeaderSetTestSuite) TestSetGetDeleteGet() {
	suite.set.SetBytes(suite.bytesKey, []byte("BytesValue"))
	suite.set.SetString(suite.strKey, "StrValue")

	bytesValue, bytesOk := suite.set.GetBytes(suite.bytesKey)
	strValue, strOk := suite.set.GetString(suite.strKey)

	suite.True(bytesOk)
	suite.Equal(bytesValue, []byte("BytesValue"))
	suite.True(strOk)
	suite.Equal(strValue, "StrValue")

	suite.set.DeleteBytes(suite.bytesKey)
	suite.set.DeleteString(suite.strKey)

	_, bytesOk = suite.set.GetBytes(suite.bytesKey)
	_, strOk = suite.set.GetString(suite.strKey)

	suite.False(bytesOk)
	suite.False(strOk)
}

func (suite *HeaderSetTestSuite) TestHeaderCaseOrder() {
	suite.set.SetString("str1", "val1")
	suite.set.SetString("Str2", "vAl2")
	suite.set.SetString("STR3", "val3")

	items := suite.set.Items()
	suite.Len(items, 3)
	suite.Equal(items[0].Key, []byte("str1"))
	suite.Equal(items[0].Value, []byte("val1"))
	suite.Equal(items[1].Key, []byte("Str2"))
	suite.Equal(items[1].Value, []byte("vAl2"))
	suite.Equal(items[2].Key, []byte("STR3"))
	suite.Equal(items[2].Value, []byte("val3"))
}

func (suite *HeaderSetTestSuite) TestHeaderOverride() {
	suite.set.SetString("str1", "val1")
	suite.set.SetString("Str1", "vAl2")
	suite.set.SetString("STR3", "val3")

	items := suite.set.Items()
	suite.Len(items, 2)
	suite.Equal(items[0].Key, []byte("str1"))
	suite.Equal(items[0].Value, []byte("vAl2"))
	suite.Equal(items[1].Key, []byte("STR3"))
	suite.Equal(items[1].Value, []byte("val3"))
}

func (suite *HeaderSetTestSuite) TestDeleteKeepsOrder() {
	suite.set.SetString("str1", "val1")
	suite.set.SetString("Str2", "vAl2")
	suite.set.SetString("STR3", "val3")
	suite.set.SetString("sTR4", "val4")

	suite.set.DeleteString("STR3")

	items := suite.set.Items()
	suite.Len(items, 3)
	suite.Equal(items[0].Key, []byte("str1"))
	suite.Equal(items[0].Value, []byte("val1"))
	suite.Equal(items[1].Key, []byte("Str2"))
	suite.Equal(items[1].Value, []byte("vAl2"))
	suite.Equal(items[2].Key, []byte("sTR4"))
	suite.Equal(items[2].Value, []byte("val4"))
}

func (suite *HeaderSetTestSuite) TestDeleteWriteRestore() {
	suite.set.SetString("str1", "val1")
	suite.set.SetString("Str2", "vAl2")
	suite.set.SetString("STR3", "val3")
	suite.set.SetString("sTR4", "val4")

	suite.set.DeleteString("STR3")
	suite.set.SetString("str5", "val5")
	suite.set.SetString("str3", "---")

	items := suite.set.Items()
	suite.Len(items, 5)
	suite.Equal(items[0].Key, []byte("str1"))
	suite.Equal(items[0].Value, []byte("val1"))
	suite.Equal(items[1].Key, []byte("Str2"))
	suite.Equal(items[1].Value, []byte("vAl2"))
	suite.Equal(items[2].Key, []byte("STR3"))
	suite.Equal(items[2].Value, []byte("---"))
	suite.Equal(items[3].Key, []byte("sTR4"))
	suite.Equal(items[3].Value, []byte("val4"))
	suite.Equal(items[4].Key, []byte("str5"))
	suite.Equal(items[4].Value, []byte("val5"))
}

type ParseHeadersTestSuite struct {
	suite.Suite

	set *HeaderSet
}

func (suite *ParseHeadersTestSuite) SetupTest() {
	suite.set = getHeaderSet()
}

func (suite *ParseHeadersTestSuite) TearDownTest() {
	releaseHeaderSet(suite.set)
}

func (suite *ParseHeadersTestSuite) TestEmpty() {
	suite.NotNil(parseHeaders(suite.set, nil))
}

func (suite *ParseHeadersTestSuite) TestOnlyFirstLine() {
	suite.NotNil(parseHeaders(suite.set, []byte("GET / HTTP/1.1\r\n")))
}

func (suite *ParseHeadersTestSuite) TestNoHeaders() {
	suite.Nil(parseHeaders(suite.set, []byte("GET / HTTP/1.1\r\n\r\n")))
	suite.Len(suite.set.Items(), 0)
}

func (suite *ParseHeadersTestSuite) TestMalformedHeader() {
	text := []byte(`GET / HTTP/1.1
Header: Value
H

`)
	suite.NotNil(parseHeaders(suite.set, text))
}

func (suite *ParseHeadersTestSuite) TestHeaderOrder() {
	text := []byte(`GET / HTTP/1.1
STR1: Value1
Str2: value2
Val3: xxx
	lalala
val3: 1

`)
	suite.Nil(parseHeaders(suite.set, text))

	items := suite.set.Items()
	suite.Len(items, 3)
	suite.Equal(items[0].Key, []byte("STR1"))
	suite.Equal(items[0].Value, []byte("Value1"))
	suite.Equal(items[1].Key, []byte("Str2"))
	suite.Equal(items[1].Value, []byte("value2"))
	suite.Equal(items[2].Key, []byte("Val3"))
	suite.Equal(items[2].Value, []byte("1"))
}

func TestHeaderSet(t *testing.T) {
	suite.Run(t, &HeaderSetTestSuite{})
}

func TestParseHeaders(t *testing.T) {
	suite.Run(t, &ParseHeadersTestSuite{})
}

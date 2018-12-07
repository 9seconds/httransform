package httransform

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type NoopMetricsTestSuite struct {
	suite.Suite

	metrics Metrics
}

func (suite *NoopMetricsTestSuite) SetupTest() {
	suite.metrics = &NoopMetrics{}
}

func (suite *NoopMetricsTestSuite) TestDummy() {
	suite.metrics.NewConnection()
	suite.metrics.DropConnection()
	suite.metrics.NewGet()
	suite.metrics.NewHead()
	suite.metrics.NewPost()
	suite.metrics.NewPut()
	suite.metrics.NewDelete()
	suite.metrics.NewConnect()
	suite.metrics.NewOptions()
	suite.metrics.NewTrace()
	suite.metrics.NewPatch()
	suite.metrics.DropGet()
	suite.metrics.DropHead()
	suite.metrics.DropPost()
	suite.metrics.DropPut()
	suite.metrics.DropDelete()
	suite.metrics.DropConnect()
	suite.metrics.DropOptions()
	suite.metrics.DropTrace()
	suite.metrics.DropPatch()
	suite.metrics.DropOther()
	suite.metrics.NewCertificate()
	suite.metrics.DropCertificate()
}

type MetricsValueTestSuite struct {
	suite.Suite

	metrics *MockMetrics
}

func (suite *MetricsValueTestSuite) SetupTest() {
	suite.metrics = &MockMetrics{}
}

func (suite *MetricsValueTestSuite) TestGet() {
	suite.metrics.On("NewGet")
	suite.metrics.On("DropGet")

	newMethodMetricsValue(suite.metrics, metricsStrGet)
	dropMethodMetricsValue(suite.metrics, metricsStrGet)

	suite.metrics.AssertExpectations(suite.T())
}

func (suite *MetricsValueTestSuite) TestHead() {
	suite.metrics.On("NewHead")
	suite.metrics.On("DropHead")

	newMethodMetricsValue(suite.metrics, metricsStrHead)
	dropMethodMetricsValue(suite.metrics, metricsStrHead)

	suite.metrics.AssertExpectations(suite.T())
}

func (suite *MetricsValueTestSuite) TestPost() {
	suite.metrics.On("NewPost")
	suite.metrics.On("DropPost")

	newMethodMetricsValue(suite.metrics, metricsStrPost)
	dropMethodMetricsValue(suite.metrics, metricsStrPost)

	suite.metrics.AssertExpectations(suite.T())
}

func (suite *MetricsValueTestSuite) TestPut() {
	suite.metrics.On("NewPut")
	suite.metrics.On("DropPut")

	newMethodMetricsValue(suite.metrics, metricsStrPut)
	dropMethodMetricsValue(suite.metrics, metricsStrPut)

	suite.metrics.AssertExpectations(suite.T())
}

func (suite *MetricsValueTestSuite) TestDelete() {
	suite.metrics.On("NewDelete")
	suite.metrics.On("DropDelete")

	newMethodMetricsValue(suite.metrics, metricsStrDelete)
	dropMethodMetricsValue(suite.metrics, metricsStrDelete)

	suite.metrics.AssertExpectations(suite.T())
}

func (suite *MetricsValueTestSuite) TestConnect() {
	suite.metrics.On("NewConnect")
	suite.metrics.On("DropConnect")

	newMethodMetricsValue(suite.metrics, metricsStrConnect)
	dropMethodMetricsValue(suite.metrics, metricsStrConnect)

	suite.metrics.AssertExpectations(suite.T())
}

func (suite *MetricsValueTestSuite) TestOptions() {
	suite.metrics.On("NewOptions")
	suite.metrics.On("DropOptions")

	newMethodMetricsValue(suite.metrics, metricsStrOptions)
	dropMethodMetricsValue(suite.metrics, metricsStrOptions)

	suite.metrics.AssertExpectations(suite.T())
}

func (suite *MetricsValueTestSuite) TestTrace() {
	suite.metrics.On("NewTrace")
	suite.metrics.On("DropTrace")

	newMethodMetricsValue(suite.metrics, metricsStrTrace)
	dropMethodMetricsValue(suite.metrics, metricsStrTrace)

	suite.metrics.AssertExpectations(suite.T())
}

func (suite *MetricsValueTestSuite) TestPatch() {
	suite.metrics.On("NewPatch")
	suite.metrics.On("DropPatch")

	newMethodMetricsValue(suite.metrics, metricsStrPatch)
	dropMethodMetricsValue(suite.metrics, metricsStrPatch)

	suite.metrics.AssertExpectations(suite.T())
}

func (suite *MetricsValueTestSuite) TestOther() {
	suite.metrics.On("NewOther")
	suite.metrics.On("DropOther")

	newMethodMetricsValue(suite.metrics, nil)
	dropMethodMetricsValue(suite.metrics, nil)

	suite.metrics.AssertExpectations(suite.T())
}

func TestNoopMetrics(t *testing.T) {
	suite.Run(t, &NoopMetricsTestSuite{})
}

func TestMetricsValue(t *testing.T) {
	suite.Run(t, &MetricsValueTestSuite{})
}

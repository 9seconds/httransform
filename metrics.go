package httransform

import "bytes"

var (
	// copy-pasted from fasthttp for the faster access
	metricsStrGet     = []byte("GET")
	metricsStrHead    = []byte("HEAD")
	metricsStrPost    = []byte("POST")
	metricsStrPut     = []byte("PUT")
	metricsStrDelete  = []byte("DELETE")
	metricsStrConnect = []byte("CONNECT")
	metricsStrOptions = []byte("OPTIONS")
	metricsStrTrace   = []byte("TRACE")
	metricsStrPatch   = []byte("PATCH")
)

// Metrics is an interface for notifying on multiple events which can
// happen within a framework. For example, open/close connections, some
// method counters etc.
type Metrics interface {
	// NewConnection reports new established connection.
	NewConnection()

	// DropConnection reports connection closed.
	DropConnection()

	// NewGet reports about new HTTP GET request.
	NewGet()

	// NewHead reports about new HTTP HEAD request.
	NewHead()

	// NewPost reports about new HTTP POST request.
	NewPost()

	// NewPut reports about new HTTP PUT request.
	NewPut()

	// NewDelete reports about new HTTP DELETE request.
	NewDelete()

	// NewConnect reports about new HTTP CONNECT request.
	NewConnect()

	// NewOptions reports about new HTTP OPTIONS request.
	NewOptions()

	// NewTrace reports about new HTTP TRACE request.
	NewTrace()

	// NewPatch reports about new HTTP PATCH request.
	NewPatch()

	// NewOther reports about new HTTP request with rate verb.
	NewOther()

	// DropGet reports about closed HTTP GET request.
	DropGet()

	// DropHead reports about closed HTTP HEAD request.
	DropHead()

	// DropPost reports about closed HTTP POST request.
	DropPost()

	// DropPut reports about closed HTTP PUT request.
	DropPut()

	// DropDelete reports about closed HTTP DELETE request.
	DropDelete()

	// DropConnect reports about closed HTTP CONNECT request.
	DropConnect()

	// DropOptions reports about closed HTTP OPTIONS request.
	DropOptions()

	// DropTrace reports about closed HTTP TRACE request.
	DropTrace()

	// DropPatch reports about closed HTTP PATCH request.
	DropPatch()

	// DropOther reports about closed HTTP request with rate verb.
	DropOther()

	// NewCertificate reports about new generated TLS certificate.
	NewCertificate()

	// DropCertificate reports about pruned TLS certificate.
	DropCertificate()
}

// NoopMetrics is a metrics structure which does nothing.
type NoopMetrics struct{}

// NewConnection reports new established connection.
func (n *NoopMetrics) NewConnection() {}

// DropConnection reports connection closed.
func (n *NoopMetrics) DropConnection() {}

// NewGet reports about new HTTP GET request.
func (n *NoopMetrics) NewGet() {}

// NewHead reports about new HTTP HEAD request.
func (n *NoopMetrics) NewHead() {}

// NewPost reports about new HTTP POST request.
func (n *NoopMetrics) NewPost() {}

// NewPut reports about new HTTP PUT request.
func (n *NoopMetrics) NewPut() {}

// NewDelete reports about new HTTP DELETE request.
func (n *NoopMetrics) NewDelete() {}

// NewConnect reports about new HTTP CONNECT request.
func (n *NoopMetrics) NewConnect() {}

// NewOptions reports about new HTTP OPTIONS request.
func (n *NoopMetrics) NewOptions() {}

// NewTrace reports about new HTTP TRACE request.
func (n *NoopMetrics) NewTrace() {}

// NewPatch reports about new HTTP PATCH request.
func (n *NoopMetrics) NewPatch() {}

// NewOther reports about new HTTP request with rate verb.
func (n *NoopMetrics) NewOther() {}

// DropGet reports about closed HTTP GET request.
func (n *NoopMetrics) DropGet() {}

// DropHead reports about closed HTTP HEAD request.
func (n *NoopMetrics) DropHead() {}

// DropPost reports about closed HTTP POST request.
func (n *NoopMetrics) DropPost() {}

// DropPut reports about closed HTTP PUT request.
func (n *NoopMetrics) DropPut() {}

// DropDelete reports about closed HTTP DELETE request.
func (n *NoopMetrics) DropDelete() {}

// DropConnect reports about closed HTTP CONNECT request.
func (n *NoopMetrics) DropConnect() {}

// DropOptions reports about closed HTTP OPTIONS request.
func (n *NoopMetrics) DropOptions() {}

// DropTrace reports about closed HTTP TRACE request.
func (n *NoopMetrics) DropTrace() {}

// DropPatch reports about closed HTTP PATCH request.
func (n *NoopMetrics) DropPatch() {}

// DropOther reports about closed HTTP request with rate verb.
func (n *NoopMetrics) DropOther() {}

// NewCertificate reports about new generated TLS certificate.
func (n *NoopMetrics) NewCertificate() {}

// DropCertificate reports about pruned TLS certificate.
func (n *NoopMetrics) DropCertificate() {}

func newMethodMetricsValue(m Metrics, value []byte) {
	switch {
	case bytes.Equal(value, metricsStrConnect):
		m.NewConnect()
	case bytes.Equal(value, metricsStrGet):
		m.NewGet()
	case bytes.Equal(value, metricsStrOptions):
		m.NewOptions()
	case bytes.Equal(value, metricsStrPost):
		m.NewPost()
	case bytes.Equal(value, metricsStrPut):
		m.NewPut()
	case bytes.Equal(value, metricsStrHead):
		m.NewHead()
	case bytes.Equal(value, metricsStrDelete):
		m.NewDelete()
	case bytes.Equal(value, metricsStrTrace):
		m.NewTrace()
	case bytes.Equal(value, metricsStrPatch):
		m.NewPatch()
	default:
		m.NewOther()
	}
}

func dropMethodMetricsValue(m Metrics, value []byte) {
	switch {
	case bytes.Equal(value, metricsStrConnect):
		m.DropConnect()
	case bytes.Equal(value, metricsStrGet):
		m.DropGet()
	case bytes.Equal(value, metricsStrOptions):
		m.DropOptions()
	case bytes.Equal(value, metricsStrPost):
		m.DropPost()
	case bytes.Equal(value, metricsStrPut):
		m.DropPut()
	case bytes.Equal(value, metricsStrHead):
		m.DropHead()
	case bytes.Equal(value, metricsStrDelete):
		m.DropDelete()
	case bytes.Equal(value, metricsStrTrace):
		m.DropTrace()
	case bytes.Equal(value, metricsStrPatch):
		m.DropPatch()
	default:
		m.DropOther()
	}
}

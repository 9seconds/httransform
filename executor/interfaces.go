package executor

import "github.com/9seconds/httransform/v2/layers"

// Executor transforms HTTP request to HTTP response and does some
// additional actions like management of connection upgrades.
type Executor func(*layers.Context) error

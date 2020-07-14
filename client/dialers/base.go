package dialers

import "time"

type baseDialer struct {
	timeout time.Duration
}

func (b *baseDialer) GetTimeout() time.Duration {
	return b.timeout
}

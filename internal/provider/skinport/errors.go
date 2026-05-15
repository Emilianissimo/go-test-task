package skinport

import (
	"fmt"
)

var (
	ErrUpstreamError   = fmt.Errorf("skinport api returned non-200 status")
	ErrRateLimit       = fmt.Errorf("skinport rate limit exceeded")
	ErrDecodingContent = fmt.Errorf("failed to decode skinport response")
)

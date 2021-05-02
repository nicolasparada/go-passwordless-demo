package notification

import (
	"context"
	"io"
)

type ComposeFunc func(ctx context.Context, to string, w io.Writer, data interface{}) error

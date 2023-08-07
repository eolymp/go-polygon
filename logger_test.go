package polygon

import "testing"

type loggerMock struct {
	t *testing.T
}

func (l *loggerMock) Warning(message string, params map[string]any) {
}

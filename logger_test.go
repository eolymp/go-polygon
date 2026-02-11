package polygon

import (
	"fmt"
	"testing"
)

type loggerMock struct {
	t *testing.T
}

func (l *loggerMock) Printf(format string, args ...any) {
	l.t.Log("logger.PRINT: " + fmt.Sprintf(format, args...))
}

func (l *loggerMock) Errorf(format string, args ...any) {
	l.t.Log("logger.ERROR: " + fmt.Sprintf(format, args...))
}

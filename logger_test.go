package polygon

import (
	"fmt"
	"testing"
)

type loggerMock struct {
	t *testing.T
}

func (l *loggerMock) Printf(format string, args ...any) {
	l.t.Logf("logger.PRINT: " + fmt.Sprintf(format, args...))
}

func (l *loggerMock) Errorf(format string, args ...any) {
	l.t.Logf("logger.ERROR: " + fmt.Sprintf(format, args...))
}

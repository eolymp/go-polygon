package polygon

import (
	"encoding/json"
	"testing"
)

type loggerMock struct {
	t *testing.T
}

func (l *loggerMock) Warning(message string, params map[string]any) {
	payload, _ := json.Marshal(params)
	l.t.Logf("logger.WARNING: %s %s", message, payload)
}

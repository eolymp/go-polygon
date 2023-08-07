package polygon

type logger interface {
	Warning(message string, attr map[string]any)
}

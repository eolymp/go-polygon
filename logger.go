package polygon

type logger interface {
	Printf(format string, args ...any)
	Errorf(format string, args ...any)
}

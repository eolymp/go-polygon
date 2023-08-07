package polygon

var runtimeMapping = map[string][]string{
	"cpp:17-gnu10": {"c.gcc", "cpp.g++", "cpp.g++11", "cpp.g++14", "cpp.g++17", "cpp.ms", "cpp.msys2-mingw64-9-g++17"},
	"csharp":       {"csharp.mono"},
	"d":            {"d"},
	"go":           {"go"},
	"java":         {"java11", "java8"},
	"kotlin":       {"kotlin"},
	"fpc":          {"pas.dpr", "pas.fpc"},
	"php":          {"php.5"},
	"python":       {"python.2", "python.3"},
	"pypy":         {"python.pypy2", "python.pypy3"},
	"ruby":         {"ruby"},
	"rust":         {"rust"},
}

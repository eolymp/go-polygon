package polygon

var runtimeMapping = map[string][]string{
	"c:17-gnu10":      {"c.gcc"},
	"cpp:17-gnu10":    {"cpp.g++", "cpp.g++11", "cpp.g++14", "cpp.g++17", "cpp.ms", "cpp.msys2-mingw64-9-g++17"},
	"cpp:20-gnu10":    {"cpp.g++20", "cpp.gcc11-64-winlibs-g++20", "cpp.gcc13-64-winlibs-g++20"},
	"csharp:5-dotnet": {"csharp.mono"},
	"d:1-gdc":         {"d"},
	"go:1.20":         {"go"},
	"java:1.8":        {"java8"},
	"java:1.17":       {"java11"},
	"java:1.21":       {"java21"},
	"kotlin:1.7":      {"kotlin", "kotlin16", "kotlin17"},
	"kotlin:1.9":      {"kotlin19"},
	"pascal:3.2":      {"pas.dpr", "pas.fpc"},
	"php:7.4":         {"php.5"},
	"python:3-python": {"python.2", "python.3"},
	"python:3-pypy":   {"python.pypy2", "python.pypy3", "python.pypy3-64"},
	"ruby:2.4":        {"ruby", "ruby.2"},
	"rust:1.46":       {"rust"},
}

func SourceTypeToRuntime(t string) (string, bool) {
	for runtime, kinds := range runtimeMapping {
		for _, kind := range kinds {
			if kind == t {
				return runtime, true
			}
		}
	}

	return "", false
}

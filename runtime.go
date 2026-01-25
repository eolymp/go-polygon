package polygon

// LanguageMapping polygon to eolymp language name mapping
var LanguageMapping = map[string]string{
	"c":       "c",
	"cpp":     "cpp",
	"d":       "d",
	"csharp":  "csharp",
	"go":      "go",
	"java":    "java",
	"kotlin":  "kotlin",
	"pas":     "pascal",
	"php":     "php",
	"py":      "python",
	"python":  "python",
	"ruby":    "ruby",
	"rust":    "rust",
	"haskell": "haskell",
	"js":      "js",
	"lua":     "lua",
	"perl":    "perl",
	"swift":   "swift",
}

// ReverseLanguageMapping eolymp language to polygon name mapping
func ReverseLanguageMapping(t string) (string, bool) {
	if v, ok := LanguageMapping[t]; ok {
		return v, true
	}
	return "", false
}

// LanguageExtensions eolymp language to file extension
var LanguageExtensions = map[string]string{
	"c":       "c",
	"cpp":     "cpp",
	"d":       "d",
	"csharp":  "cs",
	"go":      "go",
	"java":    "java",
	"kotlin":  "kt",
	"pascal":  "pas",
	"php":     "php",
	"python":  "py",
	"ruby":    "rb",
	"rust":    "rs",
	"haskell": "hs",
	"js":      "js",
	"lua":     "lua",
	"perl":    "pl",
	"swift":   "swift",
}

// polygon to eolymp runtime name mapping
var RuntimeMapping = map[string]string{
	"c.gcc":                      "c:17-gnu10",
	"cpp.g++":                    "cpp:23-gnu14",
	"cpp.g++11":                  "cpp:17-gnu10",
	"cpp.g++14":                  "cpp:17-gnu10",
	"cpp.g++17":                  "cpp:17-gnu10",
	"cpp.ms":                     "cpp:23-gnu14",
	"cpp.msys2-mingw64-9-g++17":  "cpp:17-gnu10",
	"cpp.g++20":                  "cpp:20-gnu14",
	"cpp.gcc11-64-winlibs-g++20": "cpp:20-gnu14",
	"cpp.gcc13-64-winlibs-g++20": "cpp:20-gnu14",
	"cpp.gcc14-64-msys2-g++23":   "cpp:23-gnu14",
	"csharp.mono":                "csharp:5-dotnet",
	"d":                          "d:1-gdc",
	"go":                         "go:1.20",
	"java8":                      "java:1.8",
	"java11":                     "java:1.17",
	"java21":                     "java:1.21",
	"kotlin":                     "kotlin:1.7",
	"kotlin16":                   "kotlin:1.7",
	"kotlin17":                   "kotlin:1.7",
	"kotlin19":                   "kotlin:1.9",
	"pas.dpr":                    "pascal:3.2",
	"pas.fpc":                    "pascal:3.2",
	"php.5":                      "php:7.4",
	"python.2":                   "python:3.13-python",
	"python.3":                   "python:3.13-python",
	"python.pypy2":               "python:3.10-pypy",
	"python.pypy3":               "python:3.10-pypy",
	"python.pypy3-64":            "python:3.10-pypy",
	"ruby":                       "ruby:2.4",
	"ruby.2":                     "ruby:2.4",
	"rust":                       "rust:1.78",
}

// eolymp language mapping to runtimes for templates
// ie. what runtimes should template in given language be generated for
var TemplateMapping = map[string][]string{
	"c":       {"c:17-gnu10"},
	"cpp":     {"cpp:23-gnu14", "cpp:23-gnu14-extra"},
	"csharp":  {"csharp:5-dotnet", "csharp:5-mono"},
	"d":       {"d:1-dmd", "d:1-gdc"},
	"go":      {"go:1.20"},
	"haskell": {"haskell:8.8-ghc"},
	"java":    {"java:1.21", "java:1.25"},
	"js":      {"js:18"},
	"kotlin":  {"kotlin:1.9"},
	"lua":     {"lua:5.1"},
	"pascal":  {"pascal:3.2"},
	"perl":    {"perl:5.32"},
	"php":     {"php:7.4"},
	"python":  {"python:3.10-pypy", "python:3.10-pypy-extra", "python:3.13-python", "python:3.13-python-extra", "python:3.13-ai"},
	"ruby":    {"ruby:2.4"},
	"rust":    {"rust:1.78"},
	"swift":   {"swift:5.6"},
}

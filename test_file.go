package polygon

import (
	"os"
	"path/filepath"
	"sort"
)

type testFile struct {
	input  string
	output string
}

func findTestFiles(path string) ([]testFile, error) {
	files, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	inputExtensions := map[string]bool{"": true, ".in": true, ".dat": true}
	outputExtensions := map[string]bool{".a": true, ".out": true, ".sol": true, ".ans": true}

	inputs := map[string]string{}
	outputs := map[string]string{}

	for _, file := range files {
		extension := filepath.Ext(file.Name())
		filename := file.Name()[:len(file.Name())-len(extension)]
		dest := filepath.Join(path, file.Name())
		if inputExtensions[extension] {
			inputs[filename] = dest
		} else if outputExtensions[extension] {
			outputs[filename] = dest
		} else {
			inputs[file.Name()] = dest
		}
	}

	keys := make([]string, 0)
	for k, _ := range inputs {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	var tests []testFile
	for _, filename := range keys {
		inputName, ok1 := inputs[filename]
		outputName, ok2 := outputs[filename]
		if ok1 && ok2 {
			test := testFile{input: inputName, output: outputName}
			tests = append(tests, test)
		}
	}

	return tests, nil
}

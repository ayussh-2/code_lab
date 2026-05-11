//the per-language config used by the runner. Adding a new
package docker

// LangSpec describes how to compile and run code for one language inside the
// sandbox container.
type LangSpec struct {
	Image string

	FileName string

	NeedsCompile bool


	CompileCmd []string

	RunCmd []string
}


var Languages = map[string]LangSpec{
	"python":     {Image: "codelab/sandbox-python:latest", FileName: "main.py", RunCmd: []string{"python3", "main.py"}},
	"javascript": {Image: "codelab/sandbox-node:latest", FileName: "main.js", RunCmd: []string{"node", "main.js"}},
	"cpp": {Image: "codelab/sandbox-cpp:latest", FileName: "main.cpp", NeedsCompile: true,
		CompileCmd: []string{"sh", "-c", "g++ -O2 -std=c++17 main.cpp -o main"},
		RunCmd:     []string{"./main"}},
}

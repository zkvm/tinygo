package interp

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"tinygo.org/x/go-llvm"
)

func TestInterp(t *testing.T) {
	for _, name := range []string{
		"basic",
		"slice-copy",
	} {
		name := name // make tc local to this closure
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			runTest(t, "testdata/"+name)
		})
	}
}

func runTest(t *testing.T, pathPrefix string) {
	// Read the input IR.
	ctx := llvm.NewContext()
	buf, err := llvm.NewMemoryBufferFromFile(pathPrefix + ".ll")
	os.Stat(pathPrefix + ".ll") // make sure this file is tracked by `go test` caching
	if err != nil {
		t.Fatalf("could not read file %s: %v", pathPrefix+".ll", err)
	}
	mod, err := ctx.ParseIR(buf)
	if err != nil {
		t.Fatalf("could not load module:\n%v", err)
	}

	// Perform the transform.
	err = Run(mod, false)
	if err != nil {
		t.Fatal(err)
	}

	// Run some cleanup passes to get easy-to-read outputs.
	pm := llvm.NewPassManager()
	defer pm.Dispose()
	pm.AddGlobalOptimizerPass()
	pm.AddDeadStoreEliminationPass()
	pm.Run(mod)

	// Read the expected output IR.
	out, err := ioutil.ReadFile(pathPrefix + ".out.ll")
	if err != nil {
		t.Fatalf("could not read output file %s: %v", pathPrefix+".out.ll", err)
	}

	// See whether the transform output matches with the expected output IR.
	expected := string(out)
	actual := mod.String()
	if !fuzzyEqualIR(expected, actual) {
		t.Logf("output does not match expected output:\n%s", actual)
		t.Fail()
	}
}

// fuzzyEqualIR returns true if the two LLVM IR strings passed in are roughly
// equal. That means, only relevant lines are compared (excluding comments
// etc.).
func fuzzyEqualIR(s1, s2 string) bool {
	lines1 := filterIrrelevantIRLines(strings.Split(s1, "\n"))
	lines2 := filterIrrelevantIRLines(strings.Split(s2, "\n"))
	if len(lines1) != len(lines2) {
		return false
	}
	for i, line := range lines1 {
		if line != lines2[i] {
			return false
		}
	}

	return true
}

// filterIrrelevantIRLines removes lines from the input slice of strings that
// are not relevant in comparing IR. For example, empty lines and comments are
// stripped out.
func filterIrrelevantIRLines(lines []string) []string {
	var out []string
	for _, line := range lines {
		if line == "" || line[0] == ';' {
			continue
		}
		if strings.HasPrefix(line, "source_filename = ") {
			continue
		}
		out = append(out, line)
	}
	return out
}

package testutils

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/pkg/errors"

	"github.com/solo-io/solo-kit/pkg/utils/log"
	"github.com/solo-io/sqoop/cli/pkg/cmd"
)

func Sqoopctl(args string) error {
	app := cmd.App("test")
	app.SetArgs(strings.Split(args, " "))
	return app.Execute()
}
func SqoopctlOut(args string) (string, error) {
	stdOut := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		return "", err
	}
	os.Stdout = w

	app := cmd.App("test")
	app.SetArgs(strings.Split(args, " "))
	err = app.Execute()

	outC := make(chan string)

	// copy the output in a separate goroutine so printing can't block indefinitely
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf.String()
	}()

	// back to normal state
	w.Close()
	os.Stdout = stdOut // restoring the real stdout
	out := <-outC

	return strings.TrimSuffix(out, "\n"), nil
}

func Make(dir, args string) error {
	make := exec.Command("make", strings.Split(args, " ")...)
	make.Dir = dir
	out, err := make.CombinedOutput()
	if err != nil {
		return errors.Errorf("make failed with err: %s", out)
	}
	return nil
}

func MustWriteTestFile(contents string) string {
	tmpFile, err := ioutil.TempFile("", "test-")

	if err != nil {
		log.Fatalf("Failed to create test file: %v", err)
	}

	text := []byte(contents)
	if _, err = tmpFile.Write(text); err != nil {
		log.Fatalf("Failed to write to test file: %v", err)
	}

	if err := tmpFile.Close(); err != nil {
		log.Fatalf("Failed to write to test file: %v", err)
	}

	return tmpFile.Name()
}

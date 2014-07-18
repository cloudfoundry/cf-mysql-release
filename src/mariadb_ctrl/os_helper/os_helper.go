package os_helper


import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
)

type OsHelper interface {
	RunCommand(executable string, args ...string) (string, error)
	RunCommandPanicOnErr(executable string, args ...string) string
	FileExists(filename string) bool
	ReadFile(filename string) (string, error)
	WriteStringToFile(filename string, contents string) (error)
}

type OsHelperImpl struct {}

func NewImpl() *OsHelperImpl {
	return &OsHelperImpl {}
}

// Runs command with stdout and stderr pipes connected to process
func (h OsHelperImpl) RunCommand(executable string, args ...string) (string, error) {
	cmd := exec.Command(executable, args...)
	out, err := cmd.Output()
	if (err != nil) {
		return string(out), err
	}
	return string(out), nil
}

// Runs command with stdout and stderr pipes connected to process
func (h OsHelperImpl) RunCommandPanicOnErr(executable string, args ...string) string {
	cmd := exec.Command(executable, args...)
	out, err := cmd.Output()
	if (err != nil) {
		panic(fmt.Sprintf("Output: %s. Error: %v", string(out), err))
	}
	return string(out)
}

func (h OsHelperImpl) FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return (err == nil)
}

// Read the whole file, panic on err
func (h OsHelperImpl) ReadFile(filename string) (string, error) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return string(b[:]), nil
}

// Overwrite the contents, creating if necessary. Panic on err
func (h OsHelperImpl) WriteStringToFile(filename string, contents string) (error) {
	err := ioutil.WriteFile(filename, []byte(contents), 0644)
	return err
}

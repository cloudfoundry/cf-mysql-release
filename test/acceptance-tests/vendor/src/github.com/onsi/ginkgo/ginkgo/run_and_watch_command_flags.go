package main

import (
	"flag"
	"runtime"

	"github.com/onsi/ginkgo/config"
)

type RunAndWatchCommandFlags struct {
	NumCPU         int
	ParallelStream bool
	Recurse        bool
	Race           bool
	Cover          bool
	Notify         bool
	SkipPackage    string
	Tags           string
	AutoNodes      bool

	//not for watch command
	KeepGoing    bool
	UntilItFails bool

	FlagSet *flag.FlagSet
}

func NewRunCommandFlags(flagSet *flag.FlagSet) *RunAndWatchCommandFlags {
	c := &RunAndWatchCommandFlags{
		FlagSet: flagSet,
	}
	c.flags(false)
	return c
}

func NewWatchCommandFlags(flagSet *flag.FlagSet) *RunAndWatchCommandFlags {
	c := &RunAndWatchCommandFlags{
		FlagSet: flagSet,
	}
	c.flags(true)
	return c
}

func (c *RunAndWatchCommandFlags) wasSet(flagName string) bool {
	wasSet := false
	c.FlagSet.Visit(func(f *flag.Flag) {
		if f.Name == flagName {
			wasSet = true
		}
	})

	return wasSet
}

func (c *RunAndWatchCommandFlags) computeNodes() {
	if c.wasSet("nodes") {
		return
	}
	if c.AutoNodes {
		switch n := runtime.NumCPU(); {
		case n <= 4:
			c.NumCPU = n
		default:
			c.NumCPU = n - 1
		}
	}
}

func (c *RunAndWatchCommandFlags) flags(forWatchCommand bool) {
	onWindows := (runtime.GOOS == "windows")
	onOSX := (runtime.GOOS == "darwin")

	config.Flags(c.FlagSet, "", false)

	c.FlagSet.IntVar(&(c.NumCPU), "nodes", 1, "The number of parallel test nodes to run")
	c.FlagSet.BoolVar(&(c.AutoNodes), "p", false, "Run in parallel with auto-detected number of nodes")
	c.FlagSet.BoolVar(&(c.ParallelStream), "stream", onWindows, "stream parallel test output in real time: less coherent, but useful for debugging")
	c.FlagSet.BoolVar(&(c.Recurse), "r", false, "Find and run test suites under the current directory recursively")
	c.FlagSet.BoolVar(&(c.Race), "race", false, "Run tests with race detection enabled")
	c.FlagSet.BoolVar(&(c.Cover), "cover", false, "Run tests with coverage analysis, will generate coverage profiles with the package name in the current directory")
	c.FlagSet.StringVar(&(c.SkipPackage), "skipPackage", "", "Packages with names matching this regular expression will be skipped")
	if onOSX {
		c.FlagSet.BoolVar(&(c.Notify), "notify", false, "Send desktop notifications when a test run completes")
	}
	c.FlagSet.StringVar(&(c.Tags), "tags", "", "A list of build tags to consider satisfied during the build")
	if !forWatchCommand {
		c.FlagSet.BoolVar(&(c.KeepGoing), "keepGoing", false, "When true, failures from earlier test suites do not prevent later test suites from running")
		c.FlagSet.BoolVar(&(c.UntilItFails), "untilItFails", false, "When true, Ginkgo will keep rerunning tests until a failure occurs")
	}
}

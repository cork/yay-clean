package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/ogier/pflag"
)

type Config struct {
	CalculateSpace bool
	NotInstalled   bool
	NumberToKeep   int
	PackageFiles   bool
	PKGFolders     bool
	Quiet          bool
	Remove         bool
	SourceFiles    bool
	Verbose        bool
	RemovedFiles   int
	Space          int64
}

func NewConfig() *Config {
	calculateSpace := pflag.BoolP("saved-space", "c", false, "Calculate saved space.")
	notInstalled := pflag.BoolP("not-installed", "i", false, "Check for folders for packages not installed")
	numberToKeep := pflag.IntP("keep", "k", 3, "Number of build and source files to keep")
	packageFiles := pflag.BoolP("package-files", "p", false, "Check for old package files (pkg.tar.zst)")
	pkgFolders := pflag.BoolP("build-folders", "b", false, "Check for src and pkg folders")
	quiet := pflag.BoolP("quiet", "q", false, "Silens all output.")
	remove := pflag.BoolP("remove", "r", false, "Remove files and folders from the system")
	sourceFiles := pflag.BoolP("source-files", "s", false, "Check for source files")
	verbose := pflag.BoolP("verbose", "v", false, "Verbose output.")
	help := pflag.Bool("help", false, "display this help and exit")

	pflag.Parse()

	if *help {
		pflag.Usage()
		os.Exit(1)
	}

	return &Config{
		CalculateSpace: *calculateSpace,
		NotInstalled:   *notInstalled,
		NumberToKeep:   *numberToKeep,
		PackageFiles:   *packageFiles,
		PKGFolders:     *pkgFolders,
		Quiet:          *quiet,
		Remove:         *remove,
		SourceFiles:    *sourceFiles,
		Verbose:        *verbose,
	}
}

func (c *Config) Println(a ...any) (n int, err error) {
	c.RemovedFiles++
	if c.Verbose {
		return fmt.Println(a...)
	}

	return 0, nil
}

func (c *Config) CalculateTotalSize(basePath string) {
	if !c.CalculateSpace {
		return
	}

	file, err := os.Stat(basePath)
	if err != nil {
		return
	}

	if file.IsDir() {
		var size int64
		filepath.Walk(basePath, func(path string, info fs.FileInfo, err error) error {
			if !info.IsDir() {
				size += info.Size()
			}
			return nil
		})

		c.Space += size
		return
	}

	c.Space += file.Size()
}

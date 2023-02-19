package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"

	"github.com/dustin/go-humanize"
	"github.com/fvbommel/sortorder"
)

var sourceExts = map[string]struct{}{
	"asc": {},
	"bz2": {},
	"deb": {},
	"gz":  {},
	"jar": {},
	"rpm": {},
	"zip": {},
	"zst": {},
	"zx":  {},
}

func main() {
	config := NewConfig()

	base, err := os.UserHomeDir()
	if err != nil {
		panic(err.Error())
	}

	base = filepath.Join(base, ".cache/yay")
	prefixMatcher := regexp.MustCompile(`-\d+[.:]`)

	sources := make(map[string]map[string]map[string][]string)
	packages := make(map[string]map[string][]string)
	filepath.Walk(base, func(path string, info fs.FileInfo, err error) error {
		if path == base {
			return nil
		}

		if info.IsDir() {
			if config.NotInstalled {
				if ok, _ := filepath.Match(filepath.Join(base, "*"), path); ok {
					if !CheckInstalled(info.Name()) {
						config.CalculateTotalSize(path)

						if config.Remove {
							config.Println("Not installed, removing", path)
							os.RemoveAll(path)
						} else {
							config.Println("Not installed", path)
						}
						return filepath.SkipDir
					}
				}
			}

			if config.PKGFolders {
				if ok, _ := filepath.Match(filepath.Join(base, "*/*"), path); ok {
					if info.Name() == "src" || info.Name() == "pkg" {
						config.CalculateTotalSize(path)

						if config.Remove {
							config.Println("Removing", path)
							os.RemoveAll(path)
						} else {
							config.Println("Build or source folder found", path)
						}
						return filepath.SkipDir
					}
				}
			}

			if ok, _ := filepath.Match(filepath.Join(base, "*"), path); !ok {
				return filepath.SkipDir
			}
		} else {
			if ok, _ := filepath.Match(filepath.Join(base, "*/*"), path); ok {
				parent := filepath.Dir(path)
				name := info.Name()
				prefix := prefixMatcher.Split(name, 2)[0]

				if ok, _ := filepath.Match("*.tar.zst", info.Name()); ok {
					if packages[parent] == nil {
						packages[parent] = make(map[string][]string)
					}
					packages[parent][prefix] = append(packages[parent][prefix], info.Name())
				} else if _, ok := sourceExts[filepath.Ext(info.Name())]; ok {
					if sources[parent] == nil {
						sources[parent] = make(map[string]map[string][]string)
					}
					if sources[parent][prefix] == nil {
						sources[parent][prefix] = make(map[string][]string)
					}

					ext := filepath.Ext(name)
					sources[parent][prefix][ext] = append(sources[parent][prefix][ext], name)
				}
			}
		}

		return nil
	})

	if config.PackageFiles {
		for path, prefixes := range packages {
			for _, set := range prefixes {
				if len(set) > config.NumberToKeep {
					sort.Sort(sort.Reverse(sortorder.Natural(set)))

					for _, file := range set[config.NumberToKeep:] {
						config.CalculateTotalSize(filepath.Join(path, file))

						if config.Remove {
							config.Println("Removing:", filepath.Join(path, file))
							os.RemoveAll(filepath.Join(path, file))
						} else {
							config.Println("Exceeding old package:", filepath.Join(path, file))
						}
					}
				}
			}
		}
	}

	if config.SourceFiles {
		for path, prefixes := range sources {
			for _, ext := range prefixes {
				for _, set := range ext {
					if len(set) > config.NumberToKeep {
						sort.Sort(sort.Reverse(sortorder.Natural(set)))

						for _, file := range set[config.NumberToKeep:] {
							config.CalculateTotalSize(filepath.Join(path, file))

							if config.Remove {
								config.Println("Removing:", filepath.Join(path, file))
								os.RemoveAll(path)
							} else {
								config.Println("Exceeding old source file:", filepath.Join(path, file))
							}
						}
					}
				}
			}
		}
	}

	if !config.Quiet {
		size := ""
		if config.CalculateSpace {
			size = fmt.Sprintf(" (disk space saved: %s)", humanize.IBytes(uint64(config.Space)))
		}
		if config.Remove {
			fmt.Printf("\033[32m==>\033[0m\033[1m finished: %d files removed%s\033[0m\n", config.RemovedFiles, size)
		} else {
			fmt.Printf("\033[32m==>\033[0m\033[1m finished dry run: %d candidates%s\033[0m\n", config.RemovedFiles, size)
		}
	}
}

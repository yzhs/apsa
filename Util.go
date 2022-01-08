package apsa

import (
	"fmt"
	"os"
)

func LogError(err interface{}) {
	fmt.Fprintln(os.Stderr, err)
}

// TryLogError logs an error if `err != nil`
func TryLogError(err interface{}) {
	if err != nil {
		LogError(err)
	}
}

// Compute the combined size of all files in a given directory.
func getDirSize(dir string) (int, int64) {
	directory, err := os.Open(dir)
	TryLogError(err)
	defer directory.Close()
	fileInfo, err := directory.Readdir(0)
	if err != nil {
		panic(err)
	}
	result := int64(0)
	for _, file := range fileInfo {
		result += file.Size()
	}
	return len(fileInfo), result
}

// Get the time a given file was last modified as a Unix time.
func getModTime(file string) (int64, error) {
	info, err := os.Stat(file)
	if err != nil {
		return -1, err
	}
	return info.ModTime().Unix(), nil
}

// Cache the newest modification of the template files as a Unix time
// (i.e. seconds since 1970-01-01).
var templatesModTime int64 = -1

// All recognized template files
// TODO Generate the list⁈
var templateFiles = []string{"header.html", "footer.html"}

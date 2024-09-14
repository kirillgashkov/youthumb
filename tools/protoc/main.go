// protoc is a wrapper around the protoc command that installs the protoc binary
// and the protoc-gen-go and protoc-gen-go-grpc plugins and then runs the protoc
// command with the given arguments.
package main

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// protocGenGoPackage and protocGenGoGRPCPackage are the Go package paths
// for the protoc-gen-go and protoc-gen-go-grpc plugins, respectively.
// The versions of these plugins are determined by go.mod.
const (
	protocGenGoPackage     = "google.golang.org/protobuf/cmd/protoc-gen-go"
	protocGenGoGRPCPackage = "google.golang.org/grpc/cmd/protoc-gen-go-grpc"
)

var protocZipURL = func() string {
	const protocVersion = "27.2"

	var protocPlatform string
	const runtimePlatform = runtime.GOOS + "/" + runtime.GOARCH
	switch runtimePlatform {
	case "darwin/amd64":
		protocPlatform = "osx-x86_64"
	case "darwin/arm64":
		protocPlatform = "osx-aarch_64"
	case "linux/amd64":
		protocPlatform = "linux-x86_64"
	case "linux/arm64":
		protocPlatform = "linux-aarch_64"
	default:
		log.Fatalf("unknown platform: %s", runtimePlatform)
	}

	return fmt.Sprintf(
		"https://github.com/protocolbuffers/protobuf/releases/download/v%s/protoc-%s-%s.zip",
		protocVersion,
		protocVersion,
		protocPlatform,
	)
}()

// protocExitError is an error that is used to communicate the exit code of the protoc command.
// The code is not necessarily a non-zero value, it is the actual exit code of the protoc command.
type protocExitError struct {
	code int
}

func (e protocExitError) Error() string {
	return fmt.Sprintf("protoc exited with code %d", e.code)
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("protoc: ")

	var protocExitErr protocExitError
	if err := mainErr(); errors.As(err, &protocExitErr) {
		os.Exit(protocExitErr.code)
	} else {
		log.Fatal(err)
	}
}

// mainErr always returns a non-nil error.
// The exit code of the protoc command is returned via protocExitError.
func mainErr() error {
	// Create a temporary directory to store the protoc binary and the plugins.

	workDir, err := os.MkdirTemp("", "protoc")
	if err != nil {
		return err
	}
	defer func(x string) {
		if err := os.RemoveAll(x); err != nil {
			log.Print(err)
		}
	}(workDir)

	binDir := filepath.Join(workDir, "bin")
	includeDir := filepath.Join(workDir, "include")

	// Install the protoc binary and the protoc-gen-go and protoc-gen-go-grpc plugins.

	if err := installProtoc(binDir, includeDir); err != nil {
		return err
	}

	if err := installProtocGenGo(binDir); err != nil {
		return err
	}

	if err := installProtocGenGoGRPC(binDir); err != nil {
		return err
	}

	// Run the protoc command.

	args := os.Args[1:]

	pathWithBin := fmt.Sprintf("PATH=%s%c%s", binDir, filepath.ListSeparator, os.Getenv("PATH"))
	environ := os.Environ()
	environ = append(environ, pathWithBin)

	// The protoc command is executed as a subprocess because using syscall.Exec
	// would replace the current process with protoc, which would prevent the
	// deferred cleanup code from running.
	cmd := exec.Command(filepath.Join(binDir, "protoc"), args...)
	cmd.Env = environ
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return protocExitError{exitErr.ExitCode()}
		} else {
			return err
		}
	} else {
		return protocExitError{0}
	}
}

// installProtoc downloads the protoc binary and include files and installs them to the given directories.
func installProtoc(binDir string, includeDir string) error {
	// Download the protoc zip file.

	zipFile, err := os.CreateTemp("", "protoc-*.zip")
	if err != nil {
		return err
	}
	defer func(x *os.File) {
		if err := x.Close(); err != nil && !errors.Is(err, os.ErrClosed) {
			log.Print(err)
		}
		if err := os.Remove(x.Name()); err != nil {
			log.Print(err)
		}
	}(zipFile)

	if err := downloadURL(protocZipURL, zipFile); err != nil {
		return err
	}

	if err := zipFile.Close(); err != nil {
		return err
	}

	// Unzip the protoc binary.

	if err := os.MkdirAll(binDir, 0755); err != nil {
		return err
	}
	if err := unzipPath(zipFile.Name(), filepath.Join("bin", "protoc"), filepath.Join(binDir, "protoc")); err != nil {
		return err
	}
	if err := os.Chmod(filepath.Join(binDir, "protoc"), 0755); err != nil {
		return err
	}

	// Unzip the protoc include files.

	if err := os.MkdirAll(filepath.Dir(includeDir), 0755); err != nil {
		return err
	}
	if err := unzipPath(zipFile.Name(), "include", includeDir); err != nil {
		return err
	}

	return nil
}

// installProtocGenGo installs the protoc-gen-go plugin to the given directory.
// The plugin is built from the protoc-gen-go package.
func installProtocGenGo(binDir string) error {
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return err
	}

	if err := buildPackage(protocGenGoPackage, filepath.Join(binDir, "protoc-gen-go")); err != nil {
		return err
	}

	return nil
}

// installProtocGenGoGRPC installs the protoc-gen-go-grpc plugin to the given directory.
// The plugin is built from the protoc-gen-go-grpc package.
func installProtocGenGoGRPC(binDir string) error {
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return err
	}

	if err := buildPackage(protocGenGoGRPCPackage, filepath.Join(binDir, "protoc-gen-go-grpc")); err != nil {
		return err
	}

	return nil
}

// downloadURL downloads the file at the given URL and writes it to the given file.
func downloadURL(url string, file *os.File) error {
	r, err := http.Get(url)
	if err != nil {
		return err
	}
	defer func(x io.Closer) {
		if err := x.Close(); err != nil {
			log.Print(err)
		}
	}(r.Body)

	if r.StatusCode != http.StatusOK {
		return errors.New("download failed: " + r.Status)
	}

	if _, err := io.Copy(file, r.Body); err != nil {
		return err
	}

	return nil
}

// unzipPath extracts files from the given zip archive that have the given
// source path prefix and writes them to the given destination directory
// stripping the source path prefix.
func unzipPath(zipPath, src, dst string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer func(x *zip.ReadCloser) {
		if err := x.Close(); err != nil {
			log.Print(err)
		}
	}(r)

	for _, f := range r.File {
		if !strings.HasPrefix(f.Name, src) {
			continue
		}

		dstPath := filepath.Join(dst, f.Name[len(src):])
		if err := writeZipFile(f, dstPath); err != nil {
			return err
		}
	}

	return nil
}

// writeZipFile writes the contents of the given zip file to the given path.
func writeZipFile(srcFile *zip.File, dstPath string) error {
	if srcFile.FileInfo().IsDir() {
		if err := os.Mkdir(dstPath, 0755); err != nil {
			return err
		}
	} else {
		rc, err := srcFile.Open()
		if err != nil {
			return err
		}
		defer func(x io.Closer) {
			if err := x.Close(); err != nil {
				log.Print(err)
			}
		}(rc)

		wc, err := os.Create(dstPath)
		if err != nil {
			return err
		}
		defer func(x io.Closer) {
			if err := x.Close(); err != nil {
				log.Print(err)
			}
		}(wc)

		if _, err := io.Copy(wc, rc); err != nil {
			return err
		}
	}

	return nil
}

// buildPackage builds the Go package at the given path and writes the binary to the given destination.
func buildPackage(pkg, dst string) error {
	cmd := exec.Command("go", "build", "-o", dst, pkg)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

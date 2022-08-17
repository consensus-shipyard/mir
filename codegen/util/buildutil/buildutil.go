package buildutil

import (
	"fmt"
	"go/build"
	"os"
)

func GetSourceDirForPackage(pkgPath string) (string, error) {
	// Get the current working directory.
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("cannot get the working directory: %w", err)
	}

	// Find the package location.
	pkg, err := build.Import(pkgPath, wd, build.FindOnly)
	if err != nil {
		return "", fmt.Errorf("cannot find package %v: %w", pkgPath, err)
	}
	return pkg.Dir, nil
}

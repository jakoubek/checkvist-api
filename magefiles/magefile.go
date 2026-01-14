//go:build mage

// Package main provides Mage build targets for the checkvist-api module.
package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

// magefile.go contains build targets: Test, Coverage, Lint, Fmt, Check.

// Default target when running mage without arguments.
var Default = Test

// Test runs all tests with verbose output.
func Test() error {
	fmt.Println("Running tests...")
	return sh.RunV("go", "test", "-v", "./...")
}

// Coverage runs tests with coverage reporting.
func Coverage() error {
	fmt.Println("Running tests with coverage...")
	return sh.RunV("go", "test", "-coverprofile=coverage.out", "./...")
}

// Lint runs staticcheck for static analysis.
func Lint() error {
	fmt.Println("Running staticcheck...")
	if _, err := exec.LookPath("staticcheck"); err != nil {
		fmt.Println("staticcheck not found. Install with: go install honnef.co/go/tools/cmd/staticcheck@latest")
		os.Exit(1)
	}
	return sh.RunV("staticcheck", "./...")
}

// Fmt formats all Go source files.
func Fmt() error {
	fmt.Println("Formatting Go files...")
	return sh.RunV("gofmt", "-w", ".")
}

// Vet runs go vet for code analysis.
func Vet() error {
	fmt.Println("Running go vet...")
	return sh.RunV("go", "vet", "./...")
}

// Check runs all quality checks: fmt, vet, staticcheck, and tests.
func Check() {
	mg.SerialDeps(Fmt, Vet, Lint, Test)
}

// Clean removes build artifacts.
func Clean() error {
	fmt.Println("Cleaning build artifacts...")
	return sh.Rm("coverage.out")
}

package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/MoustaphaSaad/sabre-go/internal/compiler"
	"github.com/MoustaphaSaad/sabre-go/internal/compiler/spirv"
)

const commandUsageTemplate = `Usage: %s <command> [flags]

Commands:
  scan             scans the given file and prints tokens to stdout
                   "sabre scan <file>"
  test-scan        tests the scan phase against golden output
                   "sabre test-scan <test-data-dir>"
  parse-expr       parses an expression
                   "sabre parse-expr <file>"
  test-parse-expr  tests the expression parsing against golden output
                   "sabre test-parse-expr <test-data-dir>"
  parse-stmt       parses a statement
                   "sabre parse-stmt <file>"
  test-parse-stmt  tests the statement parsing against golden output
                   "sabre test-parse-stmt <test-data-dir>"
  parse-decl       parses a declaration
                   "sabre parse-decl <file>"
  test-parse-decl  tests the declaration parsing against golden output
                   "sabre test-parse-decl <test-data-dir>"
  check            type checks a program
                   "sabre check <file>"
  test-check       tests the type checking against golden output
                   "sabre test-check <test-data-dir>"
  spirv            emits SPIR-V bytecode
                   "sabre spirv <file>"
  test-spirv       tests the SPIR-V emission against golden output
                   "sabre test-spirv <test-data-dir>"
`

func helpString() string {
	return fmt.Sprintf(commandUsageTemplate, os.Args[0])
}

func help() {
	fmt.Fprint(os.Stderr, helpString())
}

func cleanString(s string) string {
	return strings.TrimSpace(strings.ReplaceAll(s, "\r\n", "\n"))
}

func scan(args []string, out io.Writer) error {
	if len(args) < 1 {
		return fmt.Errorf("no file provided\n%v", helpString())
	}

	file := filepath.ToSlash(filepath.Clean(args[0]))
	unit, err := compiler.UnitFromFile(file)
	if err != nil {
		return fmt.Errorf("failed to create unit from file '%s': %v", file, err)
	}

	if !unit.Scan() {
		unit.PrintErrors(out)
		return nil
	}

	for _, token := range unit.RootFile().Tokens() {
		fmt.Fprintf(out, "%-15s %-20s %4d:%-4d %4d:%-4d [%d-%d]\n",
			token.Kind().String(),
			strconv.Quote(token.Value()),
			token.SourceRange().BeginPosition.Line,
			token.SourceRange().BeginPosition.Column,
			token.SourceRange().EndPosition.Line,
			token.SourceRange().EndPosition.Column,
			token.SourceRange().BeginOffset,
			token.SourceRange().EndOffset)
	}
	return nil
}

func testFunc(f func([]string, io.Writer) error, args []string, out io.Writer) error {
	flagSet := flag.NewFlagSet("test-scan", flag.ContinueOnError)
	update := flagSet.Bool("update", false, "updates test outputs")
	err := flagSet.Parse(args)
	if err != nil {
		return err
	}

	if flagSet.NArg() < 1 {
		return fmt.Errorf("no test data directory provided\n%v", helpString())
	}

	dir := flagSet.Arg(0)
	var goldenFiles []string
	err = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() && filepath.Ext(path) == ".golden" {
			goldenFiles = append(goldenFiles, path)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to walk test data directory: %v", err)
	}

	var failedTests []string

	for i, goldenFile := range goldenFiles {
		testFile := strings.TrimSuffix(goldenFile, filepath.Ext(goldenFile))
		expectedBytes, err := os.ReadFile(goldenFile)
		if err != nil {
			return fmt.Errorf("failed to read test file '%v': %v", testFile, err)
		}
		expectedOutput := cleanString(string(expectedBytes))

		fmt.Fprintf(out, "%v/%v) testing %v\n", i, len(goldenFiles), testFile)

		var actualOutputBuffer bytes.Buffer
		err = f([]string{testFile}, &actualOutputBuffer)
		if err != nil {
			return err
		}
		actualOutput := cleanString(actualOutputBuffer.String())

		if expectedOutput != actualOutput {
			if *update {
				f, err := os.Create(goldenFile)
				if err != nil {
					return err
				}
				defer f.Close()

				fmt.Fprintf(f, "%s\n", actualOutput)
				fmt.Fprintln(out, "UPDATED")
			} else {
				fmt.Fprintln(out, "FAILURE")
				failedTests = append(failedTests, testFile)
			}
		} else {
			fmt.Fprintln(out, "SUCCESS")
		}
	}

	if len(failedTests) > 0 {
		return fmt.Errorf("some test file failed, %v", failedTests)
	}
	return nil
}

func parseExpr(args []string, out io.Writer) error {
	if len(args) < 1 {
		return fmt.Errorf("no file provided\n%v", helpString())
	}

	file := filepath.ToSlash(filepath.Clean(args[0]))
	unit, err := compiler.UnitFromFile(file)
	if err != nil {
		return fmt.Errorf("failed to create unit from file '%s': %v", file, err)
	}

	if !unit.Scan() {
		unit.PrintErrors(out)
		return nil
	}

	parser := compiler.NewParser(unit.RootFile())
	expr := parser.ParseExpr()
	if expr != nil && !unit.HasErrors() {
		expr.Visit(compiler.NewASTPrinter(out))
	} else {
		unit.PrintErrors(out)
	}
	return nil
}

func parseStmt(args []string, out io.Writer) error {
	if len(args) < 1 {
		return fmt.Errorf("no file provided\n%v", helpString())
	}

	file := filepath.ToSlash(filepath.Clean(args[0]))
	unit, err := compiler.UnitFromFile(file)
	if err != nil {
		return fmt.Errorf("failed to create unit from file '%s': %v", file, err)
	}

	if !unit.Scan() {
		unit.PrintErrors(out)
		return nil
	}

	parser := compiler.NewParser(unit.RootFile())
	stmt := parser.ParseStmt()
	if stmt != nil && !unit.HasErrors() {
		stmt.Visit(compiler.NewASTPrinter(out))
	} else {
		unit.PrintErrors(out)
	}
	return nil
}

func parseDecl(args []string, out io.Writer) error {
	if len(args) < 1 {
		return fmt.Errorf("no file provided\n%v", helpString())
	}

	file := filepath.ToSlash(filepath.Clean(args[0]))
	unit, err := compiler.UnitFromFile(file)
	if err != nil {
		return fmt.Errorf("failed to create unit from file '%s': %v", file, err)
	}

	if !unit.Scan() {
		unit.PrintErrors(out)
		return nil
	}

	parser := compiler.NewParser(unit.RootFile())
	decl := parser.ParseDecl()
	if decl != nil && !unit.HasErrors() {
		decl.Visit(compiler.NewASTPrinter(out))
	} else {
		unit.PrintErrors(out)
	}
	return nil
}

func check(args []string, out io.Writer) error {
	if len(args) < 1 {
		return fmt.Errorf("no file provided\n%v", helpString())
	}

	file := filepath.ToSlash(filepath.Clean(args[0]))
	unit, err := compiler.UnitFromFile(file)
	if err != nil {
		return fmt.Errorf("failed to create unit from file '%s': %v", file, err)
	}

	if !unit.Scan() {
		unit.PrintErrors(out)
		return nil
	}

	if !unit.Parse() {
		unit.PrintErrors(out)
		return nil
	}

	if !unit.Check() {
		unit.PrintErrors(out)
	}

	return nil
}

func emitSPIRV(args []string, out io.Writer) error {
	if len(args) < 1 {
		return fmt.Errorf("no file provided\n%v", helpString())
	}

	file := filepath.ToSlash(filepath.Clean(args[0]))
	unit, err := compiler.UnitFromFile(file)
	if err != nil {
		return fmt.Errorf("failed to create unit from file '%s': %v", file, err)
	}

	if !unit.Scan() {
		unit.PrintErrors(out)
		return nil
	}

	if !unit.Parse() {
		unit.PrintErrors(out)
		return nil
	}

	if !unit.Check() {
		unit.PrintErrors(out)
	}

	module := unit.EmitSPIRV()
	printer := spirv.NewTextPrinter(out, module)
	printer.Print()

	return nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Error: no command found\n")
		help()
		return
	}

	subArgs := os.Args[2:]
	var err error
	switch os.Args[1] {
	case "help":
		help()
	case "scan":
		err = scan(subArgs, os.Stdout)
	case "test-scan":
		err = testFunc(scan, subArgs, os.Stdout)
	case "parse-expr":
		err = parseExpr(subArgs, os.Stdout)
	case "test-parse-expr":
		err = testFunc(parseExpr, subArgs, os.Stdout)
	case "parse-stmt":
		err = parseStmt(subArgs, os.Stdout)
	case "test-parse-stmt":
		err = testFunc(parseStmt, subArgs, os.Stdout)
	case "parse-decl":
		err = parseDecl(subArgs, os.Stdout)
	case "test-parse-decl":
		err = testFunc(parseDecl, subArgs, os.Stdout)
	case "check":
		err = check(subArgs, os.Stdout)
	case "test-check":
		err = testFunc(check, subArgs, os.Stdout)
	case "spirv":
		err = emitSPIRV(subArgs, os.Stdout)
	case "test-spirv":
		err = testFunc(emitSPIRV, subArgs, os.Stdout)
	default:
		fmt.Fprintf(os.Stderr, "Error: unknown command '%s'\n", os.Args[1])
		help()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

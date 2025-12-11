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
                   "sabre test-check <test-data-dir>
  spirv            emits SPIR-V bytecode in text
                   "sabre spirv <file>"
  test-spirv       tests the SPIR-V emission against golden output
                   "sabre test-spirv <test-data-dir>"
  spirv-bin        emits SPIR-V bytecode in binary
                   "sabre spirv-bin <file>"
  test-spirv-bin   tests the SPIR-V emission against golden binary output
                   "sabre test-spirv-bin <test-data-dir>"
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

func compareOutput(a, b []byte, outputIsBinary bool) int {
	if outputIsBinary {
		return bytes.Compare(a, b)
	} else {
		return strings.Compare(cleanString(string(a)), cleanString(string(b)))
	}
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

func testFunc(f func([]string, io.Writer) error, args []string, out io.Writer, outputExt string, outputIsBinary bool) error {
	flagSet := flag.NewFlagSet("test-func", flag.ContinueOnError)
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
		if !d.IsDir() && strings.HasSuffix(path, outputExt) {
			goldenFiles = append(goldenFiles, path)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to walk test data directory: %v", err)
	}

	var failedTests []string

	for i, goldenFile := range goldenFiles {
		testFile := strings.TrimSuffix(goldenFile, outputExt)
		expectedBytes, err := os.ReadFile(goldenFile)
		if err != nil {
			return fmt.Errorf("failed to read test file '%v': %v", testFile, err)
		}
		expectedOutputBytes := expectedBytes

		fmt.Fprintf(out, "%v/%v) testing %v\n", i, len(goldenFiles), testFile)

		var actualOutputBuffer bytes.Buffer
		err = f([]string{testFile}, &actualOutputBuffer)
		if err != nil {
			return err
		}
		actualOutputBytes := actualOutputBuffer.Bytes()

		if compareOutput(expectedOutputBytes, actualOutputBytes, outputIsBinary) != 0 {
			if *update {
				f, err := os.Create(goldenFile)
				if err != nil {
					return err
				}
				defer f.Close()

				if outputIsBinary {
					io.Copy(f, bytes.NewReader(actualOutputBytes))
				} else {
					fmt.Fprintf(f, "%s\n", string(actualOutputBytes))
				}
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

func emitSPIRV(args []string, out io.Writer, binary bool) error {
	flagSet := flag.NewFlagSet("emit-spirv", flag.ContinueOnError)
	err := flagSet.Parse(args)
	if err != nil {
		return err
	}

	args = flagSet.Args()
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
	if binary {
		printer := spirv.NewBinaryPrinter(out, module)
		printer.Emit()
	} else {
		printer := spirv.NewTextPrinter(out, module)
		printer.Emit()
	}

	return nil
}

func emitSPIRVText(args []string, out io.Writer) error {
	return emitSPIRV(args, out, false)
}

func emitSPIRVBin(args []string, out io.Writer) error {
	return emitSPIRV(args, out, true)
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
		err = testFunc(scan, subArgs, os.Stdout, ".golden", false)
	case "parse-expr":
		err = parseExpr(subArgs, os.Stdout)
	case "test-parse-expr":
		err = testFunc(parseExpr, subArgs, os.Stdout, ".golden", false)
	case "parse-stmt":
		err = parseStmt(subArgs, os.Stdout)
	case "test-parse-stmt":
		err = testFunc(parseStmt, subArgs, os.Stdout, ".golden", false)
	case "parse-decl":
		err = parseDecl(subArgs, os.Stdout)
	case "test-parse-decl":
		err = testFunc(parseDecl, subArgs, os.Stdout, ".golden", false)
	case "check":
		err = check(subArgs, os.Stdout)
	case "test-check":
		err = testFunc(check, subArgs, os.Stdout, ".golden", false)
	case "spirv":
		err = emitSPIRVText(subArgs, os.Stdout)
	case "test-spirv":
		err = testFunc(emitSPIRVText, subArgs, os.Stdout, ".golden", false)
	case "spirv-bin":
		err = emitSPIRVBin(subArgs, os.Stdout)
	case "test-spirv-bin":
		err = testFunc(emitSPIRVBin, subArgs, os.Stdout, ".golden.bin", true)
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

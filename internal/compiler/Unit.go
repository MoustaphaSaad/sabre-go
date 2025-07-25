package compiler

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type UnitFile struct {
	path         string
	absolutePath string
	content      string
	lines        []string
	tokens       []Token
	errors       []Error
}

func UnitFileFromFile(path string) (unitFile *UnitFile, err error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	contentStr := strings.ReplaceAll(string(content), "\r\n", "\n")

	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	unitFile = &UnitFile{
		path:         path,
		absolutePath: absolutePath,
		content:      contentStr,
		lines:        strings.Split(contentStr, "\n"),
	}

	return
}

func (u *UnitFile) error(e Error) {
	u.errors = append(u.errors, e)
}

func (u *UnitFile) errorf(sourceRange SourceRange, format string, a ...any) {
	u.error(Error{
		SourceRange: sourceRange,
		Message:     fmt.Sprintf(format, a...),
	})
}

func (u *UnitFile) HasErrors() bool {
	return len(u.errors) > 0
}

func (u *UnitFile) Scan() bool {
	scanner := NewScanner(u)
	for {
		token := scanner.Scan()
		u.tokens = append(u.tokens, token)
		if token.Kind() == TokenEOF || token.Kind() == TokenInvalid {
			break
		}
	}
	return !u.HasErrors()
}

func (u *UnitFile) Tokens() []Token {
	return u.tokens
}

type CompilationStage int

const (
	CompilationStageStart CompilationStage = iota
	CompilationStageScanned
	CompilationStageFailure
)

type Unit struct {
	compilationStage CompilationStage
	rootFile         *UnitFile
}

func UnitFromFile(path string) (unit *Unit, err error) {
	unitFile, err := UnitFileFromFile(path)
	if err != nil {
		return nil, err
	}

	unit = &Unit{
		compilationStage: CompilationStageStart,
		rootFile:         unitFile,
	}
	return
}

// RootFile returns the root file of the unit
func (u *Unit) RootFile() *UnitFile {
	return u.rootFile
}

func (u *Unit) HasErrors() bool {
	return len(u.rootFile.errors) > 0
}

func (u *Unit) PrintErrors(w io.Writer) {
	for _, e := range u.rootFile.errors {
		fmt.Fprintln(w, e)
	}
}

func (u *Unit) Scan() bool {
	if u.compilationStage == CompilationStageStart {
		if u.rootFile.Scan() {
			u.compilationStage = CompilationStageScanned
			return true
		} else {
			u.compilationStage = CompilationStageFailure
			return false
		}
	}
	return !u.HasErrors()
}

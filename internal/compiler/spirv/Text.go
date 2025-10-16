package spirv

import (
	"fmt"
	"io"
	"sort"
)

type TextPrinter struct {
	out    io.Writer
	module *Module
}

func NewTextPrinter(out io.Writer, module *Module) *TextPrinter {
	return &TextPrinter{
		out:    out,
		module: module,
	}
}

func (tp *TextPrinter) Print() {
	tp.PrintCapabilities()
	tp.PrintMemoryModel()
}

func (tp *TextPrinter) PrintCapabilities() {
	for _, cap := range tp.module.Capabilities() {
		fmt.Fprintf(tp.out, "OpCapability %s\n", cap.String())
	}

	// get the objects sorted by their IDs (dependency order)
	objects := make([]Object, 0, len(tp.module.objectsByID))
	for _, obj := range tp.module.objectsByID {
		objects = append(objects, obj)
	}
	sort.Slice(objects, func(i, j int) bool {
		return objects[i].ID() < objects[j].ID()
	})

	for _, obj := range objects {
		tp.PrintObject(obj)
	}
}

func (tp *TextPrinter) PrintMemoryModel() {
	fmt.Fprintf(tp.out, "OpMemoryModel %s %s\n", tp.module.AddressingModel.String(), tp.module.MemoryModel.String())
}

func (tp *TextPrinter) PrintObject(obj Object) {
	// to be implemented later
}

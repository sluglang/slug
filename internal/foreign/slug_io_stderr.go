package foreign

import (
	"bytes"
	"slug/internal/object"
)

func fnIoStderrPrint() *object.Foreign {
	return &object.Foreign{
		Name: "print",
		Fn: func(ctx object.RuntimeContext, args ...object.Object) object.Object {
			var out bytes.Buffer
			for i, arg := range args {
				out.WriteString(arg.Inspect())
				if i < len(args)-1 {
					out.WriteString(" ")
				}
			}
			print(out.String())
			if len(args) > 0 {
				return args[0]
			}
			return ctx.Nil()
		},
	}
}

func fnIoStderrPrintLn() *object.Foreign {
	return &object.Foreign{
		Name: "println",
		Fn: func(ctx object.RuntimeContext, args ...object.Object) object.Object {
			var out bytes.Buffer
			for i, arg := range args {
				out.WriteString(arg.Inspect())
				if i < len(args)-1 {
					out.WriteString(" ")
				}
			}
			println(out.String())
			if len(args) > 0 {
				return args[0]
			}
			return ctx.Nil()
		},
	}
}

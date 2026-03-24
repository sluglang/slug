package foreign

import (
	"path/filepath"
	"slug/internal/object"
)

func fnPathJoin() *object.Foreign {
	return &object.Foreign{
		Name: "join",
		Fn: func(ctx object.EvaluatorContext, args ...object.Object) object.Object {
			if len(args) < 1 {
				return ctx.NewError("join expects at least 1 argument")
			}

			parts := make([]string, len(args))
			for i, arg := range args {
				part, err := unpackString(arg, "part")
				if err != nil {
					return ctx.NewError("join argument %d must be a string", i+1)
				}
				parts[i] = part
			}

			return &object.String{Value: filepath.Join(parts...)}
		},
	}
}

func fnPathAbs() *object.Foreign {
	return &object.Foreign{
		Name: "abs",
		Fn: func(ctx object.EvaluatorContext, args ...object.Object) object.Object {
			if len(args) != 1 {
				return ctx.NewError("abs expects exactly 1 argument, got=%d", len(args))
			}

			path, err := unpackString(args[0], "path")
			if err != nil {
				return ctx.NewError(err.Error())
			}

			absPath, err := filepath.Abs(path)
			if err != nil {
				return ctx.NewError("failed to resolve absolute path: %s", err.Error())
			}

			return &object.String{Value: filepath.Clean(absPath)}
		},
	}
}

func fnPathModuleDir() *object.Foreign {
	return &object.Foreign{
		Name: "moduleDir",
		Fn: func(ctx object.EvaluatorContext, args ...object.Object) object.Object {
			if len(args) != 0 {
				return ctx.NewError("moduleDir expects no arguments, got=%d", len(args))
			}
			dir := filepath.Dir(ctx.CurrentEnv().Path)
			return &object.String{Value: filepath.Clean(dir)}
		},
	}
}

func fnPathCwd() *object.Foreign {
	return &object.Foreign{
		Name: "cwd",
		Fn: func(ctx object.EvaluatorContext, args ...object.Object) object.Object {
			if len(args) != 0 {
				return ctx.NewError("cwd expects no arguments, got=%d", len(args))
			}
			return &object.String{Value: ctx.GetConfiguration().Cwd}
		},
	}
}

func fnPathProjectRoot() *object.Foreign {
	return &object.Foreign{
		Name: "projectRoot",
		Fn: func(ctx object.EvaluatorContext, args ...object.Object) object.Object {
			if len(args) != 0 {
				return ctx.NewError("projectRoot expects no arguments, got=%d", len(args))
			}
			return &object.String{Value: ctx.GetConfiguration().ProjectRoot}
		},
	}
}

func fnPathLibRoot() *object.Foreign {
	return &object.Foreign{
		Name: "libRoot",
		Fn: func(ctx object.EvaluatorContext, args ...object.Object) object.Object {
			if len(args) != 0 {
				return ctx.NewError("libRoot expects no arguments, got=%d", len(args))
			}
			return &object.String{Value: ctx.CurrentEnv().LibRoot}
		},
	}
}

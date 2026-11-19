// Package call invokes functions and methods through reflection with loosely
// typed arguments.
//
// Each argument is converted to the parameter type that the function
// declares, using typeconv when possible and reflect conversion otherwise.
// Extra arguments beyond the declared parameters are ignored, which lets
// callers such as the event dispatcher pass a fixed argument list to
// listeners that only need some of it.
package call

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/olegshs/go-tools/helpers/typeconv"
)

var (
	// ErrInvalidMethod is returned when the method does not exist or cannot be
	// called.
	ErrInvalidMethod = errors.New("invalid method")

	// ErrIsNotFunc is returned when the function argument is not a function.
	ErrIsNotFunc = errors.New("is not a function")

	// ErrNotEnoughArguments is returned when fewer arguments are passed than the
	// function declares.
	ErrNotEnoughArguments = errors.New("not enough arguments")

	// ErrInvalidArgument is returned when an argument cannot be converted to the
	// declared parameter type.
	ErrInvalidArgument = errors.New("invalid argument")
)

// Method looks up the named method on object, calls it with args and returns
// its results. It returns ErrInvalidMethod, ErrNotEnoughArguments or
// ErrInvalidArgument on failure.
func Method(object interface{}, method string, args ...interface{}) ([]interface{}, error) {
	f := reflect.ValueOf(object).MethodByName(method)
	if !f.IsValid() {
		err := fmt.Errorf("%w: %T.%s", ErrInvalidMethod, object, method)
		return nil, err
	}

	return Func(f, args...)
}

// Func calls function, which may be a function value or its reflect.Value,
// with args and returns its results. It returns ErrIsNotFunc,
// ErrNotEnoughArguments or ErrInvalidArgument on failure.
func Func(function interface{}, args ...interface{}) ([]interface{}, error) {
	f, t, err := funcReflect(function)
	if err != nil {
		return nil, err
	}

	numArgs := len(args)
	numIn := t.NumIn()
	if numArgs < numIn {
		err := fmt.Errorf("%w: %d < %d", ErrNotEnoughArguments, numArgs, numIn)
		return nil, err
	}

	in, err := funcArgs(t, args...)
	if err != nil {
		return nil, err
	}

	out := f.Call(in)
	res := make([]interface{}, len(out))
	for i, v := range out {
		res[i] = v.Interface()
	}

	return res, nil
}

// funcReflect returns the reflect.Value and type of function, accepting
// either a function or a reflect.Value that holds one.
func funcReflect(function interface{}) (*reflect.Value, reflect.Type, error) {
	var f reflect.Value
	if t, ok := function.(reflect.Value); ok {
		f = t
	} else {
		f = reflect.ValueOf(function)
	}
	if !f.IsValid() {
		return nil, nil, ErrIsNotFunc
	}

	t := f.Type()

	if f.Kind() != reflect.Func {
		err := fmt.Errorf("%w: %s", ErrIsNotFunc, t)
		return nil, nil, err
	}

	return &f, t, nil
}

func funcArgs(t reflect.Type, args ...interface{}) ([]reflect.Value, error) {
	if t.IsVariadic() {
		return argsVariadic(t, args...)
	} else {
		return argsNonVariadic(t, args...)
	}
}

// argsVariadic converts all args. Arguments at or after the variadic
// parameter are converted to its element type.
func argsVariadic(t reflect.Type, args ...interface{}) ([]reflect.Value, error) {
	numArgs := len(args)
	numIn := t.NumIn()
	lastIn := numIn - 1
	in := make([]reflect.Value, numArgs)

	for i, v := range args {
		var argType reflect.Type
		if i >= lastIn {
			argType = t.In(lastIn).Elem()
		} else {
			argType = t.In(i)
		}

		val, err := convert(argType, v)
		if err != nil {
			return nil, err
		}

		in[i] = val
	}

	return in, nil
}

// argsNonVariadic converts the first t.NumIn() args and drops the rest.
func argsNonVariadic(t reflect.Type, args ...interface{}) ([]reflect.Value, error) {
	numIn := t.NumIn()
	in := make([]reflect.Value, numIn)

	for i := 0; i < numIn; i++ {
		v := args[i]
		argType := t.In(i)

		val, err := convert(argType, v)
		if err != nil {
			return nil, err
		}

		in[i] = val
	}

	return in, nil
}

// convert converts v to type t. A nil v becomes the zero value of t.
// Values of the same kind are passed through unchanged.
func convert(t reflect.Type, v interface{}) (reflect.Value, error) {
	if v == nil {
		return reflect.New(t).Elem(), nil
	}

	val := reflect.ValueOf(v)
	if val.Kind() == t.Kind() {
		return val, nil
	}

	tName := t.String()
	if !typeconv.CanTo(tName) {
		valType := val.Type()
		if !valType.ConvertibleTo(t) {
			err := fmt.Errorf("%w: cannot convert %s to %s", ErrInvalidArgument, valType, tName)
			return reflect.Zero(t), err
		}

		return val.Convert(t), nil
	}

	val = reflect.ValueOf(typeconv.To(tName, v))
	return val, nil
}

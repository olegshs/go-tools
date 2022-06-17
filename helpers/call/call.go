// Пакет call предоставляет функции для вызова методов и функций.
//
// Переданные аргументы преобразуются к типам, объявленным в этих функциях.
// Если переданных аргументов больше чем в объявлении функции, то лишние аргументы игнорируются.
package call

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/olegshs/go-tools/helpers/typeconv"
)

var (
	// ErrInvalidMethod возвращается если метод отсутствует или не может быть вызван.
	ErrInvalidMethod = errors.New("invalid method")

	// ErrIsNotFunc возвращается если аргумент function не является функцией.
	ErrIsNotFunc = errors.New("is not a function")

	// ErrNotEnoughArguments возвращается если число переданных аргументов меньше
	// числа аргументов в объявлении функции.
	ErrNotEnoughArguments = errors.New("not enough arguments")

	// ErrInvalidArgument возвращается если аргумент не может быть преобразован к нужному типу.
	ErrInvalidArgument = errors.New("invalid argument")
)

// Method находит метод по его имени в переданном объекте, затем вызывает этот метод
// и возвращает результаты вызова или одну из ошибок:
// ErrInvalidMethod, ErrNotEnoughArguments, ErrInvalidArgument.
func Method(object interface{}, method string, args ...interface{}) ([]interface{}, error) {
	f := reflect.ValueOf(object).MethodByName(method)
	if !f.IsValid() {
		err := fmt.Errorf("%w: %T.%s", ErrInvalidMethod, object, method)
		return nil, err
	}

	return Func(f, args...)
}

// Func вызывает переданную функцию или reflect.Value этой функции
// и возвращает результаты вызова или одну из ошибок:
// ErrIsNotFunc, ErrNotEnoughArguments, ErrInvalidArgument.
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

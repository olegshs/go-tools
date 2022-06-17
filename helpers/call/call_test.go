package call

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"testing"
)

type TestType struct {
	Name string
}

func (*TestType) Hello(name string) string {
	return fmt.Sprintf("Hello, %s!", name)
}

func (*TestType) HelloAll(names ...string) string {
	return fmt.Sprintf("Hello: %s!", strings.Join(names, ", "))
}

func (*TestType) TestName(t TestType) string {
	return t.Name
}

func ExampleMethod() {
	obj := new(TestType)

	res, err := Method(obj, "Hello", "world")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(res[0].(string))
}

func ExampleFunc() {
	f := func(name string) string {
		return fmt.Sprintf("Hello, %s!", name)
	}

	res, err := Func(f, "world")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(res[0].(string))
}

func TestMethod(t *testing.T) {
	obj := new(TestType)

	res, err := Method(obj, "Hello", "world")
	if err != nil {
		t.Fatal(err)
	}

	if s, ok := res[0].(string); ok {
		if s != "Hello, world!" {
			t.Errorf("unexpected result: %s", s)
		}
	} else {
		t.Errorf("invalid result")
	}
}

func TestMethodVariadic(t *testing.T) {
	obj := new(TestType)

	res, err := Method(obj, "HelloAll", 1, 2, 3)
	if err != nil {
		t.Fatal(err)
	}

	if s, ok := res[0].(string); ok {
		if s != "Hello: 1, 2, 3!" {
			t.Errorf("unexpected result: %s", s)
		}
	} else {
		t.Errorf("invalid result")
	}
}

func TestErrors(t *testing.T) {
	obj := new(TestType)

	_, err := Method(obj, "Test")
	if !errors.Is(err, ErrInvalidMethod) {
		t.Errorf("invalid error: %v\nexpected: %v", err, ErrInvalidMethod)
	}

	_, err = Func(0)
	if !errors.Is(err, ErrIsNotFunc) {
		t.Errorf("invalid error: %v\nexpected: %v", err, ErrIsNotFunc)
	}

	_, err = Method(obj, "Hello")
	if !errors.Is(err, ErrNotEnoughArguments) {
		t.Errorf("invalid error: %v\nexpected: %v", err, ErrNotEnoughArguments)
	}

	_, err = Method(obj, "TestName", 0)
	if !errors.Is(err, ErrInvalidArgument) {
		t.Errorf("invalid error: %v\nexpected: %v", err, ErrInvalidArgument)
	}
}

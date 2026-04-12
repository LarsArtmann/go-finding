package finding

import (
	"errors"
	"testing"
)

func TestResultOk(t *testing.T) {
	r := Ok(42)
	if !r.IsOk() {
		t.Error("expected IsOk to be true")
	}

	if r.IsErr() {
		t.Error("expected IsErr to be false")
	}

	if r.Value() != 42 {
		t.Errorf("expected value 42, got %v", r.Value())
	}
}

func TestResultErr(t *testing.T) {
	err := errors.New("something went wrong")

	r := Err[int](err)
	if r.IsOk() {
		t.Error("expected IsOk to be false")
	}

	if !r.IsErr() {
		t.Error("expected IsErr to be true")
	}

	if !errors.Is(r.Error(), err) {
		t.Error("expected error to match")
	}
}

func TestResultValueOr(t *testing.T) {
	ok := Ok(42)
	if v := ok.ValueOr(0); v != 42 {
		t.Errorf("expected 42, got %v", v)
	}

	err := Err[int](errors.New("fail"))
	if v := err.ValueOr(0); v != 0 {
		t.Errorf("expected 0, got %v", v)
	}
}

func TestResultValueOrElse(t *testing.T) {
	ok := Ok(42)
	if v := ok.ValueOrElse(func(e error) int { return 0 }); v != 42 {
		t.Errorf("expected 42, got %v", v)
	}

	err := Err[int](errors.New("fail"))
	if v := err.ValueOrElse(func(e error) int { return 100 }); v != 100 {
		t.Errorf("expected 100, got %v", v)
	}
}

func TestResultUnwrap(t *testing.T) {
	ok := Ok(42)

	val, err := ok.Unwrap()
	if err != nil {
		t.Error("expected no error")
	}

	if val != 42 {
		t.Errorf("expected 42, got %v", val)
	}

	errRes := Err[int](errors.New("fail"))

	val, err = errRes.Unwrap()
	if err == nil {
		t.Error("expected error")
	}

	if val != 0 {
		t.Errorf("expected zero value, got %v", val)
	}
}

func TestResultMap(t *testing.T) {
	ok := Ok(5)

	doubled := ok.Map(func(x int) int { return x * 2 })
	if !doubled.IsOk() {
		t.Error("expected Ok")
	}

	if doubled.Value() != 10 {
		t.Errorf("expected 10, got %v", doubled.Value())
	}

	err := Err[int](errors.New("fail"))

	mapped := err.Map(func(x int) int { return x * 2 })
	if mapped.IsOk() {
		t.Error("expected Err")
	}
}

func TestResultFlatMap(t *testing.T) {
	ok := Ok(5)

	result := ok.FlatMap(func(x int) Result[int] {
		if x > 0 {
			return Ok(x * 2)
		}

		return Err[int](errors.New("negative"))
	})
	if !result.IsOk() || result.Value() != 10 {
		t.Error("expected Ok(10)")
	}

	err := Err[int](errors.New("fail"))

	result = err.FlatMap(func(x int) Result[int] { return Ok(x * 2) })
	if result.IsOk() {
		t.Error("expected Err")
	}
}

func TestResultAndOr(t *testing.T) {
	ok1 := Ok(1)

	ok2 := Ok(2)
	if r := ok1.And(ok2); !r.IsOk() || r.Value() != 2 {
		t.Error("expected Ok(2)")
	}

	err1 := Err[int](errors.New("fail"))
	if r := err1.And(ok2); r.IsOk() {
		t.Error("expected Err")
	}

	if r := ok1.Or(err1); !r.IsOk() || r.Value() != 1 {
		t.Error("expected Ok(1)")
	}

	if r := err1.Or(ok2); !r.IsOk() || r.Value() != 2 {
		t.Error("expected Ok(2)")
	}
}

func TestResultString(t *testing.T) {
	ok := Ok(42)
	if s := ok.String(); s != "Ok(42)" {
		t.Errorf("expected 'Ok(42)', got '%s'", s)
	}

	err := Err[int](errors.New("fail"))
	if s := err.String(); s != "Err(fail)" {
		t.Errorf("expected 'Err(fail)', got '%s'", s)
	}
}

func TestResultPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic")
		}
	}()

	err := Err[int](errors.New("fail"))
	_ = err.Value() // Should panic
}

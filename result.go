package finding

import (
	"encoding/json"
	"fmt"
)

// Result is a type that represents either a success value or an error.
// It provides a functional approach to error handling similar to Rust's Result<T, E>
// and Haskell's Either type.
type Result[T any] struct {
	value T
	err   error
	ok    bool
}

// Ok creates a Result containing a success value.
func Ok[T any](value T) Result[T] {
	return Result[T]{value: value, ok: true}
}

// Err creates a Result containing an error.
func Err[T any](err error) Result[T] {
	return Result[T]{err: err, ok: false}
}

// Errf creates a Result containing a formatted error.
func Errf[T any](format string, args ...any) Result[T] {
	return Result[T]{err: fmt.Errorf(format, args...), ok: false}
}

// IsOk returns true if the result is Ok.
func (r Result[T]) IsOk() bool {
	return r.ok
}

// IsErr returns true if the result is an error.
func (r Result[T]) IsErr() bool {
	return !r.ok
}

// Value returns the success value.
// Panics if the result is an error.
func (r Result[T]) Value() T {
	if !r.ok {
		panic("called Value on an Err Result")
	}

	return r.value
}

// ValueOr returns the success value or the provided default.
func (r Result[T]) ValueOr(def T) T {
	if r.ok {
		return r.value
	}

	return def
}

// ValueOrElse returns the success value or computes it from the error.
func (r Result[T]) ValueOrElse(f func(error) T) T {
	if r.ok {
		return r.value
	}

	return f(r.err)
}

// Error returns the error.
// Returns nil if the result is Ok.
func (r Result[T]) Error() error {
	return r.err
}

// Unwrap returns the value and error.
// This allows direct assignment: val, err := res.Unwrap().
func (r Result[T]) Unwrap() (T, error) {
	return r.value, r.err
}

// UnwrapOr returns the value if Ok, otherwise returns the default.
func (r Result[T]) UnwrapOr(def T) T {
	if r.ok {
		return r.value
	}

	return def
}

// UnwrapOrDefault returns the value if Ok, otherwise returns the zero value.
func (r Result[T]) UnwrapOrDefault() T {
	var zero T

	return r.UnwrapOr(zero)
}

// Expect returns the value if Ok, otherwise panics with the given message.
func (r Result[T]) Expect(msg string) T {
	if !r.ok {
		if r.err != nil {
			panic(fmt.Sprintf("%s: %v", msg, r.err))
		}

		panic(msg)
	}

	return r.value
}

// ExpectErr returns the error if Err, otherwise panics with the given message.
func (r Result[T]) ExpectErr(msg string) error {
	if r.ok {
		panic(msg)
	}

	return r.err
}

// Map transforms the value if Ok.
func (r Result[T]) Map(f func(T) T) Result[T] {
	if r.ok {
		return Ok(f(r.value))
	}

	return r
}

// MapErr transforms the error if Err.
func (r Result[T]) MapErr(f func(error) error) Result[T] {
	if !r.ok {
		return Err[T](f(r.err))
	}

	return r
}

// And returns other if Ok, otherwise returns the error.
func (r Result[T]) And(other Result[T]) Result[T] {
	if r.ok {
		return other
	}

	return r
}

// Or returns self if Ok, otherwise returns other.
func (r Result[T]) Or(other Result[T]) Result[T] {
	if r.ok {
		return r
	}

	return other
}

// FlatMap chains operations that return Results.
func (r Result[T]) FlatMap(f func(T) Result[T]) Result[T] {
	if r.ok {
		return f(r.value)
	}

	return r
}

// MarshalJSON implements json.Marshaler.
func (r Result[T]) MarshalJSON() ([]byte, error) {
	if r.ok {
		return json.Marshal(map[string]any{
			"ok":    true,
			"value": r.value,
		})
	}

	return json.Marshal(map[string]any{
		"ok":    false,
		"error": r.err.Error(),
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (r *Result[T]) UnmarshalJSON(data []byte) error {
	var aux struct {
		Ok    bool            `json:"ok"`
		Value json.RawMessage `json:"value"`
		Error string          `json:"error"`
	}
	err := json.Unmarshal(data, &aux)
	if err != nil {
		return err
	}

	r.ok = aux.Ok
	if aux.Ok {
		var val T
		err := json.Unmarshal(aux.Value, &val)
		if err != nil {
			return err
		}

		r.value = val
	} else {
		r.err = fmt.Errorf("%s", aux.Error)
	}

	return nil
}

// String implements fmt.Stringer.
func (r Result[T]) String() string {
	if r.ok {
		return fmt.Sprintf("Ok(%v)", r.value)
	}

	return fmt.Sprintf("Err(%v)", r.err)
}

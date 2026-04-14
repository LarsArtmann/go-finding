package finding

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"
)

func TestResultErrf(t *testing.T) {
	t.Parallel()

	r := Errf[int]("code %d: %s", 404, "not found")
	if r.IsOk() {
		t.Error("expected Err")
	}
	if r.Error() == nil {
		t.Error("expected non-nil error")
	}
	if got := r.Error().Error(); got != "code 404: not found" {
		t.Errorf("error = %q, want %q", got, "code 404: not found")
	}
}

func TestResultUnwrapOr(t *testing.T) {
	t.Parallel()

	ok := Ok(42)
	if v := ok.UnwrapOr(0); v != 42 {
		t.Errorf("UnwrapOr on Ok = %d, want 42", v)
	}

	err := Err[int](errors.New("fail"))
	if v := err.UnwrapOr(99); v != 99 {
		t.Errorf("UnwrapOr on Err = %d, want 99", v)
	}
}

func TestResultUnwrapOrDefault(t *testing.T) {
	t.Parallel()

	ok := Ok(42)
	if v := ok.UnwrapOrDefault(); v != 42 {
		t.Errorf("UnwrapOrDefault on Ok = %d, want 42", v)
	}

	err := Err[int](errors.New("fail"))
	if v := err.UnwrapOrDefault(); v != 0 {
		t.Errorf("UnwrapOrDefault on Err = %d, want 0 (zero value)", v)
	}
}

func TestResultExpect(t *testing.T) {
	t.Parallel()

	t.Run("Ok returns value", func(t *testing.T) {
		t.Parallel()

		r := Ok("hello")
		if v := r.Expect("should not panic"); v != "hello" {
			t.Errorf("Expect = %q, want %q", v, "hello")
		}
	})

	t.Run("Err panics with message", func(t *testing.T) {
		t.Parallel()

		defer func() {
			r := recover()
			if r == nil {
				t.Error("expected panic")
			}
			got := fmt.Sprintf("%v", r)
			if got != "boom: oops" {
				t.Errorf("panic = %q, want %q", got, "boom: oops")
			}
		}()

		r := Err[int](errors.New("oops"))
		r.Expect("boom")
	})
}

func TestResultExpectErr(t *testing.T) {
	t.Parallel()

	t.Run("Err returns error", func(t *testing.T) {
		t.Parallel()

		r := Err[int](errors.New("oops"))
		err := r.ExpectErr("should not panic")
		if err.Error() != "oops" {
			t.Errorf("ExpectErr = %q, want %q", err.Error(), "oops")
		}
	})

	t.Run("Ok panics", func(t *testing.T) {
		t.Parallel()

		defer func() {
			r := recover()
			if r == nil {
				t.Error("expected panic")
			}
			if got := fmt.Sprintf("%v", r); got != "unexpected success" {
				t.Errorf("panic = %q, want %q", got, "unexpected success")
			}
		}()

		r := Ok(42)
		r.ExpectErr("unexpected success")
	})
}

func TestResultMapErr(t *testing.T) {
	t.Parallel()

	t.Run("Err transforms error", func(t *testing.T) {
		t.Parallel()

		r := Err[int](errors.New("base"))
		mapped := r.MapErr(func(e error) error {
			return fmt.Errorf("wrapped: %w", e)
		})
		if mapped.IsOk() {
			t.Error("expected Err")
		}
		if got := mapped.Error().Error(); got != "wrapped: base" {
			t.Errorf("error = %q, want %q", got, "wrapped: base")
		}
	})

	t.Run("Ok passes through", func(t *testing.T) {
		t.Parallel()

		r := Ok(42)
		mapped := r.MapErr(func(e error) error {
			return errors.New("should not be called")
		})
		if !mapped.IsOk() {
			t.Error("expected Ok")
		}
		if mapped.Value() != 42 {
			t.Errorf("value = %d, want 42", mapped.Value())
		}
	})
}

func TestResultMarshalJSON(t *testing.T) {
	t.Parallel()

	t.Run("Ok result", func(t *testing.T) {
		t.Parallel()

		r := Ok(42)
		data, err := json.Marshal(r)
		if err != nil {
			t.Fatalf("MarshalJSON: %v", err)
		}

		got := string(data)
		if got != `{"ok":true,"value":42}` {
			t.Errorf("MarshalJSON = %s, want {\"ok\":true,\"value\":42}", got)
		}
	})

	t.Run("Err result", func(t *testing.T) {
		t.Parallel()

		r := Err[int](errors.New("fail"))
		data, err := json.Marshal(r)
		if err != nil {
			t.Fatalf("MarshalJSON: %v", err)
		}

		got := string(data)
		want := `{"ok":false,"error":"fail"}`
		alt := `{"error":"fail","ok":false}`
		if got != want && got != alt {
			t.Errorf("MarshalJSON = %s, want %s or %s", got, want, alt)
		}
	})
}

func TestResultUnmarshalJSON(t *testing.T) {
	t.Parallel()

	t.Run("Ok round-trip", func(t *testing.T) {
		t.Parallel()

		orig := Ok("hello")
		data, err := json.Marshal(orig)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}

		var got Result[string]
		if err := json.Unmarshal(data, &got); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}

		if !got.IsOk() {
			t.Error("expected Ok")
		}
		if got.Value() != "hello" {
			t.Errorf("value = %q, want %q", got.Value(), "hello")
		}
	})

	t.Run("Err round-trip", func(t *testing.T) {
		t.Parallel()

		orig := Err[int](errors.New("broken"))
		data, err := json.Marshal(orig)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}

		var got Result[int]
		if err := json.Unmarshal(data, &got); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}

		if got.IsOk() {
			t.Error("expected Err")
		}
		if got.Error().Error() != "broken" {
			t.Errorf("error = %q, want %q", got.Error().Error(), "broken")
		}
	})
}

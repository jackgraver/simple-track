package env

import (
	"errors"
	"testing"
)

func TestString_missing(t *testing.T) {
	t.Setenv("ENV_TEST_X", "")
	_, err := String("ENV_TEST_X")
	if err == nil || !errors.Is(err, ErrMissing) {
		t.Fatalf("String: want ErrMissing, got %v", err)
	}
}

func TestString_present(t *testing.T) {
	t.Setenv("ENV_TEST_Y", "  hello  ")
	v, err := String("ENV_TEST_Y")
	if err != nil {
		t.Fatal(err)
	}
	if v != "hello" {
		t.Fatalf("got %q", v)
	}
}

func TestStringOr_fallback(t *testing.T) {
	t.Setenv("ENV_TEST_Z", "")
	if got := StringOr("ENV_TEST_Z", "def"); got != "def" {
		t.Fatalf("got %q", got)
	}
}

func TestIntOr_invalidUsesFallback(t *testing.T) {
	t.Setenv("ENV_TEST_I", "not-a-number")
	if got := IntOr("ENV_TEST_I", 99); got != 99 {
		t.Fatalf("got %d", got)
	}
	t.Setenv("ENV_TEST_I2", "0")
	if got := IntOr("ENV_TEST_I2", 7); got != 7 {
		t.Fatalf("got %d", got)
	}
}

func TestSliceOr_trimsAndSkipsEmpty(t *testing.T) {
	t.Setenv("ENV_TEST_S", " a , , b , c ")
	got := SliceOr("ENV_TEST_S", ",", []string{"fallback"})
	if len(got) != 3 || got[0] != "a" || got[1] != "b" || got[2] != "c" {
		t.Fatalf("got %#v", got)
	}
}

func TestSliceOr_emptyUsesFallback(t *testing.T) {
	t.Setenv("ENV_TEST_T", "")
	fb := []string{"x", "y"}
	got := SliceOr("ENV_TEST_T", ",", fb)
	if len(got) != 2 || got[0] != "x" {
		t.Fatalf("got %#v", got)
	}
}

func TestSlice_emptyErr(t *testing.T) {
	t.Setenv("ENV_TEST_U", "  ,  ")
	_, err := Slice("ENV_TEST_U", ",")
	if err == nil || !errors.Is(err, ErrMissing) {
		t.Fatalf("want ErrMissing, got %v", err)
	}
}

package ptr_test

import (
	"testing"

	"github.com/mickamy/dbtop/internal/ptr"
)

func TestOrZero(t *testing.T) {
	t.Parallel()

	if got := ptr.OrZero[string](nil); got != "" {
		t.Errorf("OrZero(nil) = %q, want empty", got)
	}

	if got := ptr.OrZero(new("hello")); got != "hello" {
		t.Errorf("OrZero(*\"hello\") = %q, want \"hello\"", got)
	}
}

func TestMap(t *testing.T) {
	t.Parallel()

	double := func(n int) int { return n * 2 }

	if got := ptr.Map[int, int](nil, double); got != nil {
		t.Errorf("Map(nil) = %v, want nil", got)
	}

	got := ptr.Map(new(21), double)
	if got == nil || *got != 42 {
		t.Errorf("Map(*21) = %v, want 42", got)
	}
}

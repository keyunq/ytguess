package mymath

import "testing"

func TestMyAdd(t *testing.T) {
    a, b := 4, 2
    if x := MyAdd(a, b); x != 6 {
        t.Errorf("MyAdd(%d, %d) = %d, want %d", a, b, x, 6)
    }
}
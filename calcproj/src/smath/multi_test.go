package smath
import "testing"

func TestAdd2(t *testing.T) {
  r ,_:= Multi(1,2)
  if r != 2 {
    t.Errorf("Add(1,2) failed. Got %d, expected 2.",r)
  }
}


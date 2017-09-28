package smath
import "testing"
import "math"

func TestAdd3(t *testing.T) {
  r,err:= Divi(1,2)
  if err != nil {
    t.Errorf("Add(1,2) failed. Got %s",err)
  }
  if math.Dim(float64(r),0.5) >= 0.00001 {
    t.Errorf("Add(1,2) failed. Got %d, expected 0.5.",r)
  }
}


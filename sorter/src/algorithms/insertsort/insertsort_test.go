package insertsort

import "testing"

func TestInsertSort1(t *testing.T) {
	values := []int{5,4,3,2,1}
	InsertSort(values)
	if values[0] != 1 || values[1] != 2 || values[2] != 3 || values[3] != 4 ||
		values[4] != 5	{
		t.Error("InsertSort() failed. Got", values, "Expected 1 2 3 4 5")
	}
}

func TestInsertSort2(t *testing.T) {
	values := []int{5,5,3,2,1}
	InsertSort(values)
	if values[0] != 1 || values[1] != 2 || values[2] != 3 || values[3] != 5 ||
		values[4] != 5	{
		t.Error("InsertSort() failed. Got", values, "Expected 1 2 3 5 5")
	}
}

func TestInsertSort3(t *testing.T) {
	values := []int{1}
	InsertSort(values)
	if values[0] != 1 {
		t.Error("InsertSort() failed. Got", values, "Expected 1")
	}
}

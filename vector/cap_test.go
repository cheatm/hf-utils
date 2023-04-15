package vector

import (
	"testing"
)

func TestCap(t *testing.T) {
	v := []int{0, 1, 2, 3}
	t.Logf("PointV: %p\n", v)
	VAppend(t, &v, 4)
	t.Logf("PointV: %v\n", v)
}

func VAppend(t *testing.T, v *[]int, elem int) {
	*v = append(*v, elem)
	t.Logf("NewV : %p\n", v)
}

func TestVector(t *testing.T) {
	v := New[int]()
	v.Append(1)
	t.Logf("V: %v, Cap: %d, Ptr: %p\n", v, cap(v), v)
	v.Append(2)
	v.Append(3)
	t.Logf("V: %v, Cap: %d, Ptr: %p\n", v, cap(v), v)
	v.Append(4)
	t.Logf("V: %v, Cap: %d, Ptr: %p\n", v, cap(v), v)
	v.Append(5)
	t.Logf("V: %v, Cap: %d, Ptr: %p\n", v, cap(v), v)

}

func TestPtr(t *testing.T) {
	i := 1
	t.Logf("Ptr: %p\n", &i)
}

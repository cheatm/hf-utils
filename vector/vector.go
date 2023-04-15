package vector

type Vector[T any] []T

func New[T any]() Vector[T] {
	return make(Vector[T], 0)
}

func Make[T any](length, capacity int) Vector[T] {
	return make(Vector[T], length, capacity)
}

func (v *Vector[T]) Append(elem T) {
	l := len(*v)
	if l < cap(*v) {
		*v = (*v)[:l+1]
		(*v)[l] = elem
	} else {
		*v = append(*v, elem)
	}
}

func (v *Vector[T]) Pop() {
	l := len(*v)
	if l > 0 {
		*v = (*v)[:l-1]
	}
}

func MapVector[T any, R any](arr Vector[T], mapper func(T) R) Vector[R] {
	result := make(Vector[R], len(arr))
	for i, v := range arr {
		result[i] = mapper(v)
	}
	return result
}

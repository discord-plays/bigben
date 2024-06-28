package utils

func Map[T, U any](ts []T, f func(T) U) []U {
	return MapIndex(ts, func(t T, i int) U {
		return f(t)
	})
}

func MapIndex[T, U any](ts []T, f func(T, int) U) []U {
	us := make([]U, len(ts))
	for i := range ts {
		us[i] = f(ts[i], i)
	}
	return us
}

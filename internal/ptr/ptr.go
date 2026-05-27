package ptr

func OrZero[T any](p *T) T {
	if p == nil {
		var zero T
		return zero
	}

	return *p
}

func Map[T, U any](p *T, f func(T) U) *U {
	if p == nil {
		return nil
	}

	return new(f(*p))
}

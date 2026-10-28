package either

func Must[T any](val T, err error) T {
	if err != nil {
		panic(err)
	}
	return val
}

func Error[T any](val T, err error) error {
	if err == nil {
		panic("err is nil")
	}
	return err
}

func Left[T any](val T, err error) T {
	return val
}

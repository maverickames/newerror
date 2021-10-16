package newerror

type genError struct {
	priErr error
	pubErr error
}

func (ge genError) PriError() error {
	return ge.priErr
}

func (ge genError) PubError() error {
	return ge.pubErr
}

func (ge genError) PubContext() string {
	return ""
}

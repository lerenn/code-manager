package cgwt

type CGWT interface {
}

type cgwt struct {
}

func NewCGWT() CGWT {
	return &cgwt{}
}

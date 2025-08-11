package fs

type FS interface {
}

type fs struct {
}

func NewFS() FS {
	return &fs{}
}

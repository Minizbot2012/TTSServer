package ttsserver

import (
	"io"
	"strings"
)

type ReadNotify struct {
	r io.ReadCloser
	C chan bool
}

func (s *ReadNotify) Read(d []byte) (int, error) {
	n, err := s.r.Read(d)
	if err != nil {
		println(err.Error())
		if err == io.EOF || strings.Contains(err.Error(), "connection reset") {
			s.C <- true
			s.r.Close()
			return n, err
		}
	}
	return n, err
}

func (s *ReadNotify) Close() error {
	s.C <- true
	return s.r.Close()
}

func NewNotify(r io.ReadCloser) *ReadNotify {
	chn := make(chan bool, 1)
	return &ReadNotify{r: r, C: chn}
}

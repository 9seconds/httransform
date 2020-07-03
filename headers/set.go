package headers

import "strings"

type Set struct {
	headers []*Header
}

func (s *Set) Add(name, value string) *Header {
	header := AcquireHeader(name, value)
	s.headers = append(s.headers, header)

	return header
}

func (s *Set) Set(name, value string) *Header {
	s.Delete(name)
	return s.Add(name, value)
}

func (s *Set) Delete(name string) {
	headerID := strings.ToLower(name)
	newHeaders := make([]*Header, 0, len(s.headers))

	for _, v := range s.headers {
		if v.id != headerID {
			newHeaders = append(newHeaders, v)
		} else {
			ReleaseHeader(v)
		}
	}

	s.headers = newHeaders
}

func (s *Set) Get(name string) *Header {
	headerID := strings.ToLower(name)

	for i := len(s.headers) - 1; i >= 0; i-- {
		if s.headers[i].id == headerID {
			return s.headers[i]
		}
	}

	return nil
}

func (s *Set) GetAll(name string) []*Header {
	headerID := strings.ToLower(name)
	headers := make([]*Header, 0, len(s.headers))

	for _, v := range s.headers {
		if v.id == headerID {
			headers = append(headers, v)
		}
	}

	return headers
}

func (s *Set) All() []*Header {
	return s.headers
}

func (s *Set) SetAllHeaders(headers []*Header) {
	s.headers = headers
}

func (s *Set) Reset() {
	for _, v := range s.headers {
		ReleaseHeader(v)
	}

	s.headers = s.headers[:0]
}

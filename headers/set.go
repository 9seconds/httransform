package headers

import "strings"

type Set struct {
	index          map[string]int
	data           []*Header
	removedHeaders map[string]struct{}
}

func (s *Set) SetBytes(key, value []byte) {
	s.SetString(string(key), string(value))
}

func (s *Set) SetString(key, value string) {
	id := strings.ToLower(key)

	if pos, ok := s.index[id]; ok {
		s.data[pos].Value = value
		delete(s.removedHeaders, id)

		return
	}

	header := AcquireHeader()
	header.ID = id
	header.Name = key
	header.Value = value
	s.data = append(s.data, header)
	s.index[id] = len(s.data) - 1

	delete(s.removedHeaders, id)
}

func (s *Set) DeleteBytes(key []byte) {
	s.DeleteString(string(key))
}

func (s *Set) DeleteString(key string) {
	s.removedHeaders[strings.ToLower(key)] = struct{}{}
}

func (s *Set) GetBytes(key []byte) (string, bool) {
	return s.GetString(string(key))
}

func (s *Set) GetString(key string) (string, bool) {
	id := strings.ToLower(key)

	if _, ok := s.removedHeaders[id]; ok {
		return "", false
	}

	if pos, ok := s.index[id]; ok {
		return s.data[pos].Value, true
	}

	return "", false
}

func (s *Set) Items() []*Header {
	headers := make([]*Header, 0, len(s.data))

	for _, v := range s.data {
		if _, ok := s.removedHeaders[v.ID]; !ok {
			headers = append(headers, v)
		}
	}

	return headers
}

func (s *Set) Reset() {
	for k := range s.index {
		delete(s.index, k)
	}

	for k := range s.removedHeaders {
		delete(s.removedHeaders, k)
	}

	for _, v := range s.data {
		ReleaseHeader(v)
	}

	s.data = s.data[:0]
}

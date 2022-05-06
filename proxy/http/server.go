package http

import (
	"bufio"
	"net/http"
	"net/url"
)

type Server struct {
	url    string
	target *url.URL
	weight int
}

func NewServer(urls string) (*Server, error) {
	proxy, err := url.Parse(urls)
	if err != nil {
		return nil, err
	}
	return &Server{target: proxy}, nil
}

func (s Server) LoadConn() int32 {
	panic("implement me")
}

func (s Server) Weight() int {
	return s.weight
}

func (s Server) Target() string {
	return s.url
}

func (s Server) Reverse(request *http.Request, write http.ResponseWriter) error {
	request.URL.Scheme = s.target.Scheme
	request.URL.Host = s.target.Host
	transport := http.DefaultTransport
	resp, err := transport.RoundTrip(request)
	if err != nil {
		return err
	}
	for k, vv := range resp.Header {
		for _, v := range vv {
			write.Header().Add(k, v)
		}
	}
	defer resp.Body.Close()
	bufio.NewReader(resp.Body).WriteTo(write)
	return nil
}

func (s Server) Protocol() string {
	return "http"
}

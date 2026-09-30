package proxy

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

// HTTPServer serves HTTP CONNECT (and rejects other methods) with Basic auth.
type HTTPServer struct {
	Addr   string
	Auth   Authenticator
	Logger *log.Logger
}

// ListenAndServe binds and serves until the listener fails.
func (s *HTTPServer) ListenAndServe() error {
	ln, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return fmt.Errorf("http proxy: listen %s: %w", s.Addr, err)
	}
	return s.Serve(ln)
}

// Serve accepts connections on ln.
func (s *HTTPServer) Serve(ln net.Listener) error {
	defer ln.Close()
	for {
		conn, err := ln.Accept()
		if err != nil {
			return err
		}
		go s.handle(conn)
	}
}

func (s *HTTPServer) logf(format string, args ...any) {
	if s.Logger != nil {
		s.Logger.Printf(format, args...)
	}
}

func (s *HTTPServer) handle(conn net.Conn) {
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(30 * time.Second))

	br := bufio.NewReader(conn)
	req, err := http.ReadRequest(br)
	if err != nil {
		s.logf("http: read request: %v", err)
		return
	}

	user, pass, ok := parseBasicAuth(req.Header.Get("Proxy-Authorization"))
	if !ok || s.Auth == nil || !s.Auth.Authenticate(user, pass) {
		_, _ = io.WriteString(conn, "HTTP/1.1 407 Proxy Authentication Required\r\nProxy-Authenticate: Basic realm=\"egressd\"\r\nContent-Length: 0\r\n\r\n")
		return
	}

	if req.Method != http.MethodConnect {
		_, _ = io.WriteString(conn, "HTTP/1.1 405 Method Not Allowed\r\nContent-Length: 0\r\nConnection: close\r\n\r\n")
		return
	}

	target := req.Host
	if target == "" {
		_, _ = io.WriteString(conn, "HTTP/1.1 400 Bad Request\r\nContent-Length: 0\r\n\r\n")
		return
	}
	if !strings.Contains(target, ":") {
		target += ":443"
	}

	_ = conn.SetDeadline(time.Time{}) // clear handshake deadline

	upstream, err := net.DialTimeout("tcp", target, 15*time.Second)
	if err != nil {
		s.logf("http: dial %s: %v", target, err)
		_, _ = io.WriteString(conn, "HTTP/1.1 502 Bad Gateway\r\nContent-Length: 0\r\n\r\n")
		return
	}
	defer upstream.Close()

	if _, err := io.WriteString(conn, "HTTP/1.1 200 Connection Established\r\n\r\n"); err != nil {
		return
	}

	// Any buffered bytes after the request (rare for CONNECT) go upstream.
	if n := br.Buffered(); n > 0 {
		peek, _ := br.Peek(n)
		if _, err := upstream.Write(peek); err != nil {
			return
		}
		_, _ = br.Discard(n)
	}

	errc := make(chan error, 2)
	go func() { _, e := io.Copy(upstream, br); errc <- e }()
	go func() { _, e := io.Copy(conn, upstream); errc <- e }()
	<-errc
}

func parseBasicAuth(header string) (username, password string, ok bool) {
	const prefix = "Basic "
	if !strings.HasPrefix(header, prefix) {
		return "", "", false
	}
	decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(header[len(prefix):]))
	if err != nil {
		return "", "", false
	}
	user, pass, found := strings.Cut(string(decoded), ":")
	if !found {
		return "", "", false
	}
	return user, pass, true
}

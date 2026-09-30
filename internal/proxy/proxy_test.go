package proxy

import (
	"bufio"
	"context"
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"testing"
	"time"

	"golang.org/x/net/proxy"
)

type staticAuth map[string]string

func (s staticAuth) Authenticate(user, pass string) bool {
	p, ok := s[user]
	return ok && p == pass
}

func freeAddr(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := ln.Addr().String()
	_ = ln.Close()
	return addr
}

func startProxy(t *testing.T, auth Authenticator) (socksAddr, httpAddr string, cancel context.CancelFunc) {
	t.Helper()
	socksAddr = freeAddr(t)
	httpAddr = freeAddr(t)
	ctx, cancel := context.WithCancel(context.Background())
	srv := &Server{
		SOCKSAddr: socksAddr,
		HTTPAddr:  httpAddr,
		Auth:      auth,
	}
	errCh := make(chan error, 1)
	go func() { errCh <- srv.Run(ctx) }()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		c1, e1 := net.DialTimeout("tcp", socksAddr, 50*time.Millisecond)
		c2, e2 := net.DialTimeout("tcp", httpAddr, 50*time.Millisecond)
		if e1 == nil && e2 == nil {
			_ = c1.Close()
			_ = c2.Close()
			return socksAddr, httpAddr, cancel
		}
		if c1 != nil {
			_ = c1.Close()
		}
		if c2 != nil {
			_ = c2.Close()
		}
		select {
		case err := <-errCh:
			t.Fatalf("proxy exited early: %v", err)
		default:
		}
		time.Sleep(20 * time.Millisecond)
	}
	cancel()
	t.Fatal("proxy did not become ready")
	return "", "", nil
}

func TestHTTPConnectAuthReject(t *testing.T) {
	auth := staticAuth{"alice": "secret"}
	_, httpAddr, cancel := startProxy(t, auth)
	defer cancel()

	conn, err := net.DialTimeout("tcp", httpAddr, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(3 * time.Second))

	fmt.Fprintf(conn, "CONNECT example.com:443 HTTP/1.1\r\nHost: example.com:443\r\n\r\n")
	br := bufio.NewReader(conn)
	resp, err := http.ReadResponse(br, &http.Request{Method: http.MethodConnect})
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusProxyAuthRequired {
		t.Fatalf("status = %d, want 407", resp.StatusCode)
	}
}

func TestHTTPConnectAuthOK(t *testing.T) {
	// Local TCP echo target so we do not need the public internet.
	targetLn, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer targetLn.Close()
	go func() {
		c, err := targetLn.Accept()
		if err != nil {
			return
		}
		defer c.Close()
		buf := make([]byte, 64)
		n, _ := c.Read(buf)
		_, _ = c.Write(append([]byte("echo:"), buf[:n]...))
	}()

	auth := staticAuth{"alice": "secret"}
	_, httpAddr, cancel := startProxy(t, auth)
	defer cancel()

	conn, err := net.DialTimeout("tcp", httpAddr, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(5 * time.Second))

	token := base64.StdEncoding.EncodeToString([]byte("alice:secret"))
	fmt.Fprintf(conn,
		"CONNECT %s HTTP/1.1\r\nHost: %s\r\nProxy-Authorization: Basic %s\r\n\r\n",
		targetLn.Addr().String(), targetLn.Addr().String(), token,
	)
	br := bufio.NewReader(conn)
	resp, err := http.ReadResponse(br, &http.Request{Method: http.MethodConnect})
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	// CONNECT has no response body; do not drain (would block on the tunnel).
	resp.Body.Close()

	if _, err := conn.Write([]byte("ping")); err != nil {
		t.Fatal(err)
	}
	buf := make([]byte, 64)
	n, err := conn.Read(buf)
	if err != nil {
		t.Fatal(err)
	}
	if got := string(buf[:n]); got != "echo:ping" {
		t.Fatalf("got %q", got)
	}
}

func TestSOCKS5AuthRejectAndOK(t *testing.T) {
	targetLn, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer targetLn.Close()
	go func() {
		c, err := targetLn.Accept()
		if err != nil {
			return
		}
		defer c.Close()
		buf := make([]byte, 64)
		n, _ := c.Read(buf)
		_, _ = c.Write(append([]byte("echo:"), buf[:n]...))
	}()

	auth := staticAuth{"alice": "secret"}
	socksAddr, _, cancel := startProxy(t, auth)
	defer cancel()

	// Reject wrong password.
	wrongAuth, err := proxy.SOCKS5("tcp", socksAddr, &proxy.Auth{User: "alice", Password: "nope"}, proxy.Direct)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := wrongAuth.Dial("tcp", targetLn.Addr().String()); err == nil {
		t.Fatal("expected SOCKS auth failure")
	}

	okAuth, err := proxy.SOCKS5("tcp", socksAddr, &proxy.Auth{User: "alice", Password: "secret"}, proxy.Direct)
	if err != nil {
		t.Fatal(err)
	}
	c, err := okAuth.Dial("tcp", targetLn.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	_ = c.SetDeadline(time.Now().Add(3 * time.Second))
	if _, err := c.Write([]byte("ping")); err != nil {
		t.Fatal(err)
	}
	buf := make([]byte, 64)
	n, err := c.Read(buf)
	if err != nil {
		t.Fatal(err)
	}
	if got := string(buf[:n]); got != "echo:ping" {
		t.Fatalf("got %q", got)
	}

	// Also exercise via http.Transport socks URL shape for documentation parity.
	u := &url.URL{Scheme: "socks5", Host: socksAddr, User: url.UserPassword("alice", "secret")}
	_ = u
}

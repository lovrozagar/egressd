package proxy

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"sync"

	"github.com/things-go/go-socks5"
)

// Server runs SOCKS5 and HTTP CONNECT proxies with shared auth.
type Server struct {
	SOCKSAddr string
	HTTPAddr  string
	Auth      Authenticator
	Logger    *log.Logger
}

// credentialAdapter bridges Authenticator to go-socks5 CredentialStore.
type credentialAdapter struct {
	auth Authenticator
}

func (c credentialAdapter) Valid(user, password, _ string) bool {
	if c.auth == nil {
		return false
	}
	return c.auth.Authenticate(user, password)
}

// Run starts both listeners and blocks until ctx is cancelled or a listener fails.
func (s *Server) Run(ctx context.Context) error {
	logger := s.Logger
	if logger == nil {
		logger = log.New(os.Stdout, "", log.LstdFlags)
	}

	socksLn, err := net.Listen("tcp", s.SOCKSAddr)
	if err != nil {
		return fmt.Errorf("socks: listen %s: %w", s.SOCKSAddr, err)
	}
	httpLn, err := net.Listen("tcp", s.HTTPAddr)
	if err != nil {
		_ = socksLn.Close()
		return fmt.Errorf("http: listen %s: %w", s.HTTPAddr, err)
	}

	socksSrv := socks5.NewServer(
		socks5.WithCredential(credentialAdapter{auth: s.Auth}),
		socks5.WithLogger(socks5.NewLogger(logger)),
	)

	httpSrv := &HTTPServer{
		Addr:   s.HTTPAddr,
		Auth:   s.Auth,
		Logger: logger,
	}

	errCh := make(chan error, 2)
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		logger.Printf("SOCKS5 listening on %s", socksLn.Addr())
		if err := socksSrv.Serve(socksLn); err != nil {
			select {
			case <-ctx.Done():
			default:
				errCh <- fmt.Errorf("socks: %w", err)
			}
		}
	}()

	go func() {
		defer wg.Done()
		logger.Printf("HTTP CONNECT listening on %s", httpLn.Addr())
		if err := httpSrv.Serve(httpLn); err != nil {
			select {
			case <-ctx.Done():
			default:
				errCh <- fmt.Errorf("http: %w", err)
			}
		}
	}()

	select {
	case <-ctx.Done():
		_ = socksLn.Close()
		_ = httpLn.Close()
		wg.Wait()
		return nil
	case err := <-errCh:
		_ = socksLn.Close()
		_ = httpLn.Close()
		wg.Wait()
		return err
	}
}

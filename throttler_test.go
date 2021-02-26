package throttler

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

type spyLogger struct {
	called  bool
	message []interface{}
}

func NewSpyLogger() *spyLogger {
	return &spyLogger{
		called: false,
	}
}

func (s *spyLogger) isCalled() bool {
	return s.called
}

func (s *spyLogger) Print(v ...interface{}) {
	s.message = v
	s.called = true
}

func fakeHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
}

func setupServer(opt *opts) *httptest.Server {
	mw := Middleware(opt)

	wrap := mw(fakeHandler)

	srv := httptest.NewServer(wrap)

	return srv
}

func TestMiddleware_ShouldReturn200(t *testing.T) {
	srv := setupServer(WithOpts(UNLIMITED).WithVerbose(false))
	defer srv.Close()

	resp, _ := http.Get(srv.URL)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Status code should be 200 but was %d", resp.StatusCode)
	}
}

func TestMiddleware_ShouldReturn429(t *testing.T) {
	srv := setupServer(WithOpts(1).WithVerbose(false))
	defer srv.Close()

	resp, _ := http.Get(srv.URL)
	resp, _ = http.Get(srv.URL)

	if resp.StatusCode != http.StatusTooManyRequests {
		t.Errorf("Status code should be 429 but was %d", resp.StatusCode)
	}
}

func TestMiddleware_Verbose_False(t *testing.T) {
	fakeLogger := NewSpyLogger()

	srv := setupServer(WithOpts(1).WithVerbose(false).WithLogger(fakeLogger))
	defer srv.Close()

	http.Get(srv.URL)
	http.Get(srv.URL)

	if fakeLogger.isCalled() {
		t.Errorf("Logger should not be called.")
	}
}

func BenchmarkMiddleware(b *testing.B) {
	for i := 0; i < b.N; i++ {
		srv := setupServer(WithOpts(UNLIMITED).WithVerbose(false))
		defer srv.Close()

		for j := 0; j < i; j++ {
			resp, _ := http.Get(srv.URL)
			if resp.StatusCode != http.StatusOK {
				b.Errorf("Status code expected 200 received %d after %d requests in %d scope", resp.StatusCode, j, i)
			}
		}
	}
}

func BenchmarkMiddleware_Failure(b *testing.B) {
	for i := 0; i < b.N; i++ {
		srv := setupServer(WithOpts(1).WithVerbose(false))
		defer srv.Close()

		for j := 0; j < i; j++ {
			resp, _ := http.Get(srv.URL)
			if resp.StatusCode != http.StatusTooManyRequests && j > 1 {
				b.Errorf("Status code expected 429 received %d after %d requests in %d scope", resp.StatusCode, j+1, i)
			}
		}
	}
}

func BenchmarkMiddleware_Verbose(b *testing.B) {
	for i := 0; i < b.N; i++ {
		fakeLogger := NewSpyLogger()
		srv := setupServer(WithOpts(1).WithVerbose(true).WithLogger(fakeLogger))
		defer srv.Close()

		for j := 0; j < i; j++ {
			resp, _ := http.Get(srv.URL)
			if resp.StatusCode != http.StatusTooManyRequests && j > 1 {
				b.Errorf("Status code expected 429 received %d after %d requests in %d scope", resp.StatusCode, j+1, i)
			}
		}
	}
}

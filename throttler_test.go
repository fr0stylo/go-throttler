package throttler

import (
	"net"
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
	srv := setupServer(NewOpts(UNLIMITED).WithVerbose(false))
	defer srv.Close()

	resp, _ := http.Get(srv.URL)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Status code should be 200 but was %d", resp.StatusCode)
	}
}

func TestMiddleware_ShouldReturn429(t *testing.T) {
	srv := setupServer(NewOpts(1).WithVerbose(false))
	defer srv.Close()

	resp, _ := http.Get(srv.URL)
	resp, _ = http.Get(srv.URL)

	if resp.StatusCode != http.StatusTooManyRequests {
		t.Errorf("Status code should be 429 but was %d", resp.StatusCode)
	}
}

func TestMiddleware_Verbose_False(t *testing.T) {
	fakeLogger := NewSpyLogger()

	srv := setupServer(NewOpts(1).WithVerbose(false).WithLogger(fakeLogger))
	defer srv.Close()

	http.Get(srv.URL)
	http.Get(srv.URL)

	if fakeLogger.isCalled() {
		t.Errorf("Logger should not be called.")
	}
}

func BenchmarkMiddleware(b *testing.B) {
	httptest.NewRequest(http.MethodGet, "/", nil)
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	b.ResetTimer()
	mw := Middleware(NewOpts(int64(b.N)).WithVerbose(false))
	wrap := mw(fakeHandler)
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		wrap(w, r)
		if w.Code != http.StatusOK {
			b.Errorf("Status code expected 200 received %d after %d requests", w.Code, i)
		}
	}
}

func BenchmarkMiddleware_Unlimited_count(b *testing.B) {
	httptest.NewRequest(http.MethodGet, "/", nil)
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	b.ResetTimer()
	mw := Middleware(NewOpts(UNLIMITED).WithVerbose(false))
	wrap := mw(fakeHandler)
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		wrap(w, r)
		if w.Code != http.StatusOK {
			b.Errorf("Status code expected 200 received %d after %d requests", w.Code, i)
		}
	}
}

func BenchmarkMiddleware_Failure(b *testing.B) {
	httptest.NewRequest(http.MethodGet, "/", nil)
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	b.ResetTimer()
	mw := Middleware(NewOpts(1).WithVerbose(false))
	wrap := mw(fakeHandler)
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		wrap(w, r)
		if w.Code != http.StatusTooManyRequests && i > 1 {
			b.Errorf("Status code expected 429 received %d after %d requests", w.Code, i)
		}
	}
}

func BenchmarkMiddleware_Verbose(b *testing.B) {
	fakeLogger := NewSpyLogger()
	httptest.NewRequest(http.MethodGet, "/", nil)
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	b.ResetTimer()
	mw := Middleware(NewOpts(1).WithVerbose(true).WithLogger(fakeLogger))
	wrap := mw(fakeHandler)
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		wrap(w, r)
		if w.Code != http.StatusTooManyRequests && i > 1 {
			b.Errorf("Status code expected 429 received %d after %d requests", w.Code, i)
		}
	}
}

func BenchmarkMiddleware_ThrottlingStrategy(b *testing.B) {
	fakeLogger := NewSpyLogger()
	httptest.NewRequest(http.MethodGet, "/", nil)
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.RemoteAddr = "127.0.0.1"
	b.ResetTimer()
	ip := net.IPNet{
		IP:   net.ParseIP("127.0.0.1"),
		Mask: net.IPv4Mask(255, 255, 255, 0),
	}
	mw := Middleware(NewOpts(1).WithVerbose(true).WithLogger(fakeLogger).WithIpThrottlingStrategy(map[*net.IPNet]int64{&ip: UNLIMITED}))
	wrap := mw(fakeHandler)
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		wrap(w, r)
		if w.Code != http.StatusOK && i > 1 {
			b.Errorf("Status code expected 200 received %d after %d requests", w.Code, i)
		}
	}
}

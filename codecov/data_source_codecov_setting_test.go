package codecov

import (
	"context"
	"errors"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func newDummyServer(t *testing.T, handler func(w http.ResponseWriter, r *http.Request)) (string, func()) {
	t.Helper()
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		http.Serve(l, http.HandlerFunc(handler))
	}()
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-ctx.Done()
		l.Close()
	}()
	return "http://" + l.Addr().String(), cancel
}

func newResourceData(t *testing.T, service, owner, repo string) *schema.ResourceData {
	ds := dataSourceCodecovConfig()
	rd := schema.TestResourceDataRaw(t, ds.Schema, nil)
	rd.Set("service", service)
	rd.Set("owner", owner)
	rd.Set("repo", repo)
	return rd
}

func TestDataCodecovConfigRead(t *testing.T) {
	maxRetry = 3
	retryWaitBase = 100 * time.Millisecond
	waitOnRedirect = 500 * time.Millisecond
	ticker := time.NewTicker(time.Millisecond)
	defer ticker.Stop()

	t.Run("Success", func(t *testing.T) {
		var called bool
		endpoint, cancel := newDummyServer(t, func(w http.ResponseWriter, r *http.Request) {
			if r.RequestURI != "/api/v2/svc123/owner123/repos/repo123/config/" {
				t.Errorf("Unexpected RequestURL: %s", r.RequestURI)
			}
			w.Write([]byte(`{"upload_token":"token1234"}`))
			if called {
				t.Error("API called more than once")
			}
			called = true
		})
		defer cancel()

		rd := newResourceData(t, "svc123", "owner123", "repo123")
		err := dataCodecovConfigRead(context.Background(), rd,
			&providerConfig{
				TokenV2:      "hoge",
				EndpointBase: endpoint,
				APICallTick:  ticker.C,
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		if token := rd.Get("upload_token"); token != "token1234" {
			t.Errorf("Unexpected upload_token: %s", token)
		}
	})
	t.Run("MaxRetries", func(t *testing.T) {
		var cntReq int
		endpoint, cancel := newDummyServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusGatewayTimeout)
			cntReq++
		})
		defer cancel()

		rd := newResourceData(t, "svc123", "owner123", "repo123")
		err := dataCodecovConfigRead(context.Background(), rd,
			&providerConfig{
				TokenV2:      "hoge",
				EndpointBase: endpoint,
				APICallTick:  ticker.C,
			},
		)
		var te *temporaryError
		if !errors.As(err, &te) {
			t.Errorf("Unexpected error: %v", err)
		}
		if cntReq != 4 {
			t.Errorf("Unexpected retry count: %d", cntReq)
		}
	})
	t.Run("SuccessAfterRetries", func(t *testing.T) {
		var cntReq int
		endpoint, cancel := newDummyServer(t, func(w http.ResponseWriter, r *http.Request) {
			if cntReq == 2 {
				w.Write([]byte(`{"upload_token":"token1234"}`))
			} else {
				w.WriteHeader(http.StatusGatewayTimeout)
			}
			cntReq++
		})
		defer cancel()

		rd := newResourceData(t, "svc123", "owner123", "repo123")
		err := dataCodecovConfigRead(context.Background(), rd,
			&providerConfig{
				TokenV2:      "hoge",
				EndpointBase: endpoint,
				APICallTick:  ticker.C,
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		if token := rd.Get("upload_token"); token != "token1234" {
			t.Errorf("Unexpected upload_token: %s", token)
		}
		if cntReq != 3 {
			t.Errorf("Unexpected retry count: %d", cntReq)
		}
	})
	t.Run("FatalStatus", func(t *testing.T) {
		var called bool
		endpoint, cancel := newDummyServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTeapot)
			if called {
				t.Error("API called more than once")
			}
			called = true
		})
		defer cancel()

		rd := newResourceData(t, "svc123", "owner123", "repo123")
		err := dataCodecovConfigRead(context.Background(), rd,
			&providerConfig{
				TokenV2:      "hoge",
				EndpointBase: endpoint,
				APICallTick:  ticker.C,
			},
		)
		var fe *fatalError
		if !errors.As(err, &fe) {
			t.Errorf("Unexpected error: %v", err)
		}
	})
	t.Run("NetworkError", func(t *testing.T) {
		rd := newResourceData(t, "svc123", "owner123", "repo123")
		err := dataCodecovConfigRead(context.Background(), rd,
			&providerConfig{
				TokenV2:      "hoge",
				EndpointBase: "http://localhost:0",
				APICallTick:  ticker.C,
			},
		)
		var ne net.Error
		if !errors.As(err, &ne) {
			t.Errorf("Unexpected error: %v", err)
		}
	})
	t.Run("RetryOnRedirect", func(t *testing.T) {
		var called bool
		endpoint, cancel := newDummyServer(t, func(w http.ResponseWriter, r *http.Request) {
			if called {
				w.Write([]byte(`{"upload_token":"token1234"}`))
			} else {
				w.WriteHeader(http.StatusTemporaryRedirect)
			}
			called = true
		})
		defer cancel()

		rd := newResourceData(t, "svc123", "owner123", "repo123")
		t0 := time.Now()
		err := dataCodecovConfigRead(context.Background(), rd,
			&providerConfig{
				TokenV2:      "hoge",
				EndpointBase: endpoint,
				APICallTick:  ticker.C,
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		dur := time.Now().Sub(t0)
		if token := rd.Get("upload_token"); token != "token1234" {
			t.Errorf("Unexpected upload_token: %s", token)
		}
		if dur < 450*time.Millisecond || 650*time.Millisecond < dur {
			t.Errorf("Unexpected delay: %v", dur)
		}
	})
}

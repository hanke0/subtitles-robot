package httpclient_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hanke0/subtitles-robot/httpclient"
)

func runTestServer(o interface{}) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		enc := json.NewEncoder(w)
		enc.Encode(o)
	})
	return httptest.NewServer(mux)
}

type obj struct {
	A string
}

func TestGet(t *testing.T) {
	c, err := httpclient.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	s := runTestServer(obj{})
	err = c.Get(s.URL).Invoke().Drop()
	if err != nil {
		t.Fatal(err)
	}
}

func TestJSON(t *testing.T) {
	c, err := httpclient.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	var (
		o1 obj
		o2 obj
	)
	o1.A = "12343"
	s := runTestServer(&o1)
	err = c.Get(s.URL).Invoke().JSON(&o2)
	if err != nil {
		t.Fatal(err)
	}
	if o1.A != o2.A {
		t.Fatalf("%s != %s", o2.A, o1.A)
	}
}

func TestWriteTo(t *testing.T) {
	c, err := httpclient.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	s := runTestServer(1)
	var b strings.Builder
	_, err = c.Get(s.URL).Invoke().WriteTo(&b)
	if err != nil {
		t.Fatal(err)
	}
	if b.String() != "1\n" {
		t.Fatalf("%s != %s", b.String(), "1\n")
	}
}

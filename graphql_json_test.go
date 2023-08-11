package graphql

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/matryer/is"
)

type graphResponse struct {
	Data map[string]interface{}
	Errors []interface{}
}

func TestDoJSON(t *testing.T) {
	is := is.New(t)
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		is.Equal(r.Method, http.MethodPost)
		b, err := io.ReadAll(r.Body)
		is.NoErr(err)
		is.Equal(string(b), `{"query":"query {}","variables":null}`+"\n")
		_, err = io.WriteString(w, `{
			"data": {
				"something": "yes"
			}
		}`)
		is.NoErr(err)
	}))
	defer srv.Close()

	ctx := context.Background()
	client := NewClient(srv.URL)

	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()
	gr := &graphResponse{}
	err := client.Run(ctx, &Request{q: "query {}"}, gr)
	is.NoErr(err)
	is.Equal(calls, 1) // calls
	is.Equal(gr.Data["something"], "yes")
}

func TestDoJSONServerError(t *testing.T) {
	is := is.New(t)
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		is.Equal(r.Method, http.MethodPost)
		b, err := io.ReadAll(r.Body)
		is.NoErr(err)
		is.Equal(string(b), `{"query":"query {}","variables":null}`+"\n")
		w.WriteHeader(http.StatusInternalServerError)
		_, err = io.WriteString(w, `Internal Server Error`)
		is.NoErr(err)
	}))
	defer srv.Close()

	ctx := context.Background()
	client := NewClient(srv.URL)

	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()
	gr := &graphResponse{}
	err := client.Run(ctx, &Request{q: "query {}"}, gr)
	is.Equal(calls, 1) // calls
	is.Equal(err.Error(), "graphql: server returned a non-200 status code: 500")
}

func TestDoJSONBadRequestErr(t *testing.T) {
	is := is.New(t)
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		is.Equal(r.Method, http.MethodPost)
		b, err := io.ReadAll(r.Body)
		is.NoErr(err)
		is.Equal(string(b), `{"query":"query {}","variables":null}`+"\n")
		w.WriteHeader(http.StatusBadRequest)
		_, err = io.WriteString(w, `{
			"errors": [{
				"message": "miscellaneous message as to why the the request was bad"
			}]
		}`)
		is.NoErr(err)
	}))
	defer srv.Close()

	ctx := context.Background()
	client := NewClient(srv.URL)

	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()
	gr := &graphResponse{}
	err := client.Run(ctx, &Request{q: "query {}"}, gr)
	is.Equal(calls, 1) // calls
	is.NoErr(err)
	var errors []map[string]interface{}
	for i := range gr.Errors {
		errors = append(errors, gr.Errors[i].(map[string]interface{}))
	}
	is.Equal(errors[0]["message"], "miscellaneous message as to why the the request was bad")
}

func TestQueryJSON(t *testing.T) {
	is := is.New(t)

	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		b, err := io.ReadAll(r.Body)
		is.NoErr(err)
		is.Equal(string(b), `{"query":"query {}","variables":{"username":"matryer"}}`+"\n")
		_, err = io.WriteString(w, `{"data":{"value":"some data"}}`)
		is.NoErr(err)
	}))
	defer srv.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	client := NewClient(srv.URL)

	req := NewRequest("query {}")
	req.Var("username", "matryer")

	// check variables
	is.True(req != nil)
	is.Equal(req.vars["username"], "matryer")

	gr := &graphResponse{}
	err := client.Run(ctx, req, gr)
	is.NoErr(err)
	is.Equal(calls, 1)

	is.Equal(gr.Data["value"], "some data")
}

func TestHeader(t *testing.T) {
	is := is.New(t)

	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		is.Equal(r.Header.Get("X-Custom-Header"), "123")

		_, err := io.WriteString(w, `{"data":{"value":"some data"}}`)
		is.NoErr(err)
	}))
	defer srv.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	client := NewClient(srv.URL)

	req := NewRequest("query {}")
	req.Header.Set("X-Custom-Header", "123")

	gr := &graphResponse{}
	err := client.Run(ctx, req, gr)
	is.NoErr(err)
	is.Equal(calls, 1)

	is.Equal(gr.Data["value"], "some data")
}

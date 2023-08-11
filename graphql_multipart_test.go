package graphql

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/matryer/is"
)

func TestWithClient(t *testing.T) {
	is := is.New(t)
	var calls int
	testClient := &http.Client{
		Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			calls++
			resp := &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"data":{"key":"value"}}`)),
			}
			return resp, nil
		}),
	}

	ctx := context.Background()
	client := NewClient("", WithHTTPClient(testClient), UseMultipartForm())

	req := NewRequest(``)
	gr := &graphResponse{}
	err := client.Run(ctx, req, gr)
	is.NoErr(err)

	is.Equal(calls, 1) // calls
}

func TestDoUseMultipartForm(t *testing.T) {
	is := is.New(t)
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		is.Equal(r.Method, http.MethodPost)
		query := r.FormValue("query")
		is.Equal(query, `query {}`)
		_, err := io.WriteString(w, `{
			"data": {
				"something": "yes"
			}
		}`)
		is.NoErr(err)
	}))
	defer srv.Close()

	ctx := context.Background()
	client := NewClient(srv.URL, UseMultipartForm())

	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()
	gr := &graphResponse{}
	err := client.Run(ctx, &Request{q: "query {}"}, gr)
	is.NoErr(err)
	is.Equal(calls, 1) // calls
	is.Equal(gr.Data["something"], "yes")
}
func TestImmediatelyCloseReqBody(t *testing.T) {
	is := is.New(t)
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		is.Equal(r.Method, http.MethodPost)
		query := r.FormValue("query")
		is.Equal(query, `query {}`)
		_, err := io.WriteString(w, `{
			"data": {
				"something": "yes"
			}
		}`)
		is.NoErr(err)
	}))
	defer srv.Close()

	ctx := context.Background()
	client := NewClient(srv.URL, ImmediatelyCloseReqBody(), UseMultipartForm())

	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()
	gr := &graphResponse{}
	err := client.Run(ctx, &Request{q: "query {}"}, gr)
	is.NoErr(err)
	is.Equal(calls, 1) // calls
	is.Equal(gr.Data["something"], "yes")
}

func TestDoErr(t *testing.T) {
	is := is.New(t)
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		is.Equal(r.Method, http.MethodPost)
		query := r.FormValue("query")
		is.Equal(query, `query {}`)
		_, err := io.WriteString(w, `{
			"errors": [{
				"message": "Something went wrong"
			}]
		}`)
		is.NoErr(err)
	}))
	defer srv.Close()

	ctx := context.Background()
	client := NewClient(srv.URL, UseMultipartForm())

	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()
	gr := &graphResponse{}
	err := client.Run(ctx, &Request{q: "query {}"}, gr)
	is.NoErr(err)
	var errors []map[string]interface{}
	for i := range gr.Errors {
		errors = append(errors, gr.Errors[i].(map[string]interface{}))
	}
	is.Equal(errors[0]["message"], "Something went wrong")
}

func TestDoServerErr(t *testing.T) {
	is := is.New(t)
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		is.Equal(r.Method, http.MethodPost)
		query := r.FormValue("query")
		is.Equal(query, `query {}`)
		w.WriteHeader(http.StatusInternalServerError)
		_, err := io.WriteString(w, `Internal Server Error`)
		is.NoErr(err)
	}))
	defer srv.Close()

	ctx := context.Background()
	client := NewClient(srv.URL, UseMultipartForm())

	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()
	gr := &graphResponse{}
	err := client.Run(ctx, &Request{q: "query {}"}, gr)
	is.Equal(err.Error(), "graphql: server returned a non-200 status code: 500")
}

func TestDoBadRequestErr(t *testing.T) {
	is := is.New(t)
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		is.Equal(r.Method, http.MethodPost)
		query := r.FormValue("query")
		is.Equal(query, `query {}`)
		w.WriteHeader(http.StatusBadRequest)
		_, err := io.WriteString(w, `{
			"errors": [{
				"message": "miscellaneous message as to why the the request was bad"
			}]
		}`)
		is.NoErr(err)
	}))
	defer srv.Close()

	ctx := context.Background()
	client := NewClient(srv.URL, UseMultipartForm())

	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()
	gr := &graphResponse{}
	err := client.Run(ctx, &Request{q: "query {}"}, gr)
	is.NoErr(err)
	var errors []map[string]interface{}
	for i := range gr.Errors {
		errors = append(errors, gr.Errors[i].(map[string]interface{}))
	}
	is.Equal(errors[0]["message"], "miscellaneous message as to why the the request was bad")
}

func TestDoNoResponse(t *testing.T) {
	is := is.New(t)
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		is.Equal(r.Method, http.MethodPost)
		query := r.FormValue("query")
		is.Equal(query, `query {}`)
		_, err := io.WriteString(w, `{
			"data": {
				"something": "yes"
			}
		}`)
		is.NoErr(err)
	}))
	defer srv.Close()

	ctx := context.Background()
	client := NewClient(srv.URL, UseMultipartForm())

	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()
	gr := &graphResponse{}
	err := client.Run(ctx, &Request{q: "query {}"}, gr)
	is.NoErr(err)
	is.Equal(calls, 1) // calls
}

func TestQuery(t *testing.T) {
	is := is.New(t)

	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		query := r.FormValue("query")
		is.Equal(query, "query {}")
		is.Equal(r.FormValue("variables"), `{"username":"matryer"}`+"\n")
		_, err := io.WriteString(w, `{"data":{"value":"some data"}}`)
		is.NoErr(err)
	}))
	defer srv.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	client := NewClient(srv.URL, UseMultipartForm())

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

func TestFile(t *testing.T) {
	is := is.New(t)

	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		file, header, err := r.FormFile("file")
		is.NoErr(err)
		defer file.Close()
		is.Equal(header.Filename, "filename.txt")

		b, err := io.ReadAll(file)
		is.NoErr(err)
		is.Equal(string(b), `This is a file`)

		_, err = io.WriteString(w, `{"data":{"value":"some data"}}`)
		is.NoErr(err)
	}))
	defer srv.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	client := NewClient(srv.URL, UseMultipartForm())
	f := strings.NewReader(`This is a file`)
	req := NewRequest("query {}")
	req.File("file", "filename.txt", f)
	gr := &graphResponse{}
	err := client.Run(ctx, req, gr)
	is.NoErr(err)
}

type roundTripperFunc func(req *http.Request) (*http.Response, error)

func (fn roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

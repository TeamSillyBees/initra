package httpdemo

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/teamsillybees/initra/pkg/httpclient"
)

func TestServiceGetHTTPBingo(t *testing.T) {
	client := &fakeHTTPClient{}
	service := NewService(client)

	vo, err := service.GetHTTPBingo(context.Background(), "hello", "trace-1")

	require.NoError(t, err)
	require.Equal(t, "/get", client.path)
	require.Equal(t, "hello", client.options.QueryParams["message"])
	require.Equal(t, "trace-1", client.options.Headers["X-Trace-ID"])
	require.Equal(t, "GET", vo.Method)
	require.Equal(t, []string{"hello"}, vo.Args["message"])
}

func TestServiceGetHTTPBingoUsesDefaultMessage(t *testing.T) {
	client := &fakeHTTPClient{}
	service := NewService(client)

	_, err := service.GetHTTPBingo(context.Background(), "", "")

	require.NoError(t, err)
	require.Equal(t, defaultMessage, client.options.QueryParams["message"])
}

func TestServiceGetHTTPBingoFormPage(t *testing.T) {
	client := &fakeHTTPClient{}
	service := NewService(client)

	vo, err := service.GetHTTPBingoFormPage(context.Background(), "trace-1")

	require.NoError(t, err)
	require.Equal(t, "/forms/post", client.path)
	require.Equal(t, "trace-1", client.options.Headers["X-Trace-ID"])
	require.Equal(t, "text/html; charset=utf-8", vo.ContentType)
	require.Equal(t, "<form>demo</form>", vo.Body)
	require.Equal(t, len("<form>demo</form>"), vo.Size)
}

func TestServiceMapsHTTPClientError(t *testing.T) {
	service := NewService(&fakeHTTPClient{
		err: &httpclient.Error{
			Kind:       httpclient.ErrorKindResponse,
			Service:    "httpbingo",
			StatusCode: http.StatusBadGateway,
			Message:    "bad gateway",
		},
	})

	_, err := service.GetHTTPBingo(context.Background(), "", "")

	require.Error(t, err)
	require.True(t, errors.Is(err, service.client.(*fakeHTTPClient).err))
}

type fakeHTTPClient struct {
	path    string
	options httpclient.RequestOptions
	err     error
}

func (f *fakeHTTPClient) Get(_ context.Context, path string, opts ...httpclient.RequestOption) (*httpclient.Response, error) {
	f.path = path
	f.options = httpclient.RequestOptions{
		Headers:     map[string]string{},
		QueryParams: map[string]string{},
		PathParams:  map[string]string{},
	}
	for _, opt := range opts {
		opt(&f.options)
	}
	if f.err != nil {
		return nil, f.err
	}
	if f.options.Result != nil {
		payload := f.options.Result.(*httpBingoGetPayload)
		*payload = httpBingoGetPayload{
			Args: map[string][]string{
				"message": {f.options.QueryParams["message"]},
			},
			Headers: map[string][]string{
				"X-Trace-ID": {f.options.Headers["X-Trace-ID"]},
			},
			Method: "GET",
			Origin: "127.0.0.1",
			URL:    "https://httpbingo.org/get",
		}
		return &httpclient.Response{StatusCode: http.StatusOK}, nil
	}
	return &httpclient.Response{
		StatusCode: http.StatusOK,
		Header: http.Header{
			"Content-Type": {"text/html; charset=utf-8"},
		},
		Body: []byte("<form>demo</form>"),
	}, nil
}

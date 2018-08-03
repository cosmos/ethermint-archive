package server

import (
	"testing"
	"net/http/httptest"
	"net/http"
	"strings"
	"github.com/stretchr/testify/require"
	"github.com/gorilla/rpc/json"
	gojson "encoding/json"
	"io"
)

type OkArgs struct {
	Message string
}

type OkReply struct {
	Message        string
	AdditionalData uint64
}

type internalArgs struct {
}

func TestServer_GeneralJSONRPC(t *testing.T) {
	s := NewServer()
	s.RegisterCodec(json.NewCodec(), "application/json")
	s.RegisterHandler("test_method", func(args *OkArgs, reply *OkReply) error {
		reply.Message = args.Message
		reply.AdditionalData = 1234
		return nil
	})
	ts := httptest.NewServer(s)
	defer ts.Close()

	res, err := http.Post(ts.URL, "application/json", strings.NewReader(`
		{
			"id": 1,
			"method": "test_method",
			"params": [{
				"message": "this is a test message"
			}]
		}
	`))
	require.Nil(t, err, "unexpected error making request")
	deserBody, err := readJson(res.Body)
	require.Nil(t, err, "unexpected error deserializing json")
	result := deserBody["result"].(map[string]interface{})
	require.Equal(t, float64(1), deserBody["id"])
	require.Equal(t, "this is a test message", result["Message"])
	require.Equal(t, float64(1234), result["AdditionalData"])
}

func readJson(r io.ReadCloser) (map[string]interface{}, error) {
	defer r.Close()
	var out map[string]interface{}
	decoder := gojson.NewDecoder(r)
	if err := decoder.Decode(&out); err != nil {
		return nil, err
	}
	return out, nil
}

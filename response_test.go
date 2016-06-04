package hrr

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Sirupsen/logrus/hooks/test"
)

func Test_ResponseError(t *testing.T) {
	tc := struct {
		Err                Error
		ExpectedBody       string
		ExpectedLogMessage string
	}{
		Err: NewError("X-Men are gay", errors.New("Who is this girl Wolverin")),
		ExpectedBody: `
		{
			"id": "\w*",
			"message": "X-Men are gay"
		}
		`,
		ExpectedLogMessage: "Who is this girl Wolverin",
	}

	// Run test
	{
		logger, mock := test.NewNullLogger()
		Logger = logger

		req := NewRequest(t, "GET", "/", bytes.NewBufferString("X-Men are so cool!"))
		req.RemoteAddr = "127.0.0.1"

		resp := httptest.NewRecorder()
		Response(resp, req).Error(tc.Err)

		d := mock.LastEntry().Data
		if d["remote_addr"] != "127.0.0.1" ||
			d["url"] != "/" ||
			d["method"] != "GET" ||
			d["body"] != "X-Men are so cool!" ||
			mock.LastEntry().Message != tc.ExpectedLogMessage {
			t.Fatalf("Expected (%v, %v, %v, %v, %v) was (%v, %v, %v, %v, %v)",
				"127.0.0.1", "/", "GET", "X-Men are so cool!", tc.ExpectedLogMessage,
				d["remote_addr"], d["url"], d["method"], d["body"], mock.LastEntry().Message,
			)
		}

		if resp.Code != http.StatusBadRequest {
			t.Fatalf("Expected %v was %v", http.StatusBadRequest, resp.Code)
		}

		EqualJSONBody(t, tc.ExpectedBody, resp.Body)

	}
}

func Test_ResponseData(t *testing.T) {
	tc := struct {
		Data        map[string]string
		ExpctedBody string
	}{
		Data: map[string]string{
			"key": "value",
		},
		ExpctedBody: `
		{
			"key": "value"
		}
		`,
	}

	// Run test
	{
		req := NewRequest(t, "GET", "/", &bytes.Buffer{})
		resp := httptest.NewRecorder()

		Response(resp, req).Data(func() (interface{}, Error) {
			return tc.Data, nil
		})

		if resp.Code != http.StatusOK {
			t.Fatal("Expected %v was %v", http.StatusOK, resp.Code)
		}

		EqualJSONBody(t, tc.ExpctedBody, resp.Body)
	}

}

func Test_ResponseDataError(t *testing.T) {
	tc := struct {
		ExpectedBodyMessage string
		ExpectedLogMessage  string
	}{
		ExpectedBodyMessage: "No love in here",
		ExpectedLogMessage:  "Love me hard",
	}

	// Run test
	{
		logger, mock := test.NewNullLogger()
		Logger = logger

		req := NewRequest(t, "GET", "/", &bytes.Buffer{})
		req.RemoteAddr = "127.0.0.1"
		resp := httptest.NewRecorder()

		Response(resp, req).Data(func() (interface{}, Error) {
			return nil, NewError("No love in here", errors.New(tc.ExpectedLogMessage))
		})

		d := mock.LastEntry().Data
		if d["remote_addr"] != "127.0.0.1" ||
			d["url"] != "/" ||
			d["method"] != "GET" ||
			d["body"] != "" ||
			mock.LastEntry().Message != tc.ExpectedLogMessage {
			t.Fatalf("Expected (%v, %v, %v, %v, %v) was (%v, %v, %v, %v, %v)",
				"127.0.0.1", "/", "GET", tc.ExpectedBodyMessage, tc.ExpectedLogMessage,
				d["remote_addr"], d["url"], d["method"], d["body"], mock.LastEntry().Message,
			)
		}

		if resp.Code != http.StatusBadRequest {
			t.Fatal("Expected %v was %v", http.StatusBadRequest, resp.Code)
		}

		b := fmt.Sprintf(`
		{
			"id":"\w*",
			"message": "%v"
		}
		`, tc.ExpectedBodyMessage)
		EqualJSONBody(t, b, resp.Body)
	}

}

func Test_ResponseDataWithStatusCode(t *testing.T) {
	tc := struct {
		Data               map[string]string
		ExpectedBody       string
		ExpectedStatusCode int
	}{
		Data: map[string]string{
			"key": "value",
		},
		ExpectedBody: `
		{
			"key": "value"
		}
		`,
		ExpectedStatusCode: http.StatusCreated,
	}

	// Run test
	{
		req := NewRequest(t, "GET", "/", &bytes.Buffer{})
		resp := httptest.NewRecorder()

		Response(resp, req).
			StatusCode(tc.ExpectedStatusCode).
			Data(func() (interface{}, Error) {
				return tc.Data, nil
			})

		if resp.Code != tc.ExpectedStatusCode {
			t.Fatal("Expected %v was %v", tc.ExpectedStatusCode, resp.Code)
		}

		EqualJSONBody(t, tc.ExpectedBody, resp.Body)
	}

}

func Test_EmptyResponseOK(t *testing.T) {
	tc := struct {
		ExpectedBody       string
		ExpectedStatusCode int
	}{
		ExpectedBody:       ``,
		ExpectedStatusCode: http.StatusOK,
	}

	// Run test
	{
		req := NewRequest(t, "GET", "/", &bytes.Buffer{})
		resp := httptest.NewRecorder()

		Response(resp, req).OK()

		if resp.Code != tc.ExpectedStatusCode {
			t.Fatal("Expected %v was %v", tc.ExpectedStatusCode, resp.Code)
		}

		EqualJSONBody(t, tc.ExpectedBody, resp.Body)
	}
}

func Test_PostResponse(t *testing.T) {
	tc := struct {
		Data               map[string]string
		ExpectedBody       string
		ExpectedStatusCode int
	}{
		Data: map[string]string{
			"key": "value",
		},
		ExpectedBody: `
		{
			"key": "value"
		}
		`,
		ExpectedStatusCode: http.StatusCreated,
	}

	// Run test
	{
		req := NewRequest(t, "GET", "/", &bytes.Buffer{})
		resp := httptest.NewRecorder()

		Response(resp, req).Post(func() (interface{}, Error) {
			return tc.Data, nil
		})

		if resp.Code != tc.ExpectedStatusCode {
			t.Fatal("Expected %v was %v", tc.ExpectedStatusCode, resp.Code)
		}

		EqualJSONBody(t, tc.ExpectedBody, resp.Body)
	}
}

func Test_NewLogID(t *testing.T) {
	i := map[string]bool{}

	for x := 0; x < 100000; x++ {
		k, err := newLogID()
		if err != nil {
			t.Fatal(err)
		}
		_, ok := i[k]
		if ok {
			t.Fatalf("%v: %v already exists", x, k)
		}

		i[k] = true
	}
}

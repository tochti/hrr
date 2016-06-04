package hrr

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Sirupsen/logrus/hooks/test"
	"github.com/gin-gonic/gin"
	"github.com/julienschmidt/httprouter"
)

type TestObject struct {
	Value string `json:"value" validate:"required"`
}

func Test_LogRequestManual(t *testing.T) {

	tc := struct {
		Method     string
		URL        string
		RemoteAddr string
		Body       string
	}{
		Method:     "POST",
		URL:        "/",
		RemoteAddr: "192.168.1.1",
		Body:       "test body",
	}

	// Run Test
	{
		tmp, mock := test.NewNullLogger()
		Logger = tmp

		LogAllRequests = false

		req := NewRequest(t, tc.Method, tc.URL, bytes.NewBufferString(tc.Body))
		req.RemoteAddr = tc.RemoteAddr

		err := Request(req).Log().Process()
		if err != nil {
			t.Fatal(err)
		}

		le := mock.LastEntry().Data
		if tc.Method != le["method"] ||
			tc.Body != le["body"] ||
			tc.URL != le["url"] ||
			tc.RemoteAddr != le["remote_addr"] {
			t.Fatalf("Expected %v was %v", tc, le)
		}
	}

}

func Test_LogRequestGlobal(t *testing.T) {
	tc := struct {
		Method     string
		URL        string
		RemoteAddr string
		Body       string
	}{
		Method:     "POST",
		URL:        "/",
		RemoteAddr: "192.168.1.1",
		Body:       "test body",
	}

	// Run test
	{
		tmp, mock := test.NewNullLogger()
		Logger = tmp

		LogAllRequests = true

		req := NewRequest(t, tc.Method, tc.URL, bytes.NewBufferString(tc.Body))
		req.RemoteAddr = tc.RemoteAddr

		err := Request(req).Process()
		if err != nil {
			t.Fatal(err)
		}

		le := mock.LastEntry().Data
		if tc.Method != le["method"] ||
			tc.Body != le["body"] ||
			tc.URL != le["url"] ||
			tc.RemoteAddr != le["remote_addr"] {
			t.Fatalf("Expected %v was %v", tc, le)
		}
	}
}

func Test_DecodeBody(t *testing.T) {
	tc := struct {
		Body     string
		Expected struct {
			A string
		}
	}{
		Body: `
		{
			"A": "a"
		}
		`,
		Expected: struct {
			A string
		}{
			A: "a",
		},
	}

	// Run test
	{
		req := NewRequest(t, "POST", "/", bytes.NewBufferString(tc.Body))

		body := struct {
			A string
		}{}
		err := Request(req).DecodeBody(&body).Process()
		if err != nil {
			t.Fatal(err)
		}

		if tc.Expected.A != body.A {
			t.Fatalf("Expected %v was %v", tc, body)
		}
	}
}

func Test_ParamInt64httprouter(t *testing.T) {
	tc := struct {
		Name     string
		Var      string
		URL      string
		Expected int64
	}{
		Name:     "id",
		Var:      ":id",
		URL:      "/1",
		Expected: 1,
	}

	// Run test
	{

		router := httprouter.New()
		router.GET("/"+tc.Var, func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
			var result int64
			err := Request(r).ParamInt64(p, tc.Name, &result).Process()
			if err != nil {
				t.Fatal(err)
			}

			if result != tc.Expected {
				t.Fatalf("Expect %v was %v", tc.Expected, result)
			}

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(""))
		})

		srv := httptest.NewServer(router)

		u := fmt.Sprintf("%v%v", srv.URL, tc.URL)
		_, err := http.Get(u)
		if err != nil {
			t.Fatal(err)
		}

	}
}

func Test_ParamInt64httprouterError(t *testing.T) {
	tc := struct {
		Name     string
		Var      string
		URL      string
		Expected string
	}{
		Name:     "id",
		Var:      ":id",
		URL:      "/a",
		Expected: "Error cannot find int64 parameter id",
	}

	// Run test
	{

		router := httprouter.New()
		router.GET("/"+tc.Var, func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
			var result int64
			err := Request(r).ParamInt64(p, tc.Name, &result).Process()

			if err.Message() != tc.Expected {
				t.Fatalf("Expect %v was %v", tc.Expected, err.Message())
			}

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(""))
		})

		srv := httptest.NewServer(router)

		u := fmt.Sprintf("%v%v", srv.URL, tc.URL)
		_, err := http.Get(u)
		if err != nil {
			t.Fatal(err)
		}

	}
}

func Test_ParamInt64gin(t *testing.T) {
	tc := struct {
		Name     string
		Var      string
		URL      string
		Expected int64
	}{
		Name:     "id",
		Var:      ":id",
		URL:      "/1",
		Expected: 1,
	}

	// Run test
	{

		router := gin.New()
		router.GET("/"+tc.Var, func(c *gin.Context) {
			var result int64
			err := Request(c.Request).ParamInt64(c.Params, tc.Name, &result).Process()
			if err != nil {
				t.Fatal(err)
			}

			if result != tc.Expected {
				t.Fatalf("Expect %v was %v", tc.Expected, result)
			}

			c.JSON(http.StatusOK, nil)
		})

		srv := httptest.NewServer(router)

		u := fmt.Sprintf("%v%v", srv.URL, tc.URL)
		_, err := http.Get(u)
		if err != nil {
			t.Fatal(err)
		}

	}
}

func Test_ParamInt64ErrorNotSupportedStore(t *testing.T) {
	tc := struct {
		Name     string
		Expected string
	}{
		Name:     "id",
		Expected: "Not supported parameter store",
	}

	// Run test
	{

		var result int64
		req := NewRequest(t, "GET", "/", &bytes.Buffer{})
		err := Request(req).ParamInt64(nil, tc.Name, &result).Process()
		if err.Message() != tc.Expected {
			t.Fatalf("Expect %v was %v", tc.Expected, err.Message())
		}

	}
}

func Test_ValidateBody(t *testing.T) {
	tc := struct {
		Body string
	}{
		Body: `
		{
			"value": "a"
		}
		`,
	}

	// Run test
	{
		req := NewRequest(t, "POST", "/", bytes.NewBufferString(tc.Body))

		var body TestObject
		err := Request(req).DecodeBody(&body).ValidateBody().Process()
		if err != nil {
			t.Fatal(err)
		}

	}
}

func Test_ValidateBodyError(t *testing.T) {
	tc := struct {
		Body     string
		Expected string
	}{
		Body: `
		{
			"value": ""
		}
		`,
		Expected: "Value failed due to required",
	}

	// Run test
	{
		req := NewRequest(t, "POST", "/", bytes.NewBufferString(tc.Body))

		var body TestObject
		err := Request(req).DecodeBody(&body).ValidateBody().Process()
		if err.Message() != tc.Expected {
			t.Fatal(err)
		}

	}
}

func Test_BaseAuth(t *testing.T) {
	tc := struct {
		User     string
		Password string
	}{
		User:     "Deadpool",
		Password: ":)",
	}

	// Run test
	{
		req := NewRequest(t, "GET", "/", bytes.NewBufferString(""))
		req.SetBasicAuth(tc.User, tc.Password)

		err := Request(req).BaseAuth(func(user, pass string) (bool, error) {
			if user != tc.User ||
				pass != tc.Password {
				t.Fatalf("Expect %v was {%v, %v}", tc, user, pass)
			}

			return true, nil
		}).Process()

		if err != nil {
			t.Fatal(err)
		}
	}
}

func Test_PostRequest(t *testing.T) {
	tc := struct {
		Body     string
		Expected TestObject
	}{
		Body: `
		{
			"Value": "v"
		}
		`,
		Expected: TestObject{
			Value: "v",
		},
	}

	// Run test
	{
		req := NewRequest(t, "POST", "/", bytes.NewBufferString(tc.Body))

		body := TestObject{}
		err := Request(req).Post(&body).Process()
		if err != nil {
			t.Fatal(err)
		}

		if tc.Expected.Value != body.Value {
			t.Fatalf("Expected %v was %v", tc, body)
		}

	}
}

func Test_PutRequest(t *testing.T) {
	tc := struct {
		Body           string
		Name           string
		URL            string
		ExpectedID     int64
		ExpectedObject TestObject
	}{
		Body: `
		{
			"Value": "v"
		}
		`,
		Name:       "id",
		URL:        "/1",
		ExpectedID: 1,
		ExpectedObject: TestObject{
			Value: "v",
		},
	}

	// Run test
	{
		req := NewRequest(t, "PUT", tc.URL, bytes.NewBufferString(tc.Body))

		router := httprouter.New()
		router.PUT("/:"+tc.Name, func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
			var id int64
			var body TestObject
			err := Request(req).Put(&body, p, tc.Name, &id).Process()
			if err != nil {
				t.Fatal(err)
			}

			if tc.ExpectedObject.Value != body.Value {
				t.Fatalf("Expected %v was %v", tc, body)
			}

			if tc.ExpectedID != id {
				t.Fatalf("Expected %v was %v", tc.ExpectedID, id)
			}
		})

		router.ServeHTTP(httptest.NewRecorder(), req)

	}
}

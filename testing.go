package hrr

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
)

type (
	FatalMethods interface {
		Fatal(args ...interface{})
		Fatalf(format string, args ...interface{})
	}
)

func NewRequest(t FatalMethods, method, url string, body io.Reader) *http.Request {
	r, err := http.NewRequest(method, url, body)
	if err != nil {
		t.Fatal(err)
	}

	return r
}

// Prüft ob ein ResponseRecoder Body mit einem übergeben String übereinstimmt
// In dem übergeben String können Regulärer Ausdrücke verwendet werden.
func EqualJSONBody(t FatalMethods, expected string, body *bytes.Buffer) {
	e := SimplifyJSON(t, expected)
	re := regexp.MustCompile(e)
	if !re.Match(body.Bytes()) {
		t.Fatalf("Expected body %v was %v", e, body.String())
	}
}

// Entfernt unötige whitespaces
func SimplifyJSON(t FatalMethods, j string) string {
	j = strings.Replace(j, "\n", "", -1)
	j = strings.Replace(j, "\t", "", -1)
	j = strings.Replace(j, " \"", "\"", -1)
	j = strings.Replace(j, "\" ", "\"", -1)
	j = strings.Replace(j, "\": ", "\":", -1)

	return j
}

// Teste einfachen POST Handler
func TestJSONPost(t FatalMethods, bodyObject interface{}, url, expected, body string, router http.Handler) {
	req := NewRequest(t, "POST", url, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	EqualJSONBody(t, expected, resp.Body)

	if resp.Code != http.StatusCreated {
		t.Fatalf("Expected status %v was %v", http.StatusCreated, resp.Code)
	}

	err := json.Unmarshal(resp.Body.Bytes(), bodyObject)
	if err != nil {
		t.Fatal(err)
	}

}

// Test einfachen GET Handler
func TestJSONGet(t FatalMethods, url, expected string, router http.Handler) {
	req := NewRequest(t, "GET", url, &bytes.Buffer{})
	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	EqualJSONBody(t, expected, resp.Body)

	if resp.Code != http.StatusOK {
		t.Fatalf("Expected status %v was %v", http.StatusOK, resp.Code)
	}

}

// Test einfachen PUT Handler
func TestJSONPut(t FatalMethods, bodyObject interface{}, url, expected, body string, router http.Handler) {
	req := NewRequest(t, "PUT", url, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	EqualJSONBody(t, expected, resp.Body)

	if resp.Code != http.StatusOK {
		t.Fatalf("Expected status %v was %v", http.StatusOK, resp.Code)
	}

	err := json.Unmarshal(resp.Body.Bytes(), bodyObject)
	if err != nil {
		t.Fatal(err)
	}
}

// Test einfachen DELETE Handler
func TestDelete(t FatalMethods, url string, router http.Handler) {
	req := NewRequest(t, "DELETE", url, &bytes.Buffer{})
	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("Expected status %v was %v", http.StatusOK, resp.Code)
	}
}

package hrr

import (
	"net/http"
	"testing"

	"github.com/julienschmidt/httprouter"
)

func Test_TestJSONPOSTHandler(t *testing.T) {
	tc := struct {
		Body     string
		Expected string
	}{
		Body: `
		{
			"Name": "MK-I"
		}
		`,
		Expected: `
		{
			"ID":\d*,
			"Name": "MK-I"
		}
		`,
	}

	// Run test
	{
		router := httprouter.New()
		router.POST("/", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
			suite := struct {
				Name string
			}{}

			if err := Request(r).Post(&suite).Process(); err != nil {
				t.Fatal(err)
			}

			Response(w, r).Post(func() (interface{}, Error) {
				return struct {
					ID   int64
					Name string
				}{
					ID:   1,
					Name: suite.Name,
				}, nil
			})
		})

		suite := struct {
			Name string
		}{}
		TestJSONPost(t, &suite, "/", tc.Expected, tc.Body, router)
		if suite.Name != "MK-I" {
			t.Fatalf("Expect %v was %v", "MK-I", suite.Name)
		}
	}
}

func Test_TestJSONGetHandler(t *testing.T) {
	tc := struct {
		Object   TestObject
		Expected string
	}{
		Object: TestObject{
			Value: "MK-I",
		},
		Expected: `
		{
			"value": "MK-I"
		}
		`,
	}

	// Run test
	{
		router := httprouter.New()
		router.GET("/:id", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
			var id int64
			if err := Request(r).ParamInt64(p, "id", &id).Process(); err != nil {
				t.Fatal(err)
			}

			Response(w, r).Data(func() (interface{}, Error) {
				return tc.Object, nil
			})
		})

		TestJSONGet(t, "/1", tc.Expected, router)
	}
}

func Test_TestJSONPutHandler(t *testing.T) {
	tc := struct {
		Object   TestObject
		Body     string
		Expected string
	}{
		Object: TestObject{
			Value: "MK-I",
		},
		Body: `
		{
			"value": "MK-I"
		}
		`,
		Expected: `
		{
			"value": "MK-I"
		}
		`,
	}

	// Run test
	{
		router := httprouter.New()
		router.PUT("/", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
			Response(w, r).Data(func() (interface{}, Error) {
				return tc.Object, nil
			})
		})

		var o TestObject
		TestJSONPut(t, &o, "/", tc.Expected, tc.Body, router)
		if o.Value != tc.Object.Value {
			t.Fatalf("Expect %v was %v", tc.Object.Value, o.Value)
		}
	}
}

func Test_TestDeleteHandler(t *testing.T) {
	router := httprouter.New()
	router.DELETE("/", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		Response(w, r).OK()
	})

	TestDelete(t, "/", router)
}

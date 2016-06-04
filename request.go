// Funktionen um HTTP Requests zu vearbeiten und HTTP Responses zu erstellen. Das Ziel ist es schneller REST APIs zuentwickeln. Es werden momentan nur JSON Bodys verarbeitet. Ebenso werden auch nur JSON Bodys zurückgesendet.
package hrr

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"gopkg.in/go-playground/validator.v8"

	"github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
	"github.com/julienschmidt/httprouter"
)

type (
	authFunc func(string, string) (bool, error)

	request struct {
		request            *http.Request
		logger             *logrus.Logger
		bodyObject         interface{}
		body               []byte
		paramsInt64        []*paramInt64
		enableValidateBody bool
		authFunc           authFunc
	}

	requestError struct {
		message string
		err     error
	}

	paramInt64 struct {
		data interface{}
		name string
		i64  *int64
	}

	Error interface {
		Message() string
		error
	}
)

// Globale Konfiguration
var (
	Logger           = logrus.New()
	LogAllRequests   = false
	ValidatorTagName = "validate"
)

// Konfiguriere neues request Objekt
func Request(r *http.Request) *request {
	requestLogger := newLogger(Logger)
	requestLogger.Out = ioutil.Discard

	request := &request{
		request:    r,
		logger:     requestLogger,
		bodyObject: nil,
		authFunc:   allIn,
	}

	if LogAllRequests {
		request.Log()
	}

	return request

}

// Request soll geloggt werden
func (r *request) Log() *request {
	r.logger.Out = Logger.Out
	return r
}

// Führe alle definierten Funktionen für den übergeben Request aus
func (r *request) Process() Error {
	if err := r.auth(); err != nil {
		return err
	}

	r.setBody()
	r.logRequest()

	if r.bodyObject != nil {
		if err := r.decodeBody(); err != nil {
			return err
		}

		if r.enableValidateBody {
			err := r.validateBody()
			if err != nil {
				return err
			}
		}
	}

	for _, v := range r.paramsInt64 {
		err := r.queryParamInt64(v)
		if err != nil {
			return err
		}
	}

	return nil
}

// Setze pointer zu Objekt um beim späteren decoden des JSON Bodys Daten darin abzulegen.
func (r *request) DecodeBody(body interface{}) *request {
	r.bodyObject = body
	return r
}

// Setze Daten um später params aus URL zu lesen
func (r *request) ParamInt64(p interface{}, name string, i *int64) *request {
	r.paramsInt64 = append(r.paramsInt64, &paramInt64{
		data: p,
		name: name,
		i64:  i,
	})

	return r
}

// Aktiviere Body Daten überprüfung
func (r *request) ValidateBody() *request {
	r.enableValidateBody = true

	return r
}

// Aktiviere Base Auth
func (r *request) BaseAuth(fn authFunc) *request {
	r.authFunc = fn

	return r
}

// Shortcut Post
func (r *request) Post(model interface{}) *request {
	return r.DecodeBody(model).ValidateBody()
}

// Shortcut Put
func (r *request) Put(model interface{}, p interface{}, n string, i *int64) *request {
	return r.DecodeBody(model).ValidateBody().ParamInt64(p, n, i)
}

// Decode JSON Body in übergebenes Object
func (r *request) decodeBody() Error {
	err := json.Unmarshal(r.body, r.bodyObject)
	if err != nil {
		return NewError("Error cannot parse JSON body", err)
	}

	return nil

}

// Find Parameter in url
func (r *request) queryParamInt64(v *paramInt64) Error {
	switch v.data.(type) {
	case httprouter.Params:
		return queryHTTPRouterParams(v)
	case gin.Params:
		return queryGinParams(v)
	default:
		err := errors.New("Not supported parameter store")
		return NewError(err.Error(), err)
	}

	return nil
}

// Find Parameter in url httprouter
func queryHTTPRouterParams(v *paramInt64) Error {
	params := v.data.(httprouter.Params)
	tmp := params.ByName(v.name)
	i, err := strconv.ParseInt(tmp, 10, 64)
	if err != nil {
		msg := fmt.Sprintf("Error cannot find int64 parameter %v", v.name)
		return NewError(msg, err)
	}

	p := v.i64
	*p = i

	return nil
}

// Find Parameter in url
func queryGinParams(v *paramInt64) Error {
	params := v.data.(gin.Params)
	tmp := params.ByName(v.name)
	i, err := strconv.ParseInt(tmp, 10, 64)
	if err != nil {
		msg := fmt.Sprintf("Error cannot find int64 parameter %v", v.name)
		return NewError(msg, err)
	}

	p := v.i64
	*p = i

	return nil
}

// Log request
func (r *request) logRequest() {
	r.logger.WithFields(logrus.Fields{
		"remote_addr": r.request.RemoteAddr,
		"method":      r.request.Method,
		"url":         r.request.URL.String(),
		"body":        fmt.Sprintf("%s", r.body),
	}).Infoln()
}

func (r *request) validateBody() Error {
	config := &validator.Config{TagName: ValidatorTagName}
	validate := validator.New(config)
	errs := validate.Struct(r.bodyObject)
	if errs != nil {
		errMsgs := errs.(validator.ValidationErrors)
		err := bytes.NewBufferString("")
		for _, v := range errMsgs {
			msg := fmt.Sprintf("%v failed due to %v", v.Field, v.Tag)
			err.WriteString(msg)
		}

		return NewError(err.String(), fmt.Errorf("%v", err))
	}

	return nil
}

// Auth User
func (r *request) auth() Error {
	user, password, _ := r.request.BasicAuth()
	ok, err := r.authFunc(user, password)
	if !ok || err != nil {
		return NewError(fmt.Sprintf("Unauthorized user %v", user), err)
	}

	return nil
}

func (r *request) setBody() Error {
	tmp, err := ioutil.ReadAll(r.request.Body)
	if err != nil {
		return NewError("Error while reading Process Body", err)
	}

	r.body = tmp

	return nil
}

func newLogger(old *logrus.Logger) *logrus.Logger {
	logger := logrus.New()
	logger.Formatter = old.Formatter
	logger.Hooks = old.Hooks
	logger.Level = old.Level
	logger.Out = old.Out

	return logger
}

func allIn(string, string) (bool, error) {
	return true, nil
}

// Erzeuge neuen Fehler
// Es gibt zum einen die Nachricht die im JSON Body zurück gesendet wird
// weiter gibt es ein Fehler der geloggt wird.
func NewError(message string, err error) requestError {
	return requestError{
		message: message,
		err:     err,
	}
}

func (err requestError) Message() string {
	return err.message
}

func (err requestError) Error() string {
	return err.err.Error()
}

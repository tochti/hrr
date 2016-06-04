package hrr

import (
	crand "crypto/rand"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	mrand "math/rand"
	"net/http"
	"time"

	"github.com/Sirupsen/logrus"
)

type (
	response struct {
		response   http.ResponseWriter
		request    *http.Request
		statusCode int
		logger     *logrus.Logger
		logID      string
		body       []byte
	}

	errorResponse struct {
		ID      string `json:"id"`
		Message string `json:"message"`
	}
)

func Response(w http.ResponseWriter, r *http.Request) *response {
	logID, err := newLogID()

	resp := &response{
		response:   w,
		request:    r,
		statusCode: http.StatusOK,
		logger:     Logger,
		logID:      logID,
		body:       []byte("empty body"),
	}

	if err != nil {
		resp.logError(err)
	}

	return resp
}

func (r *response) Error(err Error) {
	r.response.WriteHeader(http.StatusBadRequest)

	r.logError(err)

	r.json(errorResponse{
		ID:      r.logID,
		Message: err.Message(),
	})
}

func (r *response) json(data interface{}) {
	body, err := json.Marshal(data)
	if err != nil {
		r.logError(err)
		return
	}

	_, err = r.response.Write(body)
	if err != nil {
		r.logError(err)
		return
	}
}

func (r *response) logError(err error) {

	l := r.logger.WithFields(logrus.Fields{
		"id":          r.logID,
		"remote_addr": r.request.RemoteAddr,
		"url":         r.request.URL.String(),
		"method":      r.request.Method,
	})

	body, e := ioutil.ReadAll(r.request.Body)
	if e != nil {
		if e == io.EOF {
			body = r.body
		} else {
			l.Error(e)
			l.Error(err)
			return
		}

	} else {
		r.body = body
	}

	l.WithFields(logrus.Fields{"body": string(body)}).Error(err)
}

func (r *response) Data(fn func() (interface{}, Error)) {
	data, err := fn()
	if err != nil {
		r.Error(err)
		return
	}

	r.response.WriteHeader(r.statusCode)
	r.json(data)
}

func (r *response) StatusCode(c int) *response {
	r.statusCode = c
	return r
}

func (r *response) OK() {
	r.response.WriteHeader(http.StatusOK)
}

func (r *response) Post(fn func() (interface{}, Error)) {
	r.StatusCode(http.StatusCreated).Data(fn)
}

func newLogID() (string, error) {
	total := 500
	hash := sha1.New()
	rb := make([]byte, total)
	n, err := crand.Read(rb)
	if n != total || err != nil {
		r := mrand.New(mrand.NewSource(int64(time.Now().Nanosecond())))
		s := r.Perm(total)
		for i, v := range s {
			rb[i] = byte(v)
		}
	}

	io.WriteString(hash, fmt.Sprintf("%s", rb))
	return fmt.Sprintf("%x", hash.Sum(nil)), err
}

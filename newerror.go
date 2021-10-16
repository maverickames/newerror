package newerror

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"time"
)

type newError interface {
	PriError() error
	PubError() error
	PubContext() string
}

type ErrResponse struct {
	// Internal Only
	newError newError `json:"-"`
	funcPC   uintptr  `json:"-"`
	funcFN   string   `json:"-"`
	funcLine int      `json:"-"`

	// Public
	StatusCode int    `json:"-"`
	ErrorMsg   string `json:"error"`
	Context    string `json:"context,omitempty"`
	AppCode    int64  `json:"code,omitempty"`
}

type ErrorHandlerHTTP func(http.ResponseWriter, *http.Request) *ErrResponse

var genericError string

type ErrManager struct {
	debug  bool
	logger *log.Logger

	// ErrorHandler     processError
	// LogError         handleErrorLog
	// HandleHTTPErrors respondToError
}

func (err ErrResponse) Error() string {
	if err.GetCustomErr().PriError() != nil {
		return err.GetCustomErr().PriError().Error()
	}
	return ""
}

func New(options ...func(*ErrManager)) *ErrManager {
	em := &ErrManager{}
	em.debug = false
	genericError = "Internal Error"
	for _, opt := range options {
		opt(em)
	}
	return em
}

func Debug(debug bool) func(*ErrManager) {
	return func(em *ErrManager) {
		em.debug = debug
	}
}

func Logger(logger *log.Logger) func(*ErrManager) {
	return func(em *ErrManager) {
		em.logger = logger
	}
}

func NewError(err newError) *ErrResponse {
	return newErr(err)
}

func StdErr(err error) *ErrResponse {
	// newErr := &genError{
	// 	priErr: err,
	// }
	return newErr(&genError{
		priErr: err,
	})
}

func (er ErrResponse) GetPriErr() error {
	return er.newError.PriError()
}

func (er ErrResponse) GetPubErr() error {
	return er.newError.PubError()
}

func (er ErrResponse) GetCustomErr() newError {
	return er.newError
}

func (er ErrResponse) ToJson() ([]byte, error) {
	jsonData, err := json.Marshal(er)
	if err != nil {
		return jsonData, err
	}
	return jsonData, nil
}

func (em *ErrManager) Panicln(err *ErrResponse) {
	var recErr string
	if err.GetPriErr() == nil {
		recErr = "error not recorded"
	} else {
		recErr = err.GetPriErr().Error()
	}
	em.logger.Panicln(err.buildError(recErr, em.debug))
}

func (em *ErrManager) Fatalln(err *ErrResponse) {
	var recErr string
	if err.GetPriErr() == nil {
		recErr = "error not recorded"
	} else {
		recErr = err.GetPriErr().Error()
	}
	em.logger.Fatalln(err.buildError(recErr, em.debug))
}

func (em *ErrManager) Println(err *ErrResponse) {
	var recErr string
	if err.GetPriErr() == nil {
		recErr = "error not recorded"
	} else {
		recErr = err.GetPriErr().Error()
	}
	em.logger.Println(err.buildError(recErr, em.debug))
}

func (err ErrResponse) buildError(recErr string, debug bool) string {
	if debug {
		return fmt.Sprintf(
			"\n  -- Function: %s\n  -- SourceFile: %s\n  -- LineNumber: %d\n  -- ErrorDetails: %v\n  -- Context: %s\n  -- ErrorCode: %d\n",
			runtime.FuncForPC(err.funcPC).Name(),
			err.funcFN,
			err.funcLine,
			recErr,
			err.Context,
			err.AppCode,
		)
	}
	return fmt.Sprintf(
		"\n  -- ErrorDetails: %v\n  -- Context: %s\n  -- ErrorCode: %d\n",
		recErr,
		err.Context,
		err.AppCode,
	)
}

func newErr(err newError) *ErrResponse {
	errResponse := ErrResponse{}
	errResponse.newError = err
	if err.PriError() == nil {
		errResponse.Context = "failed to pass private error in newError"
		return &errResponse
	}
	pc, fn, line := getDebugDiagonisotics(err.PriError())

	errResponse.funcPC = pc
	errResponse.funcFN = fn
	errResponse.funcLine = line

	if errResponse.StatusCode == 0 {
		errResponse.StatusCode = 400
	}

	if err.PubContext() == "" {
		errResponse.Context = "Request was not passed. Most likely to protect the data"
	} else {
		errResponse.Context = err.PubContext()
	}

	errResponse.AppCode = rand.New(rand.NewSource(time.Now().Unix())).Int63()

	if err.PubError() == nil {
		errResponse.ErrorMsg = fmt.Sprintf("%s - ErrorCode: %d", genericError, errResponse.AppCode)
	} else {
		errResponse.ErrorMsg = err.PubError().Error()
	}
	return &errResponse
}

func getDebugDiagonisotics(err error) (uintptr, string, int) {
	var pc uintptr
	var fn string
	var line int
	pc, fn, line, _ = runtime.Caller(3)
	return pc, fn, line
}

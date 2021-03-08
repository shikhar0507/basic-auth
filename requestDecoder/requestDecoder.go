package requestDecoder

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func init()  {

}

type Result struct {
	Status int
	Message string
}

func Decode(w http.ResponseWriter, r *http.Request,stucBody interface{}) Result {

	// check for headers
	if r.Header.Get("Content-Type") != "application/json" {
		return Result{Status: http.StatusBadRequest,Message: "Content-Type is not application/json"}
	}


	// set max request size
	r.Body = http.MaxBytesReader(w,r.Body,1048576)
	decoded := json.NewDecoder(r.Body)

	//disallow unknown fields from struct
	decoded.DisallowUnknownFields()

	err := decoded.Decode(stucBody)
	if err != nil {
		fmt.Println("request err",err)
		var sytanxError *json.SyntaxError
		var unmarshallError *json.UnmarshalTypeError
		switch {
		case errors.As(err, &sytanxError):
			return Result{Status: http.StatusBadRequest, Message: "Badly formatted json at position " + string(sytanxError.Offset)}
		case errors.Is(err,io.ErrUnexpectedEOF):
			return Result{Status: http.StatusBadRequest,Message: "Badly formatted json"}
		case errors.As(err,&unmarshallError):
			msg := fmt.Sprintf("Wrong type of value for field %q at position %d. Expected %s but found %q",unmarshallError.Field,unmarshallError.Offset,unmarshallError.Type,unmarshallError.Value)
			return Result{Status: http.StatusBadRequest,Message: msg}
		case strings.HasPrefix(err.Error(),"json: unknown field"):
			unknownFieldName := strings.TrimPrefix(err.Error(),"json:unknown field")
			return Result{Status: http.StatusBadRequest,Message: fmt.Sprintf("Unknown field %s",unknownFieldName)}
		case errors.Is(err,io.EOF):
			return Result{Status: http.StatusBadRequest,Message:"Request body cannot be empty"}
		case err.Error() == "http: request body too large":
			return  Result{Status: http.StatusBadRequest,Message: "Request body too large"}
		default:
			return  Result{Status: 500}
		}
	}
	err = decoded.Decode(&struct{}{})
	if err != io.EOF {
		return Result{Status: http.StatusBadRequest,Message: "Request body should only contain a single json field"}
	}
	return  Result{Status: 200,Message: "success"}
}

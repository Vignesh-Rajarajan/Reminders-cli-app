package transport

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/vickey290/go-modules/server/models"
)

const (
	notFoundErrType         = "resource_not_found_error"
	dataValidationErrType   = "data_validation_error"
	formatValidationErrType = "format_validation_type"
	invalidJSONErrType      = "invalid_json_error"
	serviceErrType          = "service_error"
)

func SendJSON(w http.ResponseWriter, response interface{}, code int) {
	encoder := jsonEncoder(w, code)

	if err := encoder.Encode(response); err != nil {
		log.Printf("could not encode error %v", err)
	}
}

func SendError(w http.ResponseWriter, err error) {
	e := toHTTPError(err)
	encode := jsonEncoder(w, e.Code)
	if err := encode.Encode(e); err != nil {
		log.Printf("could not encode error: %v", err)
	}
}

func jsonEncoder(w http.ResponseWriter, code int) *json.Encoder {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	return json.NewEncoder(w)
}

func toHTTPError(err error) models.HTTPError {
	resErr := models.HTTPError{Message: err.Error()}
	switch e := err.(type) {
	case models.HTTPError:
		return e
	case models.NotFoundError:
		resErr.Code = http.StatusNotFound
		resErr.Type = notFoundErrType
	case models.FormatValidationError:
		resErr.Code = http.StatusBadRequest
		resErr.Type = formatValidationErrType
	case models.DataValidationError:
		resErr.Code = http.StatusNotFound
		resErr.Type = dataValidationErrType
	case models.InvalidJSONError:
		resErr.Code = http.StatusNotFound
		resErr.Type = invalidJSONErrType
	default:
		resErr.Code = http.StatusInternalServerError
		resErr.Type = serviceErrType
		resErr.Message = "Internal Server Error"
	}
	log.Printf("error: %v", err)
	return resErr
}

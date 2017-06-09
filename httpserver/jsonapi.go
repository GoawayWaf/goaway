package httpserver

import (
	_ "reflect"
	"regexp"
	"math/rand"
	"encoding/json"
	"strconv"
)

type anyStruct interface{}

var noDotsRegex = regexp.MustCompile(`\.`)

type jsonResponse struct {
	Data   []*respData `json:"data"`
	Errors []*respError `json:"errors"`
}

type respData struct {
	Id         interface{} `json:"id"`
	DataType   string `json:"data_type"`
	Attributes interface{} `json:"attributes"`
	//relationships
}

type respError struct {
	Id      int  `json:"id"`
	Title   string `json:"title"`
	Details string `json:"details"`
	Status  int `json:"status"`
}
func newRespData(id interface{}, dataType string, attributes interface{}) (*respData, error){
	return &respData{Id:id, DataType:dataType, Attributes:attributes}, nil
}
func newRespErr(title string, details string, status int) (*respError, error){
	return &respError{Id:rand.Int(), Title:title, Details:details, Status:status}, nil
}

func (r *jsonResponse)AddData(data *respData) {
	r.Data = append(r.Data, data)
}
func (r *jsonResponse)AddError(error *respError) {
	r.Errors = append(r.Errors, error)
}

func (r *respError) MarshalJSON() ([]byte, error) {
	type Alias respError
	return json.Marshal(&struct {
		Status string `json:"status"`
		Ttl    string `json:"ttl"`
		*Alias
	}{
		Status: strconv.Itoa(r.Status),
		Alias:    (*Alias)(r),
	})
}
func (r *respError) UnmarshalJSON(data []byte) error {
	type Alias respError
	var err error
	aux := &struct {
		Status string `json:"status"`
		*Alias
	}{
		Alias: (*Alias)(r),
	}
	if err = json.Unmarshal(data, &aux); err != nil {
		return err
	}
	intStat, err := strconv.ParseInt(aux.Status, 10, 0)
	r.Status = int(intStat)
	return err
}
 /*
func (r *jsonResponse) MarshalJSON() ([]byte, error) {

	for _,v  := range r.data {

	}

        errs := []error{}

	// ValueOf returns a Value representing the run-time data
	v := reflect.ValueOf(r)

	for i := 0; i < v.NumField(); i++ {
		// Get the field tag value
		tag := v.Type().Field(i).Tag.Get(tagName)

		// Skip if tag is not defined or ignored
		if tag == "" || tag == "-" {
			continue
		}
	}

	return errs
}
*/

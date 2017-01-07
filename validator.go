package tpi_data

import (
	"reflect"
	"time"
)

var (
	validatorType = reflect.TypeOf(new(Validator)).Elem()
)

type validatorFunc func(u reflect.Value, vv ...interface{}) error

type Validator interface {
	Validate(v ...interface{}) error
}

func Validate(u interface{}, v ...interface{}) error {

	uu := reflect.ValueOf(u)
	err := selectValidatorType(uu.Type())(uu, v...)
	return err

}

func selectValidatorType(t reflect.Type) validatorFunc {

	if t.Implements(validatorType) {
		return selectValidator
	} else {
		return errorValidator
	}

}

func selectValidator(u reflect.Value, v ...interface{}) error {

	if u.Kind() == reflect.Ptr && u.IsNil() {
		return DSErr{When: time.Now(), What: "Nil Value"}
	}

	m := u.Interface().(Validator)
	err := m.Validate(v...)
	if err != nil {
		return DSErr{When: time.Now(), What: err.Error()}
	}
	return nil

}

func errorValidator(u reflect.Value, v ...interface{}) error {

	return DSErr{When: time.Now(), What: "Not a Validator"}

}

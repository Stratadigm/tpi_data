package tpi_data

import (
	_ "image"
	_ "image/jpeg"
	"reflect"
	"time"
)

type Thali struct {
	Id        int64     `json:"id"`
	Name      string    `json:"name"`
	Submitted time.Time `json:"submitted"`
	Target    string    `json:"target" schema:"target"` // 1-4 target customer profile
	Limited   bool      `json:"limited" schema:"limited"`
	Region    string    `json:"region" schema:"region"` // 1-3 target cuisine
	Price     float64   `json:"price" schema:"price"`
	//Photo     *image.RGBA `json:"image"`
	Photo    string `json:"image"`
	UserId   int64  `json:"userid"`
	VenueId  int64  `json:"venue" schema:"venue"`
	Verfied  bool   `json:"verified"`
	Accepted bool   `json:"accepted"`
}

func NewThali(id int64) *Thali {

	return &Thali{Id: id}

}

//Validate checks u for errors optionally against v. If len(v)==0 then all fields of u to be checked (CREATE). If len(v)==1, then only provided fields of u to be checked and empty fields to be replaced by those of v (UPDATE)
func (t *Thali) Validate(v ...interface{}) error {

	if len(v) == 0 { // CREATE
		name := t.Name
		if len(name) < 3 {
			return DSErr{When: time.Now(), What: "validate venue invalid name "}
		}
		if t.VenueId == int64(0) || t.UserId == int64(0) {
			return DSErr{When: time.Now(), What: "validate thali missing venue id "}
		}
		return nil
	} else { // UPDATE
		v0 := v[0]
		if reflect.TypeOf(v0).Kind() != reflect.Ptr {
			return DSErr{When: time.Now(), What: "validate thali pointer reqd"}
		}
		s := reflect.TypeOf(v0).Elem()
		if _, ok := entities[s.Name()]; !ok {
			return DSErr{When: time.Now(), What: "validate want thali got " + s.Name()}
		}
		switch s.Name() {
		case "Thali":
			v0v := reflect.ValueOf(v0).Elem()
			if v0v.Kind() != reflect.Struct {
				return DSErr{When: time.Now(), What: "validate needs struct arg got " + v0v.Kind().String()}
			}
			if t.Id == int64(-999) {
				if t.Name == v0v.FieldByName("Name").String() {
					return nil
				} else {
					return DSErr{When: time.Now(), What: "delete validate name " + t.Name}
				}
			}
			set := 0
			for i := 0; i < v0v.NumField(); i++ {
				uu := reflect.ValueOf(t).Elem()
				fi := uu.Field(i)
				vi := v0v.Field(i)
				if reflect.DeepEqual(fi.Interface(), reflect.Zero(fi.Type()).Interface()) {
					if vi.IsValid() {
						fi.Set(vi)
						set++
					}
				}
			}
			name := t.Name
			if len(name) < 3 {
				return DSErr{When: time.Now(), What: "validate venue invalid name "}
			}
			if t.VenueId == int64(0) || t.UserId == int64(0) {
				return DSErr{When: time.Now(), What: "validate thali missing venue id "}
			}
			return nil
		default:
			return DSErr{When: time.Now(), What: "validate want thali got :  " + s.Name()}
		}
		return nil
	}

}

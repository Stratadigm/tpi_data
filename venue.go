package tpi_data

import (
	"google.golang.org/appengine"
	"reflect"
	"time"
)

type Venue struct {
	Id        int64     `json:"id" schema:"-"`
	Name      string    `json:"name" schema:"name"`
	Submitted time.Time `json:"submitted"`
	//Latitude  float64 `json:"latitude" schema:"latitude"`
	//Longitude float64 `json:"longitude" schema:"longitude"`
	Location appengine.GeoPoint `json:"location" schema:"location"`
	//Thalis   []Thali            `json:"thalis"`
	Thalis []int64 `json:"thalis"`
}

func NewVenue(id int64) *Venue {

	return &Venue{Id: id}

}

//Validate checks u for errors optionally against v. If len(v)==0 then all fields of u to be checked (CREATE). If len(v)==1, then only provided fields of u to be checked and empty fields to be replaced by those of v (UPDATE)
func (u *Venue) Validate(v ...interface{}) error {

	if len(v) == 0 { // CREATE
		name := u.Name
		if len(name) < 3 {
			return DSErr{When: time.Now(), What: "validate venue invalid name "}
		}
		if u.Location.Lat < -90 || u.Location.Lat > 90 || u.Location.Lng < -90 || u.Location.Lng > 90 {
			return DSErr{When: time.Now(), What: "validate venue invalid location "}
		}
		return nil

	} else { // UPDATE
		v0 := v[0]
		if reflect.TypeOf(v0).Kind() != reflect.Ptr {
			return DSErr{When: time.Now(), What: "validate venue pointer reqd"}
		}
		s := reflect.TypeOf(v0).Elem()
		if _, ok := entities[s.Name()]; !ok {
			return DSErr{When: time.Now(), What: "validate want venue got " + s.Name()}
		}
		switch s.Name() {
		case "Venue":
			v0v := reflect.ValueOf(v0).Elem()
			if v0v.Kind() != reflect.Struct {
				return DSErr{When: time.Now(), What: "validate needs struct arg got " + v0v.Kind().String()}
			}
			if u.Id == int64(-999) {
				if u.Name == v0v.FieldByName("Name").String() {
					return nil
				} else {
					return DSErr{When: time.Now(), What: "delete validate name " + u.Name}
				}
			}
			set := 0
			for i := 0; i < v0v.NumField(); i++ {
				uu := reflect.ValueOf(u).Elem()
				fi := uu.Field(i)
				vi := v0v.Field(i)
				if reflect.DeepEqual(fi.Interface(), reflect.Zero(fi.Type()).Interface()) {
					if vi.IsValid() {
						fi.Set(vi)
						set++
					}
				}
			}
			name := u.Name
			if len(name) < 3 {
				return DSErr{When: time.Now(), What: "validate venue invalid name "}
			}
			if u.Location.Lat < -90 || u.Location.Lat > 90 || u.Location.Lng < -90 || u.Location.Lng > 90 {
				return DSErr{When: time.Now(), What: "validate venue invalid location "}
			}
			return nil
		default:
			return DSErr{When: time.Now(), What: "validate want venue got :  " + s.Name()}
		}
		return nil
	}

}

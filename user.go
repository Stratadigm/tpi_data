package tpi_data

import (
	"encoding/json"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"reflect"
	"strconv"
	"time"
)

type Counter struct {
	Users  int64 `json:"users"`
	Venues int64 `json:"venues"`
	Thalis int64 `json:"thalis"`
	Datas  int64 `json:"datas"`
}

type User struct {
	Id        int64  `json:"id" schema:"-"`
	Name      string `json:"name" schema:"fullname"`
	Email     string `json:"email" schema:"email"`
	Password  string `json:"password" schema:"password"`
	Confirmed bool   `json:"conf"`
	//Points    []Data    `json:"data"`
	Thalis    []Thali   `json:"thalis"`
	Venues    []int64   `json:"venues"`
	Rep       int       `json:"rep"`
	Submitted time.Time `json:"submitted"`
}

type UserDatabase interface {
	ListUsers() ([]*User, error)

	AddUser(guesty *User) (int64, error) //create

	GetUser(id int64) (*User, error) //retrieve by id

	GetUserwEmail(email string) (*User, error) //retrieve by email

	GetUserKey(email string) (*User, *datastore.Key, error)

	UpdateUser(guesty *User) error //update

	DeleteUser(id int64) error //delete

	Close() error
}

func NewUser(id int64) *User {

	return &User{Id: id, Submitted: time.Now(), Confirmed: false}

}

//MarshalJSON - use in Production only as testing requires JSON encoding with Password
func (u *User) MarshalJSON() ([]byte, error) {

	type Alias User
	if !appengine.IsDevAppServer() {
		//	u.Password = ""
	}
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(u),
	},
	)

}

//Validate checks u for errors optionally against v. If len(v)==0 then all fields of u to be checked (CREATE). If len(v)==1, then only provided fields of u to be checked and empty fields to be replaced by those of v (UPDATE)
func (u *User) Validate(v ...interface{}) error {

	if len(v) == 0 { // CREATE
		email := u.Email
		m := validEmail.FindStringSubmatch(email)
		if m == nil {
			return DSErr{When: time.Now(), What: "Validate error: invalid email " + email}
		}
		password := u.Password
		if len(password) < 6 {
			return DSErr{When: time.Now(), What: "Validate error: password too short " + password}
		} else {
			hash, _ := bcrypt.GenerateFromPassword([]byte(password), 10)
			u.Password = string(hash)
			return nil
		}

	} else { // UPDATE
		v0 := v[0]
		if reflect.TypeOf(v0).Kind() != reflect.Ptr {
			return DSErr{When: time.Now(), What: "Validate User pointer reqd"}
		}
		s := reflect.TypeOf(v0).Elem()
		if _, ok := entities[s.Name()]; !ok {
			return DSErr{When: time.Now(), What: "Validate want User got " + s.Name()}
		}
		switch s.Name() {
		case "User":
			v0v := reflect.ValueOf(v0).Elem()
			if v0v.Kind() != reflect.Struct {
				return DSErr{When: time.Now(), What: "Validate needs struct arg got " + v0v.Kind().String()}
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
			email := u.Email
			m := validEmail.FindStringSubmatch(email)
			if m == nil {
				return DSErr{When: time.Now(), What: "Validate invalid email " + email}
			}
			password := u.Password
			if len(password) < 6 {
				return DSErr{When: time.Now(), What: "validate pswrd too short " + u.Name + email + password + strconv.Itoa(set)}
			} else {
				hash, _ := bcrypt.GenerateFromPassword([]byte(password), 10)
				u.Password = string(hash)
				return nil
			}
		default:
			return DSErr{When: time.Now(), What: "Validate want User got :  " + s.Name()}
		}
		return nil
	}

}

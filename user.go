package tpi_data

import (
	"google.golang.org/appengine/datastore"
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

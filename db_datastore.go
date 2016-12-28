package tpi_data

import (
	"cloud.google.com/go/storage"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/file"
	"google.golang.org/appengine/log"
	"image"
	"image/jpeg"
	_ "log"
	"net/http"
	"reflect"
	"regexp"
	"time"
)

type DS struct {
	//Ctx appengine.Context
	Ctx context.Context
}

type DSErr struct {
	When time.Time `json:"when"`
	What string    `json:"what"`
}

var (
	_          UserDatabase = &DS{}
	_          error        = &DSErr{}
	entities                = map[string]string{"User": "user", "Venue": "venue", "Thali": "thali", "Data": "data"}
	total                   = int64(0)
	bucket                  = "thalipriceindex.appspot.com"
	validEmail              = regexp.MustCompile("^.*@.*\\.(com|org|in|mail|io)$")
	fake                    = `eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1MDAwLCJpc3MiOiJ0ZXN0In0.QsODzZu3lUZMVdhbO76u3Jv02iYCvEHcYVUI1kOWEU0`
)

const PerPage = 20

func NewDS(r *http.Request) *DS {

	return &DS{Ctx: appengine.NewContext(r)}

}

func NewDSwc(c context.Context) *DS {

	return &DS{Ctx: c}

}

//Crude ad-hoc key gen
func (ds *DS) datastoreKey(id int64) *datastore.Key {

	c := ds.Ctx
	return datastore.NewKey(c, "user", "", id, nil)

}

func (ds *DS) datastoreKeyah(entity string, id ...int64) *datastore.Key {

	c := ds.Ctx
	if len(id) != 0 {
		return datastore.NewKey(c, entity, "", id[0], nil)
	} else {
		return datastore.NewIncompleteKey(c, entity, nil)
	}

}

//Less crude interface key gen
func (ds *DS) dsKey(t reflect.Type, id ...interface{}) *datastore.Key {

	c := ds.Ctx
	if entity, ok := entities[t.Name()]; ok {
		switch t.Name() {
		case "Counter":
			return datastore.NewKey(c, entity, "thekey", 0, nil)
		default:
			if len(id) > 0 {
				return datastore.NewKey(c, entity, "", id[0].(int64), nil)
			} else {
				return datastore.NewIncompleteKey(c, entity, nil) // shouldn't get here
			}
		}
	}
	return nil

}

func (ds *DS) dsChildKey(t reflect.Type, id int64, pk *datastore.Key) *datastore.Key {

	c := ds.Ctx
	if entity, ok := entities[t.Name()]; ok {
		return datastore.NewKey(c, entity, "", id, pk)
	}
	return nil

}

/*func (ds *DS) Add(v interface{}) (int64, error) {

        c := ds.Ctx
        k := ds.dsKey(reflect.TypeOf(v).Elem(), reflect.ValueOf(v).Elem().Field(0).Interface())
        if k == nil {
                return 0, fmt.Errorf("Add usery error - key create error")
        }
        _, err := datastore.Put(c, k, v)
        if err != nil {
                return 0, fmt.Errorf("Add usery error - datastore put error")
        }
        return k.IntID(), nil

}*/

/* Interface */

//List returns a slice of v
func (ds *DS) List(v interface{}, offset ...int) error {

	if reflect.TypeOf(v).Kind() != reflect.Ptr {
		return DSErr{When: time.Now(), What: "Get error: pointer reqd"}
	}
	c := ds.Ctx
	//s := reflect.TypeOf(v).Elem()
	//cs := reflect.MakeSlice(s, 0, 1e6)
	entity := reflect.ValueOf(v).Elem().Slice(0, 1).Index(0).Type().Name() //v is &[]User
	q := datastore.NewQuery(entities[entity]).Order("Id")
	if offset[0] != 0 {
		q = q.Limit(PerPage + offset[0]).Offset(offset[0])
	} else {
		q = q.Limit(PerPage).Offset(0)
	}
	_, err := q.GetAll(c, v)
	if err != nil {
		//return nil, fmt.Errorf("Get %s list error", entity)
		return DSErr{When: time.Now(), What: fmt.Sprintf("Get %s list error: %v", entity, err)}
	}
	//for i, k := range ks {
	//	cs[i].Id = k.IntID()
	//}
	//reflect.ValueOf(v).Elem().Set(cs)
	return nil

}

//Validate checks whether the provided interface's fields is populated with valid data. Return nil or error
func (ds *DS) Validate(v interface{}) error {

	if reflect.TypeOf(v).Kind() != reflect.Ptr {
		return DSErr{When: time.Now(), What: "Validate error: pointer reqd"}
	}
	s := reflect.TypeOf(v).Elem()
	if _, ok := entities[s.Name()]; !ok {
		log.Errorf(ds.Ctx, "Validate entity no such entity: %v", s.Name())
		return DSErr{When: time.Now(), What: "Validate error: no such entity " + s.Name()}
	}
	switch s.Name() {
	case "User":
		email := reflect.ValueOf(v).Elem().FieldByName("Email").String()
		m := validEmail.FindStringSubmatch(email)
		if m == nil {
			log.Errorf(ds.Ctx, "Invalid email entered: %v\n", email)
			return DSErr{When: time.Now(), What: "Validate error: invalid email " + email}
		}
		user, _ := ds.GetUserwEmail(email)
		if user != nil { //|| err != nil {
			log.Errorf(ds.Ctx, "Email already in use: %v \n", email)
			return DSErr{When: time.Now(), What: "Validate error: email already in use " + email}
		}
		password := reflect.ValueOf(v).Elem().FieldByName("Password").String()
		if len(password) < 6 {
			log.Errorf(ds.Ctx, "Password too short: %v, %v\n", email, len(password))
			return DSErr{When: time.Now(), What: "Validate error: password too short " + password}
		} else {
			hash, _ := bcrypt.GenerateFromPassword([]byte(password), 10)
			reflect.ValueOf(v).Elem().FieldByName("Password").SetString(string(hash))
			return nil
		}
	case "Venue":
	case "Thali":
	case "Data":
	default:
		log.Errorf(ds.Ctx, "Validate entity no such entity: %v", s.Name())
		return DSErr{When: time.Now(), What: "Validate error: no such entity " + s.Name()}
	}
	return nil

}

//Create populates entity with appropriate fields including Id after updating counter. Returns nil if there was an error retrieving/updating counter or populating entity. Need to calls Add with Id to put in datastore.
func (ds *DS) Create(v interface{}) error {

	if reflect.TypeOf(v).Kind() != reflect.Ptr {
		return DSErr{When: time.Now(), What: "Create error: pointer reqd"}
	}
	s := reflect.TypeOf(v).Elem()
	//cs := reflect.New(s)
	if _, ok := entities[s.Name()]; !ok {
		log.Errorf(ds.Ctx, "Create entity no such entity: %v", s.Name())
		return DSErr{When: time.Now(), What: "Create error: no such entity " + s.Name()}
	}
	counter := ds.GetCounter()
	if counter != nil {
		switch s.Name() {
		case "User":
			counter.Users++
			reflect.ValueOf(v).Elem().FieldByName("Id").SetInt(counter.Users)
			//cs.Elem().FieldByName("Id").SetInt(counter.Users)
			//log.Errorf(ds.Ctx, "Creating datastore key: %v", cs.Elem().FieldByName("Id").Int())
		case "Venue":
			counter.Venues++
			reflect.ValueOf(v).Elem().FieldByName("Id").SetInt(counter.Venues)
			//cs.Elem().FieldByName("Id").SetInt(counter.Venues)
		case "Thali":
			counter.Thalis++
			reflect.ValueOf(v).Elem().FieldByName("Id").SetInt(counter.Thalis)
			//cs.Elem().FieldByName("Id").SetInt(counter.Thalis)
		case "Data":
			counter.Datas++
			reflect.ValueOf(v).Elem().FieldByName("Id").SetInt(counter.Datas)
			//cs.Elem().FieldByName("Id").SetInt(counter.Datas)
		default:
			log.Errorf(ds.Ctx, "Create entity no such entity: %v", s.Name())
			return DSErr{When: time.Now(), What: "Create error: no such entity " + s.Name()}
		}
		err := ds.PutCounter(counter)
		if err != nil {
			log.Errorf(ds.Ctx, "Create user Put counter: %v", err)
			return err
		}
		//v = cs.Interface()
		log.Errorf(ds.Ctx, "Check interface Id: %v", reflect.ValueOf(v).Elem().FieldByName("Id").Int())
		return nil
	} else {
		log.Errorf(ds.Ctx, "Create user nil counter: ")
		return DSErr{time.Now(), "Create entity nil counter"}
	}

}

// Add creates an appropriate ds key for the entity passed via interface{} with the optional int64 used as Id of key and puts into datastore
func (ds *DS) Add(v interface{}, n ...int64) (int64, error) {

	c := ds.Ctx
	var k *datastore.Key
	if len(n) == 0 { // for Counter
		//log.Errorf(c, "Adding datastore key: %v", reflect.ValueOf(v).Elem().FieldByName("Id").Int())
		//k = ds.dsKey(reflect.TypeOf(v).Elem(), reflect.ValueOf(v).Elem().FieldByName("Id").Int())
		k = ds.dsKey(reflect.TypeOf(v).Elem())
	} else { // if extra args are provided use as key ID
		k = ds.dsKey(reflect.TypeOf(v).Elem(), n[0])
	}
	if k == nil {
		return 0, fmt.Errorf("Add error - key create error - unknown entity")
	}
	_, err := datastore.Put(c, k, v)
	if err != nil {
		return 0, fmt.Errorf("Add error - datastore put error: %v", err)
	}
	return k.IntID(), nil

}

// AddwParent creates keys and adds an interface along with it's parent. Parent must have Id field
func (ds *DS) AddwParent(parent interface{}, child interface{}, offset int64) error {

	if reflect.TypeOf(parent).Kind() != reflect.Ptr || reflect.TypeOf(child).Kind() != reflect.Ptr {
		return DSErr{When: time.Now(), What: "Get error: pointers reqd"}
	}

	c := ds.Ctx

	pt := reflect.TypeOf(parent).Elem()
	if _, ok := pt.FieldByName("Id"); !ok {
		return DSErr{When: time.Now(), What: "Add w parent error: parent lacks Id"}
	}

	pv := reflect.ValueOf(parent).Elem().Field(0).Interface().(int64)
	pk := ds.dsKey(pt, pv)
	ck := ds.dsChildKey(reflect.TypeOf(child).Elem(), pv+offset, pk)

	if pk == nil || ck == nil {
		return DSErr{When: time.Now(), What: "Add w parent error: during key creation"}
	}
	pk, err := datastore.Put(c, pk, parent)
	if err != nil {
		return DSErr{When: time.Now(), What: "Add w parent error: during parent put" + err.Error()}
	}
	if pk != nil {
		_, err := datastore.Put(c, ck, child)
		if err != nil {
			return DSErr{When: time.Now(), What: "AddwParent error: during child put"}
		}
	}
	return nil

}

//Get retrieves from datastore the value of v which must be a pointer & have it's Id field set. Get populates the rest of the properties/fields of v.
func (ds *DS) Get(v interface{}) error {

	if reflect.TypeOf(v).Kind() != reflect.Ptr {
		return DSErr{When: time.Now(), What: "Get error: pointer reqd"}
	}

	var id int64
	var k *datastore.Key
	// check whether Id field is available in struct - if it is, it shouldn't be 0
	if _, ok := reflect.TypeOf(v).Elem().FieldByName("Id"); ok {
		id = reflect.ValueOf(v).Elem().Field(0).Interface().(int64) // could also use FieldByName("Id") instead of Field(0)
		if id == 0 {                                                // shouldn't be zero
			return DSErr{When: time.Now(), What: "Get error: id not set"}
		}
		k = ds.dsKey(reflect.TypeOf(v).Elem(), id) //complete key
	} else {
		return DSErr{When: time.Now(), What: "Get error: id not set"}
		//k = ds.dsKey(reflect.TypeOf(v).Elem())
	}

	c := ds.Ctx
	if k == nil {
		return fmt.Errorf("Get error - key create error")
	}

	//if err := datastore.Get(c, k, reflect.ValueOf(v).Interface()); err != nil {
	if err := datastore.Get(c, k, v); err != nil {
		return fmt.Errorf("Get error - datastore get error: %v, key kind: %v", err, k.Kind())
	}
	return nil

}

//Update updates the entity in the datastore. Must have Id field set. Returns nil (success) or error
func (ds *DS) Update(v interface{}) error {

	c := ds.Ctx
	k := ds.datastoreKey(reflect.ValueOf(v).FieldByName("Id").Int())
	_, err := datastore.Put(c, k, v)
	if err != nil {
		return fmt.Errorf("Updating error %v", err)
	}
	return nil

}

func (ds *DS) Delete(id int64) error {

	c := ds.Ctx
	k := ds.datastoreKey(id)
	if err := datastore.Delete(c, k); err != nil {
		return fmt.Errorf("Deleting error")
	}
	return nil

}

/* User specific */

//ListUsers returns a slice of *User
func (ds *DS) ListUsers() ([]*User, error) {

	c := ds.Ctx
	cs := make([]*User, 0)
	q := datastore.NewQuery("user").Order("Id")
	ks, err := q.GetAll(c, &cs)
	if err != nil {
		return nil, fmt.Errorf("Get usery list error")
	}
	for i, k := range ks {
		cs[i].Id = k.IntID()
	}
	return cs, nil

}

//AddUser adds user which must already have Id (from CreateUser) to be used as datastore key id. Doesn't touch counter. Returns either (id, nil) / (0, error)
func (ds *DS) AddUser(usery *User) (int64, error) {

	c := ds.Ctx
	k := ds.datastoreKey(usery.Id)
	_, err := datastore.Put(c, k, usery)
	if err != nil {
		return 0, err
	}
	return k.IntID(), nil

}

//GetUser uses Id to get and return User, nil (success) or nil, error
func (ds *DS) GetUser(id int64) (*User, error) {

	c := ds.Ctx
	k := ds.datastoreKey(id)
	cst := &User{}
	err := datastore.Get(c, k, cst)
	if err != nil {
		return nil, fmt.Errorf("Get by id error")
	}
	return cst, nil

}

//GetUserwEmail uses unique email to return User and nil (success) or nil, error
func (ds *DS) GetUserwEmail(email string) (*User, error) {

	c := ds.Ctx
	q := datastore.NewQuery("user").Filter("Email=", email)
	cst := make([]*User, 0)
	spk, err := q.GetAll(c, &cst) // *[]*User
	if err != nil {
		return nil, DSErr{When: time.Now(), What: "Get by email not found:" + email + err.Error()}
	}
	if len(spk) > 0 {
		return cst[0], nil
	}
	return nil, DSErr{When: time.Now(), What: "Get by email not found:" + email}

}

//GetUserKey uses unique email to return User, Key and nil (sucess) or nil, nil, error
func (ds *DS) GetUserKey(email string) (*User, *datastore.Key, error) {

	c := ds.Ctx
	q := datastore.NewQuery("user").Filter("Email=", email)
	cst := &User{}
	k, err := q.GetAll(c, cst)
	if err != nil {
		return nil, nil, fmt.Errorf("Get by email error %v", err)
	}
	return cst, k[0], nil

}

//UpdateUser puts the User in the datastore. Returns nil (success) or error
func (ds *DS) UpdateUser(usery *User) error {

	c := ds.Ctx
	k := ds.datastoreKey(usery.Id)
	_, err := datastore.Put(c, k, usery)
	if err != nil {
		return fmt.Errorf("Updating error %v", err)
	}
	return nil

}

func (ds *DS) DeleteUser(id int64) error {

	c := ds.Ctx
	k := ds.datastoreKey(id)
	if err := datastore.Delete(c, k); err != nil {
		return fmt.Errorf("Deleting error")
	}
	return nil

}

//Creates User and updates counter. Returns nil if there was an error retrieving/updating counter or creating User
func (ds *DS) CreateUser() *User {

	counter := ds.GetCounter()
	if counter != nil {
		counter.Users++
		err := ds.PutCounter(counter)
		if err != nil {
			log.Errorf(ds.Ctx, "Create user Put counter: %v", err)
			return nil
		}
		return NewUser(counter.Users)
	} else {
		return nil
	}
	//total++
	//return NewUser(total)

}

func (ds *DS) Close() error {

	return nil

}

func (dse DSErr) Error() string {

	return fmt.Sprintf("%v: %v", dse.When, dse.What)

}

//Location specific
func (ds *DS) AddLoc(loc *Loc) (int64, error) {

	c := ds.Ctx
	k := ds.datastoreKeyah("loc")
	_, err := datastore.Put(c, k, loc)
	if err != nil {
		return 0, fmt.Errorf("Add loc error")
	}
	return k.IntID(), nil

}

func (ds *DS) ListLocs() ([]*Loc, error) {

	c := ds.Ctx
	cs := make([]*Loc, 0)
	q := datastore.NewQuery("loc").Order("Ip")
	_, err := q.GetAll(c, &cs)
	if err != nil {
		return nil, fmt.Errorf("Get locations list error")
	}
	//for i, k := range ks {
	//	cs[i].Id = k.IntID()
	//}
	return cs, nil

}

/* Counter specific */

//Get Counter gets the singleton counter from datastore. Doesn't try to create. Returns the counter or nil
func (ds *DS) GetCounter() *Counter {

	c := ds.Ctx
	k := ds.datastoreKeyah("counter", 1234567890)
	counter := &Counter{}
	err := datastore.Get(c, k, counter)
	if err != nil {
		log.Errorf(c, "Couldn't get counter: %v", err)
		return nil
	}
	return counter

}

//CreateCounter creates a counter and returns nil (success) or error
func (ds *DS) CreateCounter() error {

	c := ds.Ctx
	counter := &Counter{Venues: int64(0), Datas: int64(1e9), Users: int64(1e7), Thalis: int64(1e6)}
	k := ds.datastoreKeyah("counter", 1234567890)
	_, err := datastore.Put(c, k, counter)
	if err != nil {
		log.Errorf(c, "Couldn't create counter: %v", err)
		return err
	}
	return nil

}

//PutCounter puts the counter in datastore and returns nil (success) or error
func (ds *DS) PutCounter(counter *Counter) error {

	c := ds.Ctx
	k := ds.datastoreKeyah("counter", 1234567890)
	_, err := datastore.Put(c, k, counter)
	if err != nil {
		log.Errorf(c, "Couldn't put counter: %v", err)
		return err
	}
	return nil

}

//WriteCloudImage writes the image provided as argument to cloud storage with name provided as argument
func WriteCloudImage(Ctx context.Context, mth *image.Image, filename string) error {

	var err error
	//[START get_default_bucket]
	if bucket == "" {
		if bucket, err = file.DefaultBucketName(Ctx); err != nil {
			log.Errorf(Ctx, "failed to get default GCS bucket name: %v\n", err.Error())
			return err
		}
	}
	//[END get_default_bucket]
	client, err := storage.NewClient(Ctx)
	if err != nil {
		log.Errorf(Ctx, "failed to create client: %v\n", err.Error())
		return err
	}
	defer client.Close()
	wc := client.Bucket(bucket).Object(filename).NewWriter(Ctx)
	wc.ContentType = "image/jpeg"
	wc.ACL = []storage.ACLRule{{storage.AllUsers, storage.RoleReader}}
	if err = jpeg.Encode(wc, *mth, nil); err != nil {
		log.Errorf(Ctx, "failed to write: %v\n", err.Error())
		return err
	}
	if err = wc.Close(); err != nil {
		log.Errorf(Ctx, "failed to close: %v\n", err.Error())
		return err
	}
	log.Errorf(Ctx, "updated object: %v\n", wc.Attrs())

	return err

}

//ReadCloudImage reads the jpeg file with filename as argument stored in GCS bucket
func ReadCloudImage(Ctx context.Context, filename string) (*image.Image, error) {

	var err error
	//[START get_default_bucket]
	if bucket == "" {
		if bucket, err = file.DefaultBucketName(Ctx); err != nil {
			log.Errorf(Ctx, "failed to get default GCS bucket name: %v\n", err.Error())
			return nil, err
		}
	}
	//[END get_default_bucket]
	client, err := storage.NewClient(Ctx)
	if err != nil {
		log.Errorf(Ctx, "failed to create client: %v\n", err.Error())
		return nil, err
	}
	defer client.Close()

	rc, err := client.Bucket(bucket).Object(filename).NewReader(Ctx)
	if err != nil {
		log.Errorf(Ctx, "readFile: unable to open file from bucket %q, file %q: %v", bucket, filename, err.Error())
		return nil, err
	}
	defer rc.Close()

	slurp, err := jpeg.Decode(rc)
	if err != nil {
		log.Errorf(Ctx, "readFile: unable to read data from bucket %q, file %q: %v", bucket, filename, err.Error())
		return &slurp, err
	}

	return &slurp, nil
}

//GetToken gets the token with string id provided as argument. Returns (nil) if token found and (DSErr) if token not found
func (ds *DS) GetToken(token string) error {

	c := ds.Ctx
	var err error

	k := datastore.NewKey(c, "token", token, 0, nil)
	tok := &AuthToken{token}
	err = datastore.Get(c, k, tok)
	if err != nil {
		if err == datastore.ErrNoSuchEntity {
			/*if err1 := ds.PutToken(fake); err1 != nil {
				log.Errorf(c, "GetToken put fake token %v \n", err1)
				return err1
			}
			return err1*/
			return err
		}
		log.Errorf(c, "GetToken datastore get: %v \n", err)
		return err
	}

	return nil

}

//PutToken puts the token into datastore after creating complete key with string id provided as argument. Returns nil if successfully put and DSErr if not
func (ds *DS) PutToken(token string) error {

	c := ds.Ctx
	var err error

	k := datastore.NewKey(c, "token", token, 0, nil)
	tok := &AuthToken{token}
	_, err = datastore.Put(c, k, tok)
	if err != nil {
		log.Errorf(c, "PutToken datastore put: %v \n", err)
		return err
	}

	return nil

}

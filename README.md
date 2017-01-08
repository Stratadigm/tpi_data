##  TPI Data ##
Data entities and data access for TPI and GAE datastore

## Contents     
- [Data](#data)
- [Validation](#validation)
    * [Validator](#validator)
    * [Marshaler](#marshaler)
- [Datastore Access](#dao)
    * [Create](#create)
    * [Retrieve](#retrieve)
    * [Update](#update)
    * [Delete](#delete)
- [Gotchas](#gotchas)
- [References](#references)  


## Data ##
In v1 there's three data structures of interest:

+ User
    + Name string
    + Email string
    + Confirmed bool
    + Thalis []Thali // thalis contributed
    + Venues []int64 // venues contributed - []int64 due to datastore restriction of no nested slices
    + Rep int
    + Submitted time.Time

+ Venue
    + Name string
    + Latitude float64 // can be replaced with Location appengine.GeoPoint
    + Longitude float64 // can be replaced with Location appengine.GeoPoint
    + Thalis []int64
    + Submitted time.Time

+ Thali
    + Name string
    + Target int // 1-4 target customer profile
    + Limited bool
    + Region int // 1-3 target cuisine
    + Price float64 //
    + Photo string // filename in GCS
    + Venueid int64  // available at venue with id
    + Userid int64 // contributing by user with id
    + Verified bool
    + Accepted bool
    + Submitted time.Time

User -> Thali = One-to-many

We need a appengine datastore access structure and also a Postgres and/or Mongo access structure for deployment in case of move away from Appengine. All in Go.

In the appengine datastore version, Thali is slightly modified to include Id of Venue rather than a Venue (see appengine datastore reference). 

## Validation ##

All entities should implement Validator interface to define

User also implements the Marshaler interface

### Validator ###

Entities should implement their own validation by implementing Validator (defining Validate) so all data definition is encapsulated. Validate takes as argument an optional argument which can be used to differentiate between Create and Update requests. 

    type Validator interface {

        func Validate(v ...interface{}) error

    }

### Marshaler ###

Entities that want to customize json encoding/decoding should implement Marshaler interface by defining MarshalJSON function. User defines MarshalJSON to hide / blank out the password field before encoding JSON to the response writer. 

        func (u *User) MarshalJSON() ([]byte, error) {

                type Alias User

                if !appengine.IsDevAppServer() {
        
                        u.Password = ""

                }

                return json.Marshal(&struct {

                        *Alias

                }{

                        Alias: (*Alias)(u),

                },

                )

        }

Alias is used because it has the same fields as User but not the methods. The following would cause a stack overflow:

        func (u *User) MarshalJSON() ([]byte, error) {
 
                u.Password = ""

                return json.Marshal(u)

        }

## DAO ##

Datastore access struct is created using Context. Receiver methods use interface{} as argument and reflection to CRUD entities in GAE datastore 

### Create

### Retrieve

### Update

### Delete 


## References ##
+ [Appengine datastore api](https://godoc.org/google.golang.org/appengine/datastore)
+ [GCP Appengine Console](https://console.cloud.google.com/appengine?project=tpi)
+ [Method: apps.repair](https://cloud.google.com/appengine/docs/admin-api/reference/rest/v1/apps/repair)
+ [Google Cloud Platform Datastore Reference](https://cloud.google.com/appengine/docs/go/datastore/reference)



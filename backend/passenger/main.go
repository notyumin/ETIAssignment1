package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"reflect"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

//Note: Field names have to be capitalised to be public to work with Gorm
//Passenger with ID
type Passenger struct {
	Id        int `gorm:"primaryKey"`
	FirstName string
	LastName  string
	MobileNo  int
	Email     string
}

//Global Variables
var db *gorm.DB

func main() {
	loadEnv()
	initDb()
	migrateDb()
	initRouter()
}

func loadEnv() {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func initDb() {
	dsn := os.Getenv("DSN")

	var err error

	//set global var "db"
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to database with DSN: " + dsn)
	}
}

func migrateDb() {
	err := db.AutoMigrate(&Passenger{})
	if err != nil {
		panic("DB Migration failed with error: " + err.Error())
	}
}

func initRouter() {
	router := mux.NewRouter()

	router.HandleFunc("/passengers", getPassengers).Methods("GET")
	router.HandleFunc("/passengers/{id}", getPassengerById).Methods("GET")
	router.HandleFunc("/passengers", createPassenger).Methods("POST")
	router.HandleFunc("/passengers/{id}", updatePassenger).Methods("PUT")
	router.HandleFunc("/passengers/{id}", deletePassenger).Methods("DELETE")

	portNo := os.Getenv("PASSENGER_PORT")

	fmt.Printf("Passenger Microservice running on port %s...\n", portNo)
	err := http.ListenAndServe(":"+portNo, router)
	if err != nil {
		panic("InitRouter failed with error: " + err.Error())
	}
}

/////////////////////////
//                     //
//    HTTP Functions   //
//                     //
/////////////////////////

func getPassengers(w http.ResponseWriter, r *http.Request) {
	var passengers []Passenger

	urlParams := r.URL.Query()
	if queryEmail, ok := urlParams["email"]; ok {
		email := queryEmail[0]
		db.Where("email = ?", email).Find(&passengers)
	} else {
		db.Find(&passengers)
	}

	httpRespondWith(w, http.StatusOK, passengers)
}

func getPassengerById(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]

	var passenger Passenger
	err := db.Where("id = ?", id).First(&passenger).Error
	if err != nil {
		httpRespondWith(w, http.StatusNotFound, "User doesn't exist")
		return
	}

	httpRespondWith(w, http.StatusOK, passenger)
}

func createPassenger(w http.ResponseWriter, r *http.Request) {
	var passenger Passenger

	decodeErr := json.NewDecoder(r.Body).Decode(&passenger)
	if decodeErr != nil {
		httpRespondWith(w, http.StatusBadRequest, "Invalid JSON: "+decodeErr.Error())
		return
	}

	//validate empty fields
	if isFieldMissing(w, passenger.FirstName, "FirstName") ||
		isFieldMissing(w, passenger.LastName, "LastName") ||
		isFieldMissing(w, passenger.MobileNo, "MobileNo") ||
		isFieldMissing(w, passenger.Email, "Email") {
		return
	}

	//validate email exist
	emailExist := existInDb("email", passenger.Email)
	if emailExist {
		httpRespondWith(w, http.StatusConflict, "Email already in-use.")
		return
	}

	//Disallow manual setting of Id
	passenger.Id = 0

	dbErr := db.Create(&passenger).Error
	if dbErr != nil {
		httpRespondWith(w, http.StatusBadRequest, "Invalid Data")
		return
	}

	httpRespondWith(w, http.StatusCreated, passenger)
}

func updatePassenger(w http.ResponseWriter, r *http.Request) {
	var passenger Passenger

	decodeErr := json.NewDecoder(r.Body).Decode(&passenger)
	if decodeErr != nil {
		httpRespondWith(w, http.StatusBadRequest, "Invalid JSON: "+decodeErr.Error())
		return
	}

	//Disallow manual setting of Id
	passenger.Id = 0

	params := mux.Vars(r)
	id := params["id"]

	dbErr := db.Model(&Passenger{}).Where("id = ?", id).Updates(passenger).Error
	if dbErr != nil {
		httpRespondWith(w, http.StatusBadRequest, "Invalid Data")
		return
	}

	var newPassenger Passenger
	db.Where("id = ?", id).First(&newPassenger)

	httpRespondWith(w, http.StatusAccepted, newPassenger)
}

func deletePassenger(w http.ResponseWriter, r *http.Request) {
	//check admin password
	if !hasValidAdminPass(r) {
		httpRespondWith(w, http.StatusForbidden, "Unauthorized User")
		return
	}

	params := mux.Vars(r)
	id := params["id"]

	//check user exist
	idExist := existInDb("id", id)
	if !idExist {
		httpRespondWith(w, http.StatusNotFound, "User doesn't exist")
		return
	}

	//delete from passenger where id = id
	db.Delete(&Passenger{}, id)

	httpRespondWith(w, http.StatusAccepted, fmt.Sprintf("User of ID %s successfully deleted", id))
}

/////////////////////////
//                     //
//       Helpers       //
//                     //
/////////////////////////

func httpRespondWith(w http.ResponseWriter, statusCode int, data interface{}) {
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

/*
This function checks whether a data has a zero value.
If it does, it will return true and write a http response
*/
func isFieldMissing(w http.ResponseWriter, data interface{}, fieldName string) bool {
	if isZero(data) {
		errorMsg := fmt.Sprintf("%s field is missing/incorrect.", fieldName)
		httpRespondWith(w, http.StatusBadRequest, errorMsg)
		return true
	}
	return false
}

func isZero(data interface{}) bool {
	value := reflect.ValueOf(data)
	return value.IsZero()
}

/*
This function returns whether a row with the given value for the given fieldName exists
*/
func existInDb(fieldName string, value interface{}) bool {
	var dbPassenger Passenger

	//if email doens't exist, db.First returns a ErrRecordNotFound
	err := db.Where(fieldName+" = ?", value).First(&dbPassenger).Error
	return err != gorm.ErrRecordNotFound
}

func hasValidAdminPass(r *http.Request) bool {
	urlParams := r.URL.Query()
	if queryPassword, ok := urlParams["adminPassword"]; ok {
		password := os.Getenv("ADMIN_PASSWORD")
		if queryPassword[0] == password {
			return true
		}
	}
	return false
}

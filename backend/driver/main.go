package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"reflect"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Driver struct {
	Id           int `gorm:"primaryKey"`
	FirstName    string
	LastName     string
	MobileNo     int
	Email        string
	CarLicenseNo string
	Available    bool
}

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
	err := db.AutoMigrate(&Driver{})
	if err != nil {
		panic("DB Migration failed with error: " + err.Error())
	}
}

func initRouter() {
	router := mux.NewRouter()

	router.HandleFunc("/drivers", getDrivers).Methods("GET")
	router.HandleFunc("/drivers/{id}", getDriverById).Methods("GET")
	router.HandleFunc("/drivers", createDriver).Methods("POST")
	router.HandleFunc("/drivers/{id}", updateDriver).Methods("PUT")
	router.HandleFunc("/drivers/{id}", deleteDriver).Methods("DELETE")

	portNo := os.Getenv("DRIVER_PORT")

	fmt.Printf("Driver Microservice running on port %s...\n", portNo)
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

func getDrivers(w http.ResponseWriter, r *http.Request) {
	var drivers []Driver

	urlParams := r.URL.Query()
	if queryAvailable, ok := urlParams["available"]; ok {
		stringAvailable := queryAvailable[0]
		available, _ := strconv.ParseBool(stringAvailable)
		db.Where("available = ?", available).Find(&drivers)
	} else if queryEmail, ok := urlParams["email"]; ok {
		email := queryEmail[0]
		db.Where("email = ?", email).Find(&drivers)
	} else {
		db.Find(&drivers)
	}

	httpRespondWith(w, http.StatusOK, drivers)
}

func getDriverById(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]

	var driver Driver
	err := db.Where("id = ?", id).First(&driver).Error
	if err != nil {
		httpRespondWith(w, http.StatusNotFound, "User doesn't exist")
		return
	}

	httpRespondWith(w, http.StatusOK, driver)
}

func createDriver(w http.ResponseWriter, r *http.Request) {
	var driver Driver

	decodeErr := json.NewDecoder(r.Body).Decode(&driver)
	if decodeErr != nil {
		httpRespondWith(w, http.StatusBadRequest, "Invalid JSON: "+decodeErr.Error())
		return
	}

	//validate empty fields
	if isFieldMissing(w, driver.FirstName, "FirstName") ||
		isFieldMissing(w, driver.LastName, "LastName") ||
		isFieldMissing(w, driver.MobileNo, "MobileNo") ||
		isFieldMissing(w, driver.Email, "Email") ||
		isFieldMissing(w, driver.CarLicenseNo, "CarLicenseNo") {
		return
	}

	//validate email exist
	emailExist := existInDb("email", driver.Email)
	if emailExist {
		httpRespondWith(w, http.StatusConflict, "Email already in-use.")
		return
	}

	//Disallow manual setting of Id
	driver.Id = 0

	//Drivers should be available on creation
	driver.Available = true

	dbErr := db.Create(&driver).Error
	if dbErr != nil {
		httpRespondWith(w, http.StatusBadRequest, "Invalid Data")
		return
	}

	httpRespondWith(w, http.StatusCreated, driver)
}

func updateDriver(w http.ResponseWriter, r *http.Request) {
	var driver Driver

	decodeErr := json.NewDecoder(r.Body).Decode(&driver)
	if decodeErr != nil {
		httpRespondWith(w, http.StatusBadRequest, "Invalid JSON: "+decodeErr.Error())
		return
	}

	params := mux.Vars(r)
	id := params["id"]

	//use map syntax for gorm so that it can update zero values
	dbErr := db.Model(&Driver{}).Where("id = ?", id).Updates(map[string]interface{}{
		"FirstName":    driver.FirstName,
		"LastName":     driver.LastName,
		"MobileNo":     driver.MobileNo,
		"Email":        driver.Email,
		"CarLicenseNo": driver.CarLicenseNo,
		"Available":    driver.Available,
	})
	if dbErr != nil {
		httpRespondWith(w, http.StatusBadRequest, "Invalid Data")
		return
	}

	var newDriver Driver
	db.Where("id = ?", id).First(&newDriver)

	httpRespondWith(w, http.StatusAccepted, newDriver)
}

func deleteDriver(w http.ResponseWriter, r *http.Request) {
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

	//delete from driver where id = id
	db.Delete(&Driver{}, id)

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
	var dbDriver Driver

	//if email doens't exist, db.First returns a ErrRecordNotFound
	err := db.Where(fieldName+" = ?", value).First(&dbDriver).Error
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

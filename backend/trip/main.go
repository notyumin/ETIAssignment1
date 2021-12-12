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

type Trip struct {
	Id            int
	PassengerId   int
	DriverId      int
	PickUpPostal  int
	DropOffPostal int
	Status        string //"waiting", "driving" or "finished"
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
	err := db.AutoMigrate(&Trip{})
	if err != nil {
		panic("DB Migration failed with error: " + err.Error())
	}
}

func initRouter() {
	router := mux.NewRouter()

	router.HandleFunc("/trips", getTrips).Methods("GET")
	router.HandleFunc("/trips/{id}", getTripById).Methods("GET")
	router.HandleFunc("/trips", createTrip).Methods("POST")
	router.HandleFunc("/trips/{id}", updateTrip).Methods("PUT")
	router.HandleFunc("/trips/{id}", deleteTrip).Methods("DELETE")

	portNo := os.Getenv("TRIP_PORT")

	fmt.Printf("Trip Microservice running on port %s...\n", portNo)
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

func getTrips(w http.ResponseWriter, r *http.Request) {
	var trips []Trip

	urlParams := r.URL.Query()
	if queryPassengerId, ok := urlParams["passengerId"]; ok {
		stringPassengerId := queryPassengerId[0]
		passengerId, _ := strconv.Atoi(stringPassengerId)
		db.Where("passenger_id = ?", passengerId).Find(&trips)
	} else if queryDriverId, ok := urlParams["driverId"]; ok {
		stringDriverId := queryDriverId[0]
		driverId, _ := strconv.Atoi(stringDriverId)
		db.Where("driver_id = ?", driverId).Find(&trips)
	} else {
		db.Find(&trips)
	}

	httpRespondWith(w, http.StatusOK, trips)
}

func getTripById(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]

	var trip Trip
	err := db.Where("id = ?", id).First(&trip).Error
	if err != nil {
		httpRespondWith(w, http.StatusNotFound, "Trip doesn't exist")
		return
	}

	httpRespondWith(w, http.StatusOK, trip)
}

func createTrip(w http.ResponseWriter, r *http.Request) {
	var trip Trip

	decodeErr := json.NewDecoder(r.Body).Decode(&trip)
	if decodeErr != nil {
		httpRespondWith(w, http.StatusBadRequest, "Invalid JSON: "+decodeErr.Error())
		return
	}

	if isFieldMissing(w, trip.PassengerId, "PassengerId") ||
		isFieldMissing(w, trip.DriverId, "DriverId") ||
		isFieldMissing(w, trip.PickUpPostal, "PickUpPostal") ||
		isFieldMissing(w, trip.DropOffPostal, "PickUpPostal") {
		return
	}

	//Disallow manual setting of Id
	trip.Id = 0

	//initialise trips as "waiting"
	trip.Status = "waiting"

	dbErr := db.Create(&trip).Error
	if dbErr != nil {
		httpRespondWith(w, http.StatusBadRequest, "Invalid Data")
		return
	}

	httpRespondWith(w, http.StatusCreated, trip)
}

func updateTrip(w http.ResponseWriter, r *http.Request) {
	var trip Trip

	decodeErr := json.NewDecoder(r.Body).Decode(&trip)
	if decodeErr != nil {
		httpRespondWith(w, http.StatusBadRequest, "Invalid JSON: "+decodeErr.Error())
		return
	}

	//Disallow manual setting of Id
	trip.Id = 0

	params := mux.Vars(r)
	id := params["id"]

	dbErr := db.Model(&Trip{}).Where("id = ?", id).Updates(trip).Error
	if dbErr != nil {
		httpRespondWith(w, http.StatusBadRequest, "Invalid Data")
		return
	}

	var newTrip Trip
	db.Where("id = ?", id).First(&newTrip)

	httpRespondWith(w, http.StatusAccepted, newTrip)
}

func deleteTrip(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]

	//check trip exist
	idExist := existInDb("id", id)
	if !idExist {
		httpRespondWith(w, http.StatusNotFound, "Trip doesn't exist")
		return
	}

	//delete from passenger where id = id
	db.Delete(&Trip{}, id)

	httpRespondWith(w, http.StatusAccepted, fmt.Sprintf("Trip of ID %s successfully deleted", id))
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

func existInDb(fieldName string, value interface{}) bool {
	var dbTrip Trip

	//if email doens't exist, db.First returns a ErrRecordNotFound
	err := db.Where(fieldName+" = ?", value).First(&dbTrip).Error
	return err != gorm.ErrRecordNotFound
}

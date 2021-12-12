package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
)

type Passenger struct {
	Id        int
	FirstName string
	LastName  string
	MobileNo  int
	Email     string
}

type Driver struct {
	Id           int
	FirstName    string
	LastName     string
	MobileNo     int
	Email        string
	CarLicenseNo string
	Available    bool
}

type Trip struct {
	Id            int
	PassengerId   int
	DriverId      int
	PickUpPostal  int
	DropOffPostal int
	Status        string
}

var passengerUrl string = "http://localhost:5000/passengers"
var driverUrl string = "http://localhost:5001/drivers"
var tripUrl string = "http://localhost:5002/trips"

var scanner *bufio.Scanner

func main() {
	scanner = bufio.NewScanner(os.Stdin)

	mainMenu()
}

func mainMenu() {
menu:
	for {
		fmt.Println("Welcome to HytchHyke!")
		fmt.Println("[1] Login as Passenger")
		fmt.Println("[2] Login as Driver")
		fmt.Println("[3] Register Passenger Account")
		fmt.Println("[4] Register Driver Account")
		fmt.Println("[0] Quit")

		userOption := getStrInput()
		switch userOption {
		case "1":
			loginPassenger()
		case "2":
			loginDriver()
		case "3":
			registerPassenger()
		case "4":
			registerDriver()
		case "0":
			break menu
		default:
			fmt.Println("\nInvalid Option")
		}
	}
}

/////////////////////////
//                     //
//      Passenger      //
//                     //
/////////////////////////

func registerPassenger() {
	fmt.Print("First Name: ")
	firstName := getStrInput()

	fmt.Print("Last Name: ")
	lastName := getStrInput()

	fmt.Print("Mobile Number: ")
	mobileNo := getIntInput()

	fmt.Print("Email: ")
	email := getStrInput()

	newPassenger := Passenger{
		FirstName: firstName,
		LastName:  lastName,
		MobileNo:  mobileNo,
		Email:     email,
	}

	err := createPassenger(newPassenger)
	if err != nil {
		fmt.Println("Error: ", err.Error())
	} else {
		fmt.Println("Passenger Created Successfully!")
	}
}

func loginPassenger() {
login:
	for {
		fmt.Print("Please enter your email: ")
		email := getStrInput()
		passenger := getPassengerByEmail(email)
		if (passenger == Passenger{} || email != passenger.Email) {
			fmt.Println("\nInvalid email")
			break login
		} else {
			fmt.Printf("\nWelcome %s %s!\n", passenger.FirstName, passenger.LastName)
			passengerMenu(passenger)
			//after logout
			break login
		}
	}
}

func passengerMenu(passenger Passenger) {
menu:
	for {
		fmt.Println("[1] Book Trip")
		fmt.Println("[2] View Trips")
		fmt.Println("[3] Update Details")
		fmt.Println("[0] Logout")

		userOption := getStrInput()
		switch userOption {
		case "1":
			bookTrip(passenger)
		case "2":
			displayPassengerTrip(passenger)
		case "3":
			updatePassengerDetails(passenger)
			break menu
		case "0":
			break menu
		}
	}
}

func bookTrip(passenger Passenger) {
	fmt.Print("Pick Up Postal Code: ")
	pickUpPostal := getIntInput()
	fmt.Print("\nDrop Off Postal Code: ")
	dropOffPostal := getIntInput()

	//get available driver
	driver := getAvailableDriver()
	if (driver == Driver{}) {
		fmt.Println("\nNo available drivers")
		return
	}

	err := createTrip(pickUpPostal, dropOffPostal, driver.Id, passenger.Id)
	if err != nil {
		fmt.Println("Trip could not be booked: ", err.Error())
	} else {
		fmt.Println("Trip booked successfully!")
	}

	//set driver to unavailable
	driver.Available = false
	updateDriver(driver)
}

func displayPassengerTrip(passenger Passenger) {
	trips := getPassengerTrips(passenger.Id)

	//display in reverse chronological order
	for i := len(trips) - 1; i >= 0; i-- {
		trip := trips[i]
		fmt.Println("\nTrip ID: ", trip.Id)
		fmt.Println("Pick Up Postal Code: ", trip.PickUpPostal)
		fmt.Println("Drop Off Postal Code: ", trip.DropOffPostal)
		fmt.Println("Driver ID: ", trip.DriverId)
		fmt.Println("Passenger ID: ", trip.PassengerId)
		fmt.Println("Trip Status", trip.Status)
		fmt.Println()
	}
}

func updatePassengerDetails(passenger Passenger) {
	fmt.Print("New First Name: ")
	firstName := getStrInput()

	fmt.Print("New Last Name: ")
	lastName := getStrInput()

	fmt.Print("New Mobile Number: ")
	mobileNo := getIntInput()

	fmt.Print("New Email: ")
	email := getStrInput()

	newPassenger := Passenger{
		Id:        passenger.Id,
		FirstName: firstName,
		LastName:  lastName,
		MobileNo:  mobileNo,
		Email:     email,
	}
	err := updatePassenger(newPassenger)
	if err != nil {
		fmt.Println("Error: ", err.Error())
	} else {
		fmt.Println("Passenger Updated Successfully!")
		fmt.Println("Please log in again")
	}
}

/////////////////////////
//                     //
//       Driver        //
//                     //
/////////////////////////

func registerDriver() {
	fmt.Print("First Name: ")
	firstName := getStrInput()

	fmt.Print("Last Name: ")
	lastName := getStrInput()

	fmt.Print("Mobile Number: ")
	mobileNo := getIntInput()

	fmt.Print("Email: ")
	email := getStrInput()

	fmt.Print("Car Licence Plate Number: ")
	carLicenseNo := getStrInput()

	newDriver := Driver{
		FirstName:    firstName,
		LastName:     lastName,
		MobileNo:     mobileNo,
		Email:        email,
		CarLicenseNo: carLicenseNo,
	}

	err := createDriver(newDriver)
	if err != nil {
		fmt.Println("Error: ", err.Error())
	} else {
		fmt.Println("Driver Registered Successfully!")
	}
}

func loginDriver() {
login:
	for {
		fmt.Print("Please enter your email: ")
		email := getStrInput()
		driver := getDriverByEmail(email)
		if (driver == Driver{} || email != driver.Email) {
			fmt.Println("\nInvalid email")
			break login
		} else {
			fmt.Printf("\nWelcome %s %s!\n", driver.FirstName, driver.LastName)
			driverMenu(driver)
			//after logout
			break login
		}
	}
}

func driverMenu(driver Driver) {
menu:
	for {
		fmt.Println("[1] Start Trip")
		fmt.Println("[2] End Trip")
		fmt.Println("[3] Update Details")
		fmt.Println("[0] Logout")

		userOption := getStrInput()
		switch userOption {
		case "1":
			startTrip(driver)
		case "2":
			endTrip(driver)
		case "3":
			updateDriverDetails(driver)
			break menu
		case "0":
			break menu
		}
	}
}

func updateDriverDetails(driver Driver) {
	fmt.Print("New First Name: ")
	firstName := getStrInput()

	fmt.Print("New Last Name: ")
	lastName := getStrInput()

	fmt.Print("New Mobile Number: ")
	mobileNo := getIntInput()

	fmt.Print("New Email: ")
	email := getStrInput()

	fmt.Print("New Car Licence Plate Number: ")
	carLicenseNo := getStrInput()

	newDriver := Driver{
		Id:           driver.Id,
		FirstName:    firstName,
		LastName:     lastName,
		MobileNo:     mobileNo,
		Email:        email,
		CarLicenseNo: carLicenseNo,
	}

	err := updateDriver(newDriver)
	if err != nil {
		fmt.Println("Error: ", err.Error())
	} else {
		fmt.Println("Driver Updated Successfully!")
		fmt.Println("Please log in again")
	}
}

func startTrip(driver Driver) {
	//get all trips
	//update trip where driver_id = driver.id && status == "waiting"
	//into status = "driving"

	driverTrips := getDriverTrips(driver.Id)
	var waitingTrip Trip
	for _, trip := range driverTrips {
		if trip.Status == "waiting" {
			waitingTrip = trip
		}
	}
	if (waitingTrip == Trip{}) {
		fmt.Println("No waiting trips")
	} else {
		waitingTrip.Status = "driving"
		updateTrip(waitingTrip)
		fmt.Println("\nTrip started")
	}
}

func endTrip(driver Driver) {
	//update db where driver_id = driver.id && status == "driving"
	//into status = "finish"
	//update driver available to true
	driverTrips := getDriverTrips(driver.Id)
	var drivingTrip Trip
	for _, trip := range driverTrips {
		if trip.Status == "driving" {
			drivingTrip = trip
		}
	}
	if (drivingTrip == Trip{}) {
		fmt.Println("No driving trips")
	} else {
		drivingTrip.Status = "finished"
		updateTrip(drivingTrip)
		fmt.Println("\nTrip ended")
	}

	//set driver to available
	driver.Available = true
	updateDriver(driver)
}

/////////////////////////
//                     //
//    API Functions    //
//                     //
/////////////////////////

func getPassengerByEmail(email string) Passenger {
	var passenger Passenger

	url := fmt.Sprintf("%s?email=%s", passengerUrl, email)

	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err.Error())
		return passenger
	}

	//email not found
	if resp.StatusCode != http.StatusOK {
		return passenger
	}

	var listPassenger []Passenger
	json.NewDecoder(resp.Body).Decode(&listPassenger)
	if len(listPassenger) <= 0 {
		return passenger
	}
	passenger = listPassenger[0]
	return passenger
}

func getAvailableDriver() Driver {
	var driver Driver

	url := fmt.Sprintf("%s?available=%t", driverUrl, true)

	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err.Error())
		return driver
	}

	//email not found
	if resp.StatusCode != http.StatusOK {
		return driver
	}

	var listDriver []Driver
	json.NewDecoder(resp.Body).Decode(&listDriver)
	if len(listDriver) <= 0 {
		return driver
	}
	driver = listDriver[0]
	return driver
}

func updateDriver(newDriver Driver) error {
	url := fmt.Sprintf("%s/%d", driverUrl, newDriver.Id)

	_, err := httpPut(url, newDriver)
	return err
}

func getDriverByEmail(email string) Driver {
	var driver Driver

	url := fmt.Sprintf("%s?email=%s", driverUrl, email)

	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err.Error())
		return driver
	}

	//email not found
	if resp.StatusCode != http.StatusOK {
		return driver
	}

	var listDriver []Driver
	json.NewDecoder(resp.Body).Decode(&listDriver)
	if len(listDriver) <= 0 {
		return driver
	}
	driver = listDriver[0]
	return driver
}

func createTrip(pickUpPostal int, dropOffPostal int, driverId int, passengerId int) error {
	url := tripUrl

	var newTrip Trip = Trip{
		PickUpPostal:  pickUpPostal,
		DropOffPostal: dropOffPostal,
		DriverId:      driverId,
		PassengerId:   passengerId,
	}

	_, err := httpPost(url, newTrip)
	return err
}

func getPassengerTrips(id int) []Trip {
	var trips []Trip

	url := fmt.Sprintf("%s?passengerId=%d", tripUrl, id)

	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error: ", err.Error())
		return trips
	}

	json.NewDecoder(resp.Body).Decode(&trips)
	return trips
}

func updatePassenger(newPassenger Passenger) error {
	url := fmt.Sprintf("%s/%d", passengerUrl, newPassenger.Id)

	_, err := httpPut(url, newPassenger)
	return err
}

func createPassenger(newPassenger Passenger) error {
	url := passengerUrl
	_, err := httpPost(url, newPassenger)
	return err
}

func createDriver(newDriver Driver) error {
	url := driverUrl

	_, err := httpPost(url, newDriver)
	return err
}

func getDriverTrips(id int) []Trip {
	var trips []Trip

	url := fmt.Sprintf("%s?driverId=%d", tripUrl, id)

	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error: ", err.Error())
		return trips
	}

	json.NewDecoder(resp.Body).Decode(&trips)
	return trips
}

func updateTrip(newTrip Trip) error {
	url := fmt.Sprintf("%s/%d", tripUrl, newTrip.Id)

	_, err := httpPut(url, newTrip)
	return err
}

/////////////////////////
//                     //
//       Helpers       //
//                     //
/////////////////////////

func getStrInput() string {
	scanner.Scan()
	userInput := scanner.Text()
	return userInput
}

func getIntInput() int {
	scanner.Scan()
	stringInput := scanner.Text()
	intInput, _ := strconv.Atoi(stringInput)
	return intInput
}

func httpPost(url string, data interface{}) (*http.Response, error) {
	jsonData, _ := json.Marshal(data)

	response, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	return response, err
}

func httpPut(url string, data interface{}) (*http.Response, error) {
	jsonData, _ := json.Marshal(data)

	request, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(jsonData))
	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)

	return response, err
}

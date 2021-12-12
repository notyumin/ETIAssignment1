# HytchHyke Backend Setup

## 1. Create Database
Connect to your MySQL instance and run:
```
CREATE database hytchhyke;
```
> Note: You may replace `hytchhyke` with another name if you wish

## 2. Set Environment Variables
Use any text editor of your choice to add the Data Source Name (DSN) of the database you created above to the `.env` file, like so:
```
DSN=user:password@tcp(127.0.0.1:3306)/hytchhyke
```
> Note: Replace `user`, `password` and `hytchhyke` with your Database username, password and database name respectively


## 3. Run Microservices
First, cd into the `backend` folder using:
```
cd backend
```

Then, run the start script using:
```
bash start.sh
```

>Note: Microservices are all started using 1 start script for convenience sake. It is very simple to start them separately if needed by running `go run` on each program manually.

> The `start.sh` script will install the dependencies in all 3 microservices using `go mod tidy` and run them with `go run`. 
> 
> Database Migrations are handled with Gorm's AutoMigrate feature, so the database tables will be created upon running go code.

You should see the following if the microservices are successfully running:
```
Starting Microservices...
Passenger Microservice Running...
Driver Microservice Running...
Trip Microservice Running...
```

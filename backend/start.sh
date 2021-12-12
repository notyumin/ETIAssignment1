echo "Starting Microservices..."

run_passenger() {
    cd passenger
    go mod tidy
    go run main.go
}

run_driver() {
    cd driver
    go mod tidy
    go run main.go
}

run_trip() {
    cd trip
    go mod tidy
    go run main.go
}

run_passenger & 
run_driver & 
run_trip

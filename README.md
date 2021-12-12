_This project was made for Assignment 1 of Ngee Ann Polytechnic's Emerging Trends in IT module._ 

# HytchHyke

## Microservice Architecture Design

![Architecture Diagram](./architecture_diagram.png)

## Microservice Design Considerations

HytchHyke's backend is split into 3 different microservices: `passenger`, `driver` and `trip`. 

The first consideration was that microservices are supposed to only have a single responsibility. Thus, in HytchHyke, each microservice is responsible for Creating, Reading, Updating and Deleting (CRUD) their respective data type from their databases. 

> For example, the `passenger` microservice would only be responsible for CRUD for passengers. 

The second consideration was that microservices have to be loosely coupled. In fact, all of HytchHyke's microservices are completely decoupled from each other, meaning there is 0 dependency between the microservices. This is because there is minimal logic handled by the microservices; The microservices only handle simple CRUD functionality. The rest of the logic is handled in the front end, which allows the front end developers to have more control over the functionality they would like to achieve.

> This also allows for maximum maintainability, as only the frontend would need to be updated if any of the microservices are changed! 

The third consideration was that the microservices should be easily testable, meaning that side effects within the code should be minimized, because they make testing extremely difficult. Thus, none of HytchHyke's microservice functions contain any unknown side effects. Each function performs what it is clearly stated to do, and nothing more.

## Backend Set Up
Please refer to the [backend/README.md](https://github.com/notyumin/ETIAssignment1/blob/main/backend/README.md) file.

## Frontend Set up
> Note: Make sure backend is running!

cd into `console` using
```
cd console
```

Run the console by running:
```
go run main.go
```

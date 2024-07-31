# Price Alert Application

This application allows users to set price alerts for cryptocurrencies and receive notifications when the target price is reached.

## Running the project

1. Clone the repository
2. If you're on Windows, you need to install and run Docker Desktop
3. Run `docker-compose up --build`
4. The application will be available at `http://localhost:8080`

## API Endpoints

### User Registration
URL: `/users/signup`
Method: `POST`
Request Body:
```json
{
 "email": "akarsh1@gmail.com",
 "password": "pass1"
}
```

### User Login
URL: `/users/login`
Method: `POST`
Request Body:
```json
{
 "email": "akarsh1@gmail.com",
 "password": "pass1"
}
```
Response: JWT token

### Alert Creation
URL: `/alerts/create`
Method: `POST`
Authorization: JWT token as bearer token
Request Body:
```json
{
 "coin": "BTC",
 "target_price": 35000.0
}
```

### Alert Deletion
URL: `/alerts/delete/:id`
Method: `DELETE`
Authorization: JWT token as bearer token

### Get Alerts
URL: `/alerts`
Status filter query in URL(optional)
Method: `GET`
Authorization: JWT token as bearer token, 

## Alert Notification System

The application uses Binance's WebSocket connection to get real-time price updates. We establish a connection and continuously read the messages regarding price updates from the connection using an infinite loop. If there is any untriggered Alert for a specific coin, then the Alert's target price is compared with the coin's current price. When a target price is reached for an untriggered alert, the alert status is updated in the database, and a notification is printed to the console (in a production environment, this would send an email to the user).

## Caching

The "fetch all alerts" endpoint uses Redis to cache results. The cache memory is updated every time a new Alert is created, deleted or triggered

## Demo and contact

[Demo video](https://drive.google.com/file/d/1FfJiLLm3OcnygwNZWEmbucTK80XnyVv0/view?usp=sharing)
Discord id: 763377449789685760


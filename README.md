# Rest API With JWT for Kafka

[![Go](https://github.com/cploutarchou/go-kafka-rest/actions/workflows/go.yml/badge.svg)](https://github.com/cploutarchou/go-kafka-rest/actions/workflows/go.yml)

This project demonstrates building a REST API with JWT (JSON Web Token) authentication for Kafka using Golang. The API is implemented using the Golang programming language and leverages the power of Kafka for message processing and communication.

## Features

- User registration and authentication with JWT
- Publish messages to Kafka topics
- Consume messages from Kafka topics
- Secure and protect API routes using JWT authentication

## Setup

To set up the project, follow these steps:

1. Clone the repository: `git clone https://github.com/cploutarchou/go-kafka-rest.git`
2. Install the required dependencies: `go mod download`
3. Set up the necessary environment variables (e.g., Kafka broker address, JWT secret, etc.)
4. Start the application: `go run main.go`

## Usage

Once the application is running, you can perform the following actions:

- Register a new user: `POST /api/register`
- Authenticate and obtain a JWT token: `POST /api/login`
- Publish a message to a Kafka topic: `POST /api/publish`
- Consume messages from a Kafka topic: `GET /api/consume`

Make sure to include the required authentication headers (JWT token) for the protected routes.

## Contributing

Contributions are welcome! If you find any issues or have suggestions for improvement, feel free to open an issue or submit a pull request.

## License

This project is licensed under the [MIT License](LICENSE).

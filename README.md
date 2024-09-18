# Authentication Service API

This project is a **Golang-based Authentication Service** API designed to provide user authentication and authorization functionality. It offers a set of RESTful endpoints to handle user login, registration, token management, password reset, and more. It also includes support for user profile management and image uploads.

<img src="./template/docs/assets/img/logo.jpg" alt="Project Logo" width="300" height="300">

## Features

- User registration and login
- JWT-based token authentication and refresh
- Password reset via email
- User profile management
- Protected routes for logged-in users
- Configurable via `settings.yml`
- CORS and middleware for secure and robust routing
- Uses PostgreSQL as the database

#### Project tree structure
    
    .
    ├── cmd
    │   └── api
    │       ├── application
    │       │   ├── application.go
    │       │   ├── config.go
    │       │   └── middleware.go
    │       ├── auth
    │       │   └── Auth.go
    │       ├── controllers
    │       │   ├── mailgun.go
    │       │   ├── userController.go
    │       │   └── userController_test.go
    │       ├── main.go
    │       ├── server
    │       │   ├── handlers
    │       │   │   ├── httpHandlersAuth.go
    │       │   │   ├── httpHandlersImage.go
    │       │   │   └── httpHandlersUser.go
    │       │   └── routes.go
    │       └── utils
    │           └── utils.go
    ├── internal
    │   ├── models
    │   │   ├── Image.go
    │   │   └── Users.go
    │   └── repositories
    │       ├── Image.go
    │       ├── ORMRepo.go
    │       └── Users.go
    ├── serveDocs.go
    ├── Dockerfile
    ├── LICENSE
    ├── README.md
    ├── docker-compose.docker.yml
    ├── docker-compose.yml
    ├── settings.yml
    ├── run.sh
    ├── unmount_db.sh
    ├── generate-env.sh
    └── wait-for-it.sh


## Table of Contents

- [Requirements](#requirements)
- [Installation](#installation)
- [Configuration](#configuration)
- [Running the Application](#running-the-application)
- [API Endpoints](#api-endpoints)
  - [Public Routes](#public-routes)
  - [Protected Routes](#protected-routes)
- [Contributing](#contributing)
- [License](#license)

---

## Requirements

- **Golang** (version 1.18 or higher)
- **PostgreSQL** (version 10 or higher)
- **Go Modules** enabled
- **Docker** (optional, for containerized development)

## Installation

1. **Clone the repository**:

    ```bash
    git clone https://github.com/your-username/auth-service.git
    cd auth-service
    ```

2. **Install dependencies**:

    Ensure that Go Modules are enabled. Install required Go packages:

    ```bash
    go mod tidy
    ```

3. **Run the PostgreSQL Server**:

   Running the Postgres server is simple, only run the docker-compose.yml


## Configuration

All configurations are handled via the `settings.yml` file. You need to configure the `settings.yml` file before running the application.

1. **Create a `settings.yml` file**:

    You can copy the provided `settings.yml.example` file and fill in the appropriate values.

    ```bash
    cp settings.yml.example settings.yml
    ```

2. **Configure the `settings.yml` file**:

    Example structure:

    ```yaml
    api:
        name: TriceraPass Auth Service
        version: 1.0.0
        description: "God creates dinosaurs! God destroys dinosaurs! God creates man! Man destroys God! Man creates dinosaurs!"
        allowedOrigins:
            - http://localhost:3000
            - http://localhost:5051
            - https://dr-malcom.com
        rateLimiting:
            requestsPerMinute: 100
            burst: 50

    application:  
        clientName: Jurassic Login Park - Dino Auth Division
        cookieDomain: localhost
        domain: dr-malcom.com

    server:
        port: 1993
        host: 0.0.0.0
        developmentMode: true

    security:
        jwt:
            secret: TRICERATOPLESS-eb5d5e9f-86ac-4766-93e2-d760cbb86e7d
            expirationTime: 3600
            issuer: dr-malcom.com
            audience: dr-malcom.com

    logging:
        level: INFO
        format: json
        output: stdout

    database:
        type: postgres
        host: localhost
        port: 5432
        user: velociRaptor
        password: CleverGirl
        dbname: dino-division
        sslmode: disable
        timezone: UTC
        connect_timeout: 5

    emailServer:
        serverName: mailgun
        apiKey: your_mailgun_api_key
        domain: your_mailgun_domain

    redis:
        host: localhost
        port: 6379
        expirationTime: 86400

    styles:
        headerColor: "#ffff00"
        headerBackground: "#3300ff"
        headerFont: "Ac437_Acer_VGA_8x8"
        bodyFont: "Unifont"
        bodyColor: "#3300ff"
        bodyBackground: "#ccffff"
        headerFontSize: "24px"
    ```

## Running the Application

1. **Run the application**:

    Use the following command to start the service in development:

    ```bash
    ./run.sh
    ```

2. **Access the API**:

    The API will be available at `http://localhost:1993/auth/api/`.

---

## API Endpoints

The following routes are available in the authentication service. All public routes can be accessed without authentication, while the protected routes require a valid JWT token.

### Public Routes

| Method | Endpoint                                      | Description                                   |
|--------|-----------------------------------------------|-----------------------------------------------|
| `GET`  | `/auth/api/`                                  | Home, check if the service is running         |
| `POST` | `/auth/api/login`                             | Authenticate user and get a JWT token         |
| `POST` | `/auth/api/refresh`                           | Refresh JWT token                             |
| `POST` | `/auth/api/register`                          | Register a new user                           |
| `POST` | `/auth/api/confirmation/{user_id}`            | Confirm user registration                     |
| `GET`  | `/auth/api/confirmation/user/{user_id}`       | Get last confirmation by user ID              |
| `GET`  | `/auth/api/user/{user_email}`                 | Get user information by email                 |
| `POST` | `/auth/api/send_password_email`               | Send password reset email                     |
| `POST` | `/auth/api/user/password_reset/{user_id}`     | Change password using user ID                 |
| `GET`  | `/auth/api/user/password_reset/token/{user_id}` | Fetch password reset token by user ID         |
| `POST` | `/auth/api/user/password_reset/token/use/{user_id}` | Mark password reset token as used             |

### Protected Routes

The following routes require the user to be logged in with a valid JWT token:

| Method | Endpoint                                       | Description                                   |
|--------|------------------------------------------------|-----------------------------------------------|
| `POST` | `/auth/api/logged_in/logout`                   | Logout the currently authenticated user       |
| `GET`  | `/auth/api/logged_in/user/{user_email}`        | Get user details by email                     |
| `GET`  | `/auth/api/logged_in/user/{user_id}`           | Get user details by user ID                   |
| `PATCH`| `/auth/api/logged_in/user/{user_id}`           | Update user information                       |
| `GET`  | `/auth/api/logged_in/user/profile/{filename}`  | Serve static user profile image               |
| `POST` | `/auth/api/logged_in/upload/profile`           | Upload a new profile image                    |
| `POST` | `/auth/api/logged_in/user/password_reset/{user_id}` | Reset user password by user ID              |
| `POST` | `/auth/api/logged_in/user/send_password_email/{user_id}` | Send password reset email to user        |

---

## Contributing

We welcome contributions! If you'd like to contribute to this project, please fork the repository and submit a pull request. All contributions must adhere to the following guidelines:

- Write clear, concise commit messages
- Follow Go best practices
- Ensure that your code passes tests and linters

---

## License

This project is licensed under the MIT License. See the [LICENSE](./LICENSE) file for details.

---

## Contact

For any issues, questions, or feedback, feel free to reach out to:

- **Email**: jooshesh.javid@gmail.com
- **GitHub**: [Flabelatus](https://github.com/Flabelatus)

---

### Notes

- Ensure your `settings.yml` is properly configured before running the application.
- Keep your JWT secret, database credentials, and other sensitive information secure in a secure vault.

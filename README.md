# GOose - The Chirpy Backend

GOose is the Go-powered backend API, a simple microblogging service similar to Twitter. It provides a complete, RESTful API for user management, authentication, creating "chirps" (the platform's posts), and administrative functions.

This project is built with a focus on clean code, standard library features, and robust tooling for database interactions.

## Features

- **User Management**: User registration, login, and profile updates.
- **Authentication**: Secure, stateless authentication using JSON Web Tokens (JWTs), including support for access and refresh tokens.
- **Chirp Functionality**: Full CRUD (Create, Read, Delete) operations for chirps.
- **Content Moderation**: A basic profanity filter that replaces banned words in chirps.
- **Webhook Integration**: An endpoint to receive webhooks from a third-party service Polka (fictional payment service) to handle user upgrades to a premium status.
- **Database Management**: Uses `sqlc` for type-safe Go code generation from SQL queries and `goose` for database migrations.
- **Admin Tools**: Includes endpoints for viewing site metrics and resetting the database in a development environment.
- **Postman Endpoint**: Includes a json that can be imported in Postman containing all the endpoints with different scenarios for validation.

## Tech Stack

- **Language**: Go
- **Database**: PostgreSQL
- **Routing**: `net/http` (Go standard library)
- **Database Tooling**:
  - sqlc: Generates type-safe Go code from SQL.
  - goose: SQL migration tool.
- **Authentication**: golang-jwt/jwt
- **Configuration**: joho/godotenv for environment variable management.

## Getting Started

Follow these instructions to get a local copy up and running.

### Prerequisites

- Go (version 1.21 or newer recommended)
- PostgreSQL
- Goose CLI installed

### Installation

1.  **Clone the repository:**

    ```sh
    git clone https://github.com/leonardomlouzas/GOose
    cd GOose
    ```

2.  **Set up environment variables:**
    Remove the `.example` from the `.env` file in the root of the project and configure the following variables:

    ```env
    # "dev" or "production"
    ENVIRONMENT="dev"

    # Your PostgreSQL connection string
    DB_URL="postgres://user:password@localhost:5432/chirpy?sslmode=disable"

    # A long, random string for signing JWTs
    JWT_SECRET="your-super-secret-jwt-key"

    # The API key for the Polka webhook
    POLKA_KEY="your-polka-api-key"

    # A space-separated list of words to be censored in chirps
    BANNED_WORDS="kerfuffle sharbert foram"
    ```

3.  **Run database migrations:**
    Make sure your PostgreSQL server is running, then execute the following command to set up the database schema:

    ```sh
    goose -dir ./sql/schema postgres "$DB_URL" up
    ```

4.  **Run the server:**
    ```sh
    go build -o out && ./out
    ```
    The server will start on `http://localhost:8080`.

## API Endpoints

The following is a summary of the available API endpoints.

| Method   | Path                  | Description                                                        | Auth Required                      |
| -------- | --------------------- | ------------------------------------------------------------------ | ---------------------------------- |
| `POST`   | `/api/users`          | Create a new user.                                                 | No                                 |
| `PUT`    | `/api/users`          | Update the authenticated user's email/password.                    | Yes (Access Token)                 |
| `POST`   | `/api/login`          | Log in a user to get access/refresh tokens.                        | No                                 |
| `POST`   | `/api/refresh`        | Get a new access token using a refresh token.                      | Yes (Refresh Token)                |
| `POST`   | `/api/revoke`         | Revoke a refresh token.                                            | Yes (Refresh Token)                |
| `POST`   | `/api/chirps`         | Post a new chirp.                                                  | Yes (Access Token)                 |
| `GET`    | `/api/chirps`         | Get all chirps. Supports `?author_id` and `?sort_by` query params. | No                                 |
| `GET`    | `/api/chirps/{id}`    | Get a single chirp by its ID.                                      | No                                 |
| `DELETE` | `/api/chirps/{id}`    | Delete a chirp.                                                    | Yes (Access Token, must be author) |
| `POST`   | `/api/polka/webhooks` | Webhook for handling `user.upgraded` events.                       | Yes (Polka API Key)                |
| `GET`    | `/api/healthz`        | Health check endpoint.                                             | No                                 |
| `GET`    | `/admin/metrics`      | View site metrics (file server hits).                              | No                                 |
| `POST`   | `/admin/reset`        | Reset site metrics and database.                                   | No (Only available in `dev` mode)  |

### Authentication

- **Access Tokens**: Passed as a Bearer token in the `Authorization` header.
  `Authorization: Bearer <your_access_token>`
- **Refresh Tokens**: Passed as a Bearer token for the `/api/refresh` and `/api/revoke` endpoints.
  `Authorization: Bearer <your_refresh_token>`
- **API Keys**: Passed as an ApiKey token for protected webhooks.
  `Authorization: ApiKey <your_api_key>`

# School Management API

A secure, production-ready REST API built in Go for managing core school operations, including students, teachers, and executive/admin staff. The project follows a clean layered architecture, emphasizes security, and is built entirely on Go's standard library without external routing frameworks.

---

# Features

- **RESTful API** built using Go's standard library (`net/http`)

- **JWT Authentication** — stateless, token-based authentication with configurable token expiry

- **Role-Based Access Control (RBAC)** — only executives (Execs) authenticate. Six privileged roles have full access while any other authenticated exec is limited to read-only (`GET`) operations

- **Password Reset Flow** — secure, time-limited reset tokens for credential recovery

- **HTTPS/TLS Support** — configurable through SSL certificates

- **API Root Endpoint** — exposes API metadata, version, status, and documentation link through `GET /`

- **Security-first middleware pipeline**
  - CORS
  - Security Headers
  - HTTP Parameter Pollution (HPP) Protection
  - Input Sanitization
  - Rate Limiting
  - Gzip Compression
  - Response Time Logging

- **Complete resource management**
  - Students CRUD
  - Teachers CRUD
  - Executives (Execs) CRUD

- **Repository Pattern** — SQL logic separated from HTTP handlers

- **Layered Architecture** — handlers, repositories, models, routers, and middleware remain isolated and maintainable

- **Environment-based configuration** using `.env`

---

# Tech Stack

| Layer          | Technology                                                               |
| -------------- | ------------------------------------------------------------------------ |
| Language       | Go                                                                       |
| Routing        | Standard Library `net/http.ServeMux` (Go 1.22+ Method + Pattern Routing) |
| Authentication | JWT                                                                      |
| Database       | MySQL                                                                    |
| Configuration  | `.env`                                                                   |
| TLS            | OpenSSL Certificates                                                     |

---

# Project Structure

```text
SCHOOL-MANAGEMENT-API/
├── cert/
├── cmd/
│   └── api/
│       ├── .env
│       └── server.go
├── internal/
│   ├── api/
│   │   ├── handlers/
│   │   │   ├── execs.go
│   │   │   ├── helpers.go
│   │   │   ├── students.go
│   │   │   └── teachers.go
│   │   │
│   │   ├── middlewares/
│   │   │   ├── compression.go
│   │   │   ├── cors.go
│   │   │   ├── exclude_routes.go
│   │   │   ├── hpp.go
│   │   │   ├── jwt_middleware.go
│   │   │   ├── rate_limiter.go
│   │   │   ├── responseTime.go
│   │   │   ├── sanitize.go
│   │   │   └── security_headers.go
│   │   │
│   │   ├── models/
│   │   │   ├── exec.go
│   │   │   ├── student.go
│   │   │   └── teacher.go
│   │   │
│   │   ├── repositories/
│   │   │   └── sqlconnect/
│   │   │       ├── execs_CRUD.go
│   │   │       ├── helpers.go
│   │   │       ├── sqlconfig.go
│   │   │       ├── students_CRUD.go
│   │   │       └── teachers_CRUD.go
│   │   │
│   │   └── router/
│   │       ├── execs_router.go
│   │       ├── home_router.go
│   │       ├── router.go
│   │       ├── students_router.go
│   │       └── teachers_router.go
│   │
│   └── pkg/
│
├── .gitignore
├── go.mod
├── go.sum
├── openssl.cnf
└── README.md
```

---

# Architecture Overview

The application follows a layered request lifecycle.

```text
Request
   │
   ▼
Middleware Chain
(CORS → Security Headers → Rate Limiter → HPP → Sanitization → JWT → Compression → Response Time)
   │
   ▼
Router
   │
   ▼
Handler
   │
   ▼
Repository (sqlconnect)
   │
   ▼
Database
   │
   ▼
Response
```

### Layers

- **Router** maps incoming HTTP requests to their corresponding handlers.

- **Handlers** validate requests, decode payloads, perform business logic, and communicate with the repository layer.

- **Repositories** isolate database access from application logic and contain all SQL operations.

- **Models** define the application's core entities.

- **Middlewares** provide cross-cutting functionality such as authentication, security, logging, compression, and rate limiting.

---

# Getting Started

## Prerequisites

- Go 1.22 or newer
- MySQL
- OpenSSL (only required for generating local TLS certificates)

---

## 1. Clone the Repository

```bash
git clone https://github.com/BurhaanAshraf/school-management-api
cd SCHOOL-MANAGEMENT-API
```

---

## 2. Configure Environment Variables

Create or update:

```text
cmd/api/.env
```

```dotenv
DB_HOST=
DB_PORT=
DB_USER=
DB_PASSWORD=
DB_NAME=

API_PORT=

JWT_SECRET=
JWT_EXPIRES_IN=
RESET_TOKEN_EXP_DURATION=

CERT_FILE=
KEY_FILE=
```

| Variable                   | Description                 |
| -------------------------- | --------------------------- |
| `DB_HOST`                  | Database host               |
| `DB_PORT`                  | Database port               |
| `DB_USER`                  | Database username           |
| `DB_PASSWORD`              | Database password           |
| `DB_NAME`                  | Database name               |
| `API_PORT`                 | API server port             |
| `JWT_SECRET`               | Secret used to sign JWTs    |
| `JWT_EXPIRES_IN`           | JWT expiry duration         |
| `RESET_TOKEN_EXP_DURATION` | Password reset token expiry |
| `CERT_FILE`                | SSL certificate             |
| `KEY_FILE`                 | SSL private key             |

> Never commit a populated `.env` file. Keep it inside `.gitignore` and use environment variables or a secrets manager in production.

---

## 3. Install Dependencies

```bash
go mod download
```

Since `go.mod` and `go.sum` are committed, this installs the exact dependency versions used by the project.

---

## 4. Run the Server

```bash
go run cmd/api/server.go
```

The server starts on the port configured by `API_PORT`.

You can verify that the API is running by visiting:

```text
GET /
```

Example response:

```json
{
  "name": "School Management API",
  "status": "running",
  "version": "1.0",
  "docs": "https://github.com/BurhaanAshraf/school-management-api"
}
```

---

# Authentication & Authorization

Every protected endpoint requires a valid JWT.

```text
Authorization: Bearer <token>
```

The following endpoints are public:

- `POST /execs/login`
- `POST /execs/forgotpassword`
- `POST /execs/reset/resetpassword/{resetcode}`

All Student and Teacher endpoints require an authenticated Executive account.

## Role-Based Access

The following executive roles have full permissions:

- `super_admin`
- `principal`
- `vice_principal`
- `registrar`
- `student_affairs`
- `secretary`

These roles may perform:

- GET
- POST
- PUT
- PATCH
- DELETE

Any authenticated executive outside these roles is restricted to **read-only (`GET`)** operations.

## Additional Authentication Details

- JWT expiry is configurable using `JWT_EXPIRES_IN`.

- Password reset uses secure, time-limited reset codes.

- `/execs/logout` invalidates the current session.

- `/execs/{id}/updatepassword` allows an authenticated executive to change their password.

# API Endpoints

## API Information

| Method | Endpoint | Description                                                           | Access |
| ------ | -------- | --------------------------------------------------------------------- | ------ |
| GET    | `/`      | Returns API metadata, current status, version, and documentation link | Public |

---

## Executive (Execs)

| Method | Endpoint                                 | Description                     | Access     |
| ------ | ---------------------------------------- | ------------------------------- | ---------- |
| GET    | `/execs`                                 | Get all executives              | Any Exec   |
| POST   | `/execs`                                 | Create a new executive          | Privileged |
| PATCH  | `/execs`                                 | Bulk update executives          | Privileged |
| GET    | `/execs/{id}`                            | Get executive by ID             | Any Exec   |
| PATCH  | `/execs/{id}`                            | Partially update an executive   | Privileged |
| DELETE | `/execs/{id}`                            | Delete an executive             | Privileged |
| POST   | `/execs/{id}/updatepassword`             | Update executive password       | Any Exec   |
| POST   | `/execs/login`                           | Authenticate executive          | Public     |
| POST   | `/execs/logout`                          | Logout current executive        | Any Exec   |
| POST   | `/execs/forgotpassword`                  | Request password reset          | Public     |
| POST   | `/execs/reset/resetpassword/{resetcode}` | Reset password using reset code | Public     |

---

## Students

All student endpoints require an authenticated Executive account.

| Method | Endpoint         | Description                | Access     |
| ------ | ---------------- | -------------------------- | ---------- |
| GET    | `/students`      | Get all students           | Any Exec   |
| POST   | `/students`      | Create a student           | Privileged |
| PATCH  | `/students`      | Bulk update students       | Privileged |
| DELETE | `/students`      | Bulk delete students       | Privileged |
| GET    | `/students/{id}` | Get student by ID          | Any Exec   |
| PUT    | `/students/{id}` | Replace a student          | Privileged |
| PATCH  | `/students/{id}` | Partially update a student | Privileged |
| DELETE | `/students/{id}` | Delete a student           | Privileged |

---

## Teachers

All teacher endpoints require an authenticated Executive account.

| Method | Endpoint                      | Description                                      | Access     |
| ------ | ----------------------------- | ------------------------------------------------ | ---------- |
| GET    | `/teachers`                   | Get all teachers                                 | Any Exec   |
| POST   | `/teachers`                   | Create a teacher                                 | Privileged |
| PATCH  | `/teachers`                   | Bulk update teachers                             | Privileged |
| DELETE | `/teachers`                   | Bulk delete teachers                             | Privileged |
| GET    | `/teachers/{id}`              | Get teacher by ID                                | Any Exec   |
| PUT    | `/teachers/{id}`              | Replace a teacher                                | Privileged |
| PATCH  | `/teachers/{id}`              | Partially update a teacher                       | Privileged |
| DELETE | `/teachers/{id}`              | Delete a teacher                                 | Privileged |
| GET    | `/teachers/{id}/students`     | Get students assigned to a teacher               | Any Exec   |
| GET    | `/teachers/{id}/studentcount` | Get the number of students assigned to a teacher | Any Exec   |

> **Access Levels**
>
> - **Public** — No authentication required.
> - **Any Exec** — Any authenticated executive.
> - **Privileged** — One of the following roles:
>   - `super_admin`
>   - `principal`
>   - `vice_principal`
>   - `registrar`
>   - `student_affairs`
>   - `secretary`

---

# Security

The API is protected by a layered security pipeline.

- JWT Authentication
- Role-Based Access Control (RBAC)
- HTTPS/TLS Support
- Security Headers
- CORS
- HTTP Parameter Pollution (HPP) Protection
- Input Sanitization
- Rate Limiting
- Gzip Response Compression
- Response Time Logging

Every incoming request passes through the middleware chain before reaching the router, ensuring authentication, validation, sanitization, and logging are consistently applied across the application.

---

## 🚀 Live Deployment

The API is deployed on **Render** and is publicly accessible.

- **API Landing Page:** https://school-management-api-yxru.onrender.com/


> **Note:** The API is hosted on Render's free tier. If the service has been idle, the first request may take a few seconds while the server starts.

---

# Contributing

Contributions are welcome.

1. Fork the repository.
2. Create a feature branch.

```bash
git checkout -b feature/your-feature
```

3. Commit your changes.

```bash
git commit -m "feat: describe your feature"
```

4. Push the branch.

```bash
git push origin feature/your-feature
```

5. Open a Pull Request.

---

# License

This project currently has **no open-source license**.

Until a license is added, **all rights are reserved** by the author. The source code may not be copied, modified, redistributed, or used commercially without explicit permission.

---

# Author

**Burhaan Ashraf**

GitHub:
https://github.com/BurhaanAshraf

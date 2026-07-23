# School Management API

A secure, production-ready REST API built in Go, designed to manage core school operations — students, teachers, and executive/admin staff — with a hardened middleware pipeline and clean, layered architecture.

---

## Features

- **JWT Authentication** — stateless, token-based auth for protected routes, with configurable token expiry

- **Role-Based Access Control** — only Execs can authenticate; six defined exec roles hold full permissions, while any other authenticated exec is limited to read-only (`GET`) access

- **Password Reset Flow** — time-limited reset tokens for secure credential recovery

- **HTTPS/TLS Support** — configurable via cert/key files for encrypted traffic

- **Security-first middleware stack**
  - CORS handling
  - Security headers (XSS, clickjacking, MIME-sniffing protection, etc.)
  - HTTP Parameter Pollution (HPP) protection
  - Input sanitization
  - Rate limiting to prevent abuse/brute-force
  - Response compression (gzip) for faster payloads
  - Response time tracking/logging

- **Role-based resource management**
  - Students CRUD
  - Teachers CRUD
  - Executives (Execs) CRUD

- **Clean repository pattern** — database logic isolated from business logic via `sqlconnect`

- **Layered architecture** — clear separation between `handlers`, `middlewares`, `models`, `repositories`, and `router`

- **Environment-based configuration** via `.env`

---

## Tech Stack

| Layer     | Technology                                                                                                   |
| --------- | ------------------------------------------------------------------------------------------------------------ |
| Language  | Go                                                                                                           |
| Routing   | Standard library `net/http.ServeMux` (Go 1.22+ method + path-pattern routing) — no external router framework |
| Auth      | JWT + session login/logout, password reset via time-limited reset codes                                      |
| Database  | MySQL (via `sqlconnect` repository layer)                                                                    |
| Config    | `.env` file                                                                                                  |
| TLS/Certs | `openssl.cnf` + `cert/` directory (`CERT_FILE` / `KEY_FILE`)                                                 |

---

## Project Structure

```
SCHOOL-MANAGEMENT-API/
├── cert/                          # SSL/TLS certificates
├── cmd/
│   └── api/
│       ├── .env                   # Environment variables
│       └── server.go              # Application entry point
├── internal/
│   ├── api/
│   │   ├── handlers/              # HTTP request handlers
│   │   │   ├── execs.go
│   │   │   ├── helpers.go
│   │   │   ├── students.go
│   │   │   └── teachers.go
│   │   │
│   │   ├── middlewares/           # Middleware chain
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
│   │   ├── models/                # Data models
│   │   │   ├── exec.go
│   │   │   ├── student.go
│   │   │   └── teacher.go
│   │   │
│   │   ├── repositories/
│   │   │   └── sqlconnect/        # Database access layer
│   │   │       ├── execs_CRUD.go
│   │   │       ├── helpers.go
│   │   │       ├── sqlconfig.go
│   │   │       ├── students_CRUD.go
│   │   │       └── teachers_CRUD.go
│   │   │
│   │   └── router/                # Route definitions
│   │       ├── execs_router.go
│   │       ├── router.go
│   │       ├── students_router.go
│   │       └── teachers_router.go
│   │
│   └── pkg/                       # Shared/reusable packages
│
├── .gitignore
├── go.mod
├── go.sum
├── openssl.cnf
└── README.md
```

---

## Architecture Overview

The API follows a layered, request-lifecycle-driven architecture:

```
Request
   │
   ▼
Middleware Chain
 (CORS → Security Headers → Rate Limiter → HPP → Sanitize → JWT Auth → Compression → Response Time)
   │
   ▼
Router  →  Handler  →  Repository (sqlconnect)  →  Database
   │
   ▼
Response
```

- **`router/`** maps HTTP routes to handler functions per resource (students, teachers, execs).

- **`handlers/`** parse requests, validate input, and call the repository layer.

- **`repositories/sqlconnect/`** contains all raw DB interaction (CRUD operations), keeping SQL logic out of handlers.

- **`models/`** define the shape of each core entity (`Student`, `Teacher`, `Exec`).

- **`middlewares/`** wrap every request with cross-cutting concerns like security, rate limiting, and logging.

---

## Getting Started

### Prerequisites

- Go 1.22+ installed (required for stdlib method+pattern routing used in `router.go`)

- MySQL server running and accessible

- OpenSSL (for local cert generation, if running HTTPS locally)

### 1. Clone the repository

```bash
git clone https://github.com/BurhaanAshraf/school-management-api
cd SCHOOL-MANAGEMENT-API
```

### 2. Configure environment variables

Create/update `cmd/api/.env`:

```dotenv
# Database
DB_HOST=
DB_PORT=
DB_USER=
DB_PASSWORD=
DB_NAME=

# Server
API_PORT=

# Auth
JWT_SECRET=
JWT_EXPIRES_IN=
RESET_TOKEN_EXP_DURATION=

# TLS / HTTPS
CERT_FILE=
KEY_FILE=
```

| Variable                   | Description                                      |
| -------------------------- | ------------------------------------------------ |
| `DB_HOST`                  | Database host address                            |
| `DB_PORT`                  | Database port                                    |
| `DB_USER`                  | Database username                                |
| `DB_PASSWORD`              | Database password                                |
| `DB_NAME`                  | Name of the database to connect to               |
| `API_PORT`                 | Port the API server listens on                   |
| `JWT_SECRET`               | Secret key used to sign/verify JWTs              |
| `JWT_EXPIRES_IN`           | JWT token lifetime (e.g. `15m`, `24h`)           |
| `RESET_TOKEN_EXP_DURATION` | Expiry duration for password-reset tokens        |
| `CERT_FILE`                | Path to the SSL/TLS certificate file (for HTTPS) |
| `KEY_FILE`                 | Path to the SSL/TLS private key file (for HTTPS) |

Never commit a real, filled-in `.env` to version control — keep it in `.gitignore` (already set up here) and share values through a secure channel or secrets manager instead.

### 3. Install dependencies

Since `go.mod` and `go.sum` are already committed to this repo, dependencies are locked and reproducible — just download them:

```bash
go mod download
```

This reads `go.mod`/`go.sum` and fetches the exact dependency versions used to build the project, without modifying either file. Use `go mod tidy` instead only if you're adding/removing imports and need those files to be regenerated.

### 4. Run the server

```bash
go run cmd/api/server.go
```

The API should now be running at `http://localhost:<API_PORT>` (as set in your `.env`), served over HTTPS if `CERT_FILE`/`KEY_FILE` are configured.

---

## Authentication & Authorization

All routes require a valid JWT passed in the `Authorization` header:

```
Authorization: Bearer <your_token>
```

except these three, which are public (no token required, since they're how a token is obtained in the first place):

- `POST /execs/login`

- `POST /execs/forgotpassword`

- `POST /execs/reset/resetpassword/{resetcode}`

Every other route — including all Student and Teacher endpoints — is JWT-protected. There is no separate student/teacher login: only Execs authenticate, and an authenticated exec's token is what grants access to student and teacher data.

### Role-gating

Beyond simple authentication, access is role-gated at the exec level. The system recognizes six exec roles, each of which is granted full permissions:

- `super_admin`

- `principal`

- `vice_principal`

- `registrar`

- `student_affairs`

- `secretary`

Any exec assigned one of the six roles above can perform all operations — `GET`, `POST`, `PUT`, `PATCH`, and `DELETE` — on students and teachers.

Any authenticated exec outside these six roles (i.e. with no role or an unrecognized role) is restricted to `GET` (read-only) access. `POST`/`PUT`/`PATCH`/`DELETE` requests from such execs are expected to be rejected by the handler/middleware layer even though their token is otherwise valid.

### Other auth details

- Token lifetime is configurable via `JWT_EXPIRES_IN`

- Password reset uses short-lived, single-use reset codes (`RESET_TOKEN_EXP_DURATION`): request one via `/execs/forgotpassword`, consume it via `/execs/reset/resetpassword/{resetcode}`

- `/execs/logout` invalidates the current session/token

- `/execs/{id}/updatepassword` lets an authenticated exec update their own password directly (as opposed to the forgot/reset flow for logged-out users)

---

## API Endpoints

Legend: **Public** — no token required · **Any exec** — any authenticated exec, read-only unless role is one of the six defined roles · **Privileged** — requires one of the six defined exec roles (`super_admin`, `principal`, `vice_principal`, `registrar`, `student_affairs`, `secretary`)

### Execs (Admin/Executive Staff)

| Method | Endpoint                                 | Description                                 | Access     |
| ------ | ---------------------------------------- | ------------------------------------------- | ---------- |
| GET    | `/execs`                                 | Get all executives                          | Any exec   |
| POST   | `/execs`                                 | Create a new executive                      | Privileged |
| PATCH  | `/execs`                                 | Bulk-patch executives                       | Privileged |
| GET    | `/execs/{id}`                            | Get a single executive by ID                | Any exec   |
| PATCH  | `/execs/{id}`                            | Partially update a single executive         | Privileged |
| DELETE | `/execs/{id}`                            | Delete a single executive                   | Privileged |
| POST   | `/execs/{id}/updatepassword`             | Update password for a given executive       | Any exec   |
| POST   | `/execs/login`                           | Log in an executive                         | Public     |
| POST   | `/execs/logout`                          | Log out the current executive               | Any exec   |
| POST   | `/execs/forgotpassword`                  | Request a password reset (sends reset code) | Public     |
| POST   | `/execs/reset/resetpassword/{resetcode}` | Reset password using a valid reset code     | Public     |

### Students

All routes below require a valid exec JWT — students do not log in themselves.

| Method | Endpoint         | Description                       | Access     |
| ------ | ---------------- | --------------------------------- | ---------- |
| GET    | `/students`      | Get all students                  | Any exec   |
| POST   | `/students`      | Create a new student              | Privileged |
| PATCH  | `/students`      | Bulk-patch students               | Privileged |
| DELETE | `/students`      | Bulk-delete students              | Privileged |
| GET    | `/students/{id}` | Get a single student by ID        | Any exec   |
| PUT    | `/students/{id}` | Fully update a single student     | Privileged |
| PATCH  | `/students/{id}` | Partially update a single student | Privileged |
| DELETE | `/students/{id}` | Delete a single student           | Privileged |

### Teachers

All routes below require a valid exec JWT — teachers do not log in themselves.

| Method | Endpoint                      | Description                                     | Access     |
| ------ | ----------------------------- | ----------------------------------------------- | ---------- |
| GET    | `/teachers`                   | Get all teachers                                | Any exec   |
| POST   | `/teachers`                   | Create a new teacher                            | Privileged |
| PATCH  | `/teachers`                   | Bulk-patch teachers                             | Privileged |
| DELETE | `/teachers`                   | Bulk-delete teachers                            | Privileged |
| GET    | `/teachers/{id}`              | Get a single teacher by ID                      | Any exec   |
| PUT    | `/teachers/{id}`              | Fully update a single teacher                   | Privileged |
| PATCH  | `/teachers/{id}`              | Partially update a single teacher               | Privileged |
| DELETE | `/teachers/{id}`              | Delete a single teacher                         | Privileged |
| GET    | `/teachers/{id}/students`     | Get all students assigned to a teacher          | Any exec   |
| GET    | `/teachers/{id}/studentcount` | Get the count of students assigned to a teacher | Any exec   |

Note: some resources support `PUT` for full updates (Students, Teachers) while Execs only expose `PATCH` — worth keeping in mind for client implementations. All write operations (`POST`/`PUT`/`PATCH`/`DELETE`) require the exec to hold one of the six defined roles; any other authenticated exec is limited to `GET` (read-only).

---

## Security Measures

This API takes security seriously with a dedicated middleware layer:

- **CORS** — restricts which origins can access the API

- **Security Headers** — mitigates common web vulnerabilities (XSS, clickjacking, sniffing)

- **Rate Limiting** — throttles requests to prevent abuse and brute-force attacks

- **HPP Protection** — guards against HTTP Parameter Pollution attacks

- **Sanitization** — cleans incoming input to prevent injection attacks

- **JWT Middleware** — enforces authenticated access on protected routes

- **Role-Based Access Control (RBAC)** — six defined exec roles (`super_admin`, `principal`, `vice_principal`, `registrar`, `student_affairs`, `secretary`) hold full permissions (`GET`/`POST`/`PUT`/`PATCH`/`DELETE`); any other authenticated exec is restricted to `GET` (read-only)

- **HTTPS/TLS** — supported via certs in `cert/` and `openssl.cnf`

---

## Contributing

1. Fork the repo

2. Create a feature branch (`git checkout -b feature/your-feature`)

3. Commit your changes

4. Push to the branch

5. Open a Pull Request

---

## License

No open-source license has been applied to this project. By default, this means all rights are reserved by the author — the code may not be copied, modified, or redistributed without explicit permission.

---

## Author / Maintainer

**Burhaan Ashraf**
GitHub: https://github.com/burhaanAshraf

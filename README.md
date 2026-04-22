# Golang JWT Authentication API

A REST API built with Go, Gin, and MongoDB that implements JWT authentication with role-based access control.

## Features

- User Signup with password hashing
- User Login with JWT token generation
- Role based access control (ADMIN/USER)
- Protected routes with JWT middleware
- Refresh token support
- Paginated user listing

## Tech Stack

- **Go** — Programming language
- **Gin** — HTTP web framework
- **MongoDB** — Database
- **JWT** — Authentication
- **bcrypt** — Password hashing

## Project Structure

```
golang-jwt-project/
├── controllers/
│   └── userController.go
├── database/
│   └── databaseConnection.go
├── helpers/
│   └── tokenHelper.go
├── middleware/
│   └── authMiddleware.go
├── models/
│   └── userModel.go
├── routes/
│   ├── authRouter.go
│   └── userRouter.go
├── .env
├── .gitignore
├── go.mod
└── main.go
```

## Getting Started

### Prerequisites
- Go 1.21+
- MongoDB Atlas account

### Installation

1. Clone the repo
```bash
git clone https://github.com/yourusername/golang-jwt-project
cd golang-jwt-project
```

2. Install dependencies
```bash
go mod tidy
```

3. Create `.env` file in root folder
```env
PORT=9000
MONGODB_URL=your_mongodb_url
SECRET_KEY=your_secret_key
```

4. Run the server
```bash
go run main.go
```

## API Endpoints

### Public Routes
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | /users/signup | Register a new user |
| POST | /users/login | Login and get JWT token |

### Protected Routes (requires token in header)
| Method | Endpoint | Description | Access |
|--------|----------|-------------|--------|
| GET | /users | Get all users (paginated) | ADMIN only |
| GET | /users/:user_id | Get single user | ADMIN or own profile |

## Usage

### Signup
```bash
curl -X POST http://localhost:9000/users/signup \
-H "Content-Type: application/json" \
-d '{
    "first_name": "John",
    "last_name": "Doe",
    "email": "john@gmail.com",
    "password": "123456",
    "phone": "1234567890",
    "user_type": "ADMIN"
}'
```

### Login
```bash
curl -X POST http://localhost:9000/users/login \
-H "Content-Type: application/json" \
-d '{
    "email": "john@gmail.com",
    "password": "123456"
}'
```

### Get All Users (ADMIN only)
```bash
curl "http://localhost:9000/users?recordPerPage=10&page=1" \
-H "token: your_jwt_token"
```

### Get Single User
```bash
curl http://localhost:9000/users/user_id_here \
-H "token: your_jwt_token"
```

## Pagination

The `/users` endpoint supports pagination via query parameters:

| Parameter | Description | Default |
|-----------|-------------|---------|
| recordPerPage | Number of users per page | 10 |
| page | Page number | 1 |

Example:
```
GET /users?recordPerPage=5&page=2
```

## Environment Variables

| Variable | Description |
|----------|-------------|
| PORT | Server port (default 9000) |
| MONGODB_URL | MongoDB Atlas connection string |
| SECRET_KEY | JWT signing secret key |

## User Model

```json
{
    "first_name": "string (required, min=2, max=100)",
    "last_name":  "string (required, min=2, max=100)",
    "email":      "string (required, valid email)",
    "password":   "string (required, min=6)",
    "phone":      "string (required)",
    "user_type":  "string (required, ADMIN or USER)"
}
```

## Security

- Passwords are hashed using **bcrypt** before storing in DB
- JWT tokens expire after **24 hours**
- Refresh tokens expire after **7 days**
- Role based access control for all protected routes
- Never commit your `.env` file to version control

## License

MIT

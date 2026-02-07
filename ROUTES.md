# API Routes Quick Reference

**Base URL:** `http://localhost:8080`

## Authentication Routes

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/users` | Create a new user |
| GET | `/users/{id}` | Get user by ID |
| PUT | `/users/{id}` | Update user |
| DELETE | `/users/{id}` | Delete user |
| POST | `/login` | User login |
| POST | `/set-password` | Set user password |
| POST | `/user-look-up` | Lookup user by phone number |

## Merchant Routes

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/merchants` | Create merchant (multipart/form-data) |
| GET | `/merchants/{id}` | Get merchant by ID |
| PUT | `/merchants/{id}` | Update merchant (multipart/form-data) |
| DELETE | `/merchants/{id}` | Delete merchant |
| GET | `/merchants` | Get all merchants |

## Branch Routes

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/branches` | Create branch |
| GET | `/branches/{id}` | Get branch by ID |
| PUT | `/branches/{id}` | Update branch |
| DELETE | `/branches/{id}` | Delete branch |

## Menu Routes

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/menus` | Create menu (multipart/form-data) |
| GET | `/menus/{id}` | Get menu by ID |
| PUT | `/menus/{id}` | Update menu (multipart/form-data) |
| DELETE | `/menus/{id}` | Delete menu |

## Total Routes: 20

For detailed testing instructions, see [API_TESTING.md](./API_TESTING.md)

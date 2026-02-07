# API Testing Guide

Base URL: `http://localhost:8080`

## Table of Contents
- [Authentication Routes](#authentication-routes)
- [Merchant Routes](#merchant-routes)
- [Branch Routes](#branch-routes)
- [Menu Routes](#menu-routes)

---

## Authentication Routes

### 1. Create User
**POST** `/users`

**Request Body:**
```json
{
  "phone_number": "+251911234567",
  "full_name": "John Doe",
  "branch_id": "branch-uuid-here",
  "merchant_id": "merchant-uuid-here"
}
```

**cURL:**
```bash
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{
    "phone_number": "+251911234567",
    "full_name": "John Doe",
    "branch_id": "branch-uuid-here",
    "merchant_id": "merchant-uuid-here"
  }'
```

**Response:** `200 OK`
```json
{
  "data": {
    "phone_number": "+251911234567",
    "full_name": "John Doe",
    "branch_id": "branch-uuid-here",
    "merchant_id": "merchant-uuid-here"
  },
  "message": "User created successfully",
  "status_code": 200
}
```

---

### 2. Get User by ID
**GET** `/users/{id}`

**cURL:**
```bash
curl -X GET http://localhost:8080/users/USER_ID_HERE
```

**Response:** `200 OK`
```json
{
  "data": {
    "id": "user-id",
    "phone_number": "+251911234567",
    "full_name": "John Doe",
    "role": "branch_manager",
    "branch_id": "branch-uuid",
    "is_locked": false,
    "is_first_login": true
  },
  "message": "User fetched successfully",
  "status_code": 200
}
```

---

### 3. Update User
**PUT** `/users/{id}`

**Request Body:**
```json
{
  "phone_number": "+251911234567",
  "full_name": "John Updated",
  "branch_id": "branch-uuid-here",
  "merchant_id": "merchant-uuid-here"
}
```

**cURL:**
```bash
curl -X PUT http://localhost:8080/users/USER_ID_HERE \
  -H "Content-Type: application/json" \
  -d '{
    "phone_number": "+251911234567",
    "full_name": "John Updated",
    "branch_id": "branch-uuid-here",
    "merchant_id": "merchant-uuid-here"
  }'
```

**Response:** `200 OK`
```json
{
  "data": {
    "phone_number": "+251911234567",
    "full_name": "John Updated",
    "branch_id": "branch-uuid-here",
    "merchant_id": "merchant-uuid-here"
  },
  "message": "User updated successfully",
  "status_code": 200
}
```

---

### 4. Delete User
**DELETE** `/users/{id}`

**cURL:**
```bash
curl -X DELETE http://localhost:8080/users/USER_ID_HERE
```

**Response:** `200 OK`
```json
{
  "message": "User deleted successfully",
  "status_code": 200
}
```

---

### 5. Login
**POST** `/login`

**Request Body:**
```json
{
  "phone_number": "+251911234567",
  "password": "your-password"
}
```

**cURL:**
```bash
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{
    "phone_number": "+251911234567",
    "password": "your-password"
  }'
```

**Response:** `200 OK`
```json
{
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": "user-id",
      "phone_number": "+251911234567",
      "full_name": "John Doe",
      "role": "branch_manager"
    }
  },
  "message": "Login successful",
  "status_code": 200
}
```

---

### 6. Set Password
**POST** `/set-password`

**Request Body:**
```json
{
  "phone_number": "+251911234567",
  "password": "new-password-123"
}
```

**cURL:**
```bash
curl -X POST http://localhost:8080/set-password \
  -H "Content-Type: application/json" \
  -d '{
    "phone_number": "+251911234567",
    "password": "new-password-123"
  }'
```

**Response:** `200 OK`
```json
{
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": "user-id",
      "phone_number": "+251911234567",
      "full_name": "John Doe"
    }
  },
  "message": "Password set successfully",
  "status_code": 200
}
```

---

### 7. User Lookup
**POST** `/user-look-up`

**Request Body:**
```json
{
  "phone_number": "+251911234567"
}
```

**cURL:**
```bash
curl -X POST http://localhost:8080/user-look-up \
  -H "Content-Type: application/json" \
  -d '{
    "phone_number": "+251911234567"
  }'
```

**Response:** `200 OK`
```json
{
  "data": {
    "id": "user-id",
    "phone_number": "+251911234567",
    "full_name": "John Doe",
    "role": "branch_manager"
  },
  "message": "User looked up successfully",
  "status_code": 200
}
```

---

## Merchant Routes

### 1. Create Merchant
**POST** `/merchants`

**Request:** Multipart form data
- `name` (string): Merchant name
- `logo` (file): Logo image file

**cURL:**
```bash
curl -X POST http://localhost:8080/merchants \
  -F "name=My Restaurant" \
  -F "logo=@/path/to/logo.png"
```

**Response:** `200 OK`
```json
{
  "data": {
    "id": "merchant-id",
    "name": "My Restaurant",
    "logo": "logo-url",
    "created_at": "2024-02-07T10:00:00Z"
  },
  "message": "Merchant created successfully",
  "status_code": 200
}
```

---

### 2. Get Merchant by ID
**GET** `/merchants/{id}`

**cURL:**
```bash
curl -X GET http://localhost:8080/merchants/MERCHANT_ID_HERE
```

**Response:** `200 OK`
```json
{
  "data": {
    "id": "merchant-id",
    "name": "My Restaurant",
    "logo": "logo-url",
    "created_at": "2024-02-07T10:00:00Z"
  },
  "message": "Merchant fetched successfully",
  "status_code": 200
}
```

---

### 3. Update Merchant
**PUT** `/merchants/{id}`

**Request:** Multipart form data
- `name` (string): Merchant name
- `logo` (file, optional): Logo image file

**cURL:**
```bash
curl -X PUT http://localhost:8080/merchants/MERCHANT_ID_HERE \
  -F "name=Updated Restaurant Name" \
  -F "logo=@/path/to/new-logo.png"
```

**Response:** `200 OK`
```json
{
  "data": {
    "id": "merchant-id",
    "name": "Updated Restaurant Name",
    "logo": "new-logo-url"
  },
  "message": "Merchant updated successfully",
  "status_code": 200
}
```

---

### 4. Delete Merchant
**DELETE** `/merchants/{id}`

**cURL:**
```bash
curl -X DELETE http://localhost:8080/merchants/MERCHANT_ID_HERE
```

**Response:** `200 OK`
```json
{
  "message": "Merchant deleted successfully",
  "status_code": 200
}
```

---

### 5. Get All Merchants
**GET** `/merchants`

**cURL:**
```bash
curl -X GET http://localhost:8080/merchants
```

**Response:** `200 OK`
```json
{
  "data": [
    {
      "id": "merchant-id-1",
      "name": "Restaurant 1",
      "logo": "logo-url-1"
    },
    {
      "id": "merchant-id-2",
      "name": "Restaurant 2",
      "logo": "logo-url-2"
    }
  ],
  "message": "Merchants fetched successfully",
  "status_code": 200
}
```

---

## Branch Routes

### 1. Create Branch
**POST** `/branches`

**Request Body:**
```json
{
  "merchant_id": "merchant-uuid-here",
  "branch_name": "Downtown Branch",
  "address": "123 Main Street",
  "phone_number": "+251911234567"
}
```

**cURL:**
```bash
curl -X POST http://localhost:8080/branches \
  -H "Content-Type: application/json" \
  -d '{
    "merchant_id": "merchant-uuid-here",
    "branch_name": "Downtown Branch",
    "address": "123 Main Street",
    "phone_number": "+251911234567"
  }'
```

**Response:** `200 OK`
```json
{
  "data": {
    "id": "branch-id",
    "merchant_id": "merchant-uuid-here",
    "branch_name": "Downtown Branch",
    "address": "123 Main Street",
    "phone_number": "+251911234567",
    "created_at": "2024-02-07T10:00:00Z"
  },
  "message": "Branch created successfully",
  "status_code": 200
}
```

---

### 2. Get Branch by ID
**GET** `/branches/{id}`

**cURL:**
```bash
curl -X GET http://localhost:8080/branches/BRANCH_ID_HERE
```

**Response:** `200 OK`
```json
{
  "data": {
    "id": "branch-id",
    "merchant_id": "merchant-id",
    "branch_name": "Downtown Branch",
    "address": "123 Main Street",
    "phone_number": "+251911234567"
  },
  "message": "Branch fetched successfully",
  "status_code": 200
}
```

---

### 3. Update Branch
**PUT** `/branches/{id}`

**Request Body:**
```json
{
  "id": "branch-id",
  "branch_name": "Updated Branch Name",
  "address": "456 New Street",
  "phone_number": "+251911234568"
}
```

**cURL:**
```bash
curl -X PUT http://localhost:8080/branches/BRANCH_ID_HERE \
  -H "Content-Type: application/json" \
  -d '{
    "id": "branch-id",
    "branch_name": "Updated Branch Name",
    "address": "456 New Street",
    "phone_number": "+251911234568"
  }'
```

**Response:** `200 OK`
```json
{
  "data": {
    "id": "branch-id",
    "branch_name": "Updated Branch Name",
    "address": "456 New Street",
    "phone_number": "+251911234568"
  },
  "message": "Branch updated successfully",
  "status_code": 200
}
```

---

### 4. Delete Branch
**DELETE** `/branches/{id}`

**cURL:**
```bash
curl -X DELETE http://localhost:8080/branches/BRANCH_ID_HERE
```

**Response:** `200 OK`
```json
{
  "message": "Branch deleted successfully",
  "status_code": 200
}
```

---

## Menu Routes

### 1. Create Menu
**POST** `/menus`

**Request:** Multipart form data
- `name` (string): Menu name
- `image` (file): Menu image file

**cURL:**
```bash
curl -X POST http://localhost:8080/menus \
  -F "name=Breakfast Menu" \
  -F "image=@/path/to/menu-image.png"
```

**Response:** `200 OK`
```json
{
  "data": {
    "id": "menu-id",
    "name": "Breakfast Menu",
    "image": "image-url",
    "created_at": "2024-02-07T10:00:00Z"
  },
  "message": "Menu created successfully",
  "status_code": 200
}
```

---

### 2. Get Menu by ID
**GET** `/menus/{id}`

**cURL:**
```bash
curl -X GET http://localhost:8080/menus/MENU_ID_HERE
```

**Response:** `200 OK`
```json
{
  "data": {
    "id": "menu-id",
    "name": "Breakfast Menu",
    "image": "image-url",
    "created_at": "2024-02-07T10:00:00Z"
  },
  "message": "Menu fetched successfully",
  "status_code": 200
}
```

---

### 3. Update Menu
**PUT** `/menus/{id}`

**Request:** Multipart form data
- `name` (string): Menu name
- `image` (file, optional): Menu image file

**cURL:**
```bash
curl -X PUT http://localhost:8080/menus/MENU_ID_HERE \
  -F "name=Updated Menu Name" \
  -F "image=@/path/to/new-image.png"
```

**Response:** `200 OK`
```json
{
  "data": {
    "id": "menu-id",
    "name": "Updated Menu Name",
    "image": "new-image-url"
  },
  "message": "Menu updated successfully",
  "status_code": 200
}
```

---

### 4. Delete Menu
**DELETE** `/menus/{id}`

**cURL:**
```bash
curl -X DELETE http://localhost:8080/menus/MENU_ID_HERE
```

**Response:** `200 OK`
```json
{
  "message": "Menu deleted successfully",
  "status_code": 200
}
```

---

## Testing Workflow Example

### Step 1: Create a Merchant
```bash
MERCHANT_ID=$(curl -s -X POST http://localhost:8080/merchants \
  -F "name=Test Restaurant" \
  -F "logo=@/path/to/logo.png" | jq -r '.data.id')
echo "Merchant ID: $MERCHANT_ID"
```

### Step 2: Create a Branch
```bash
BRANCH_ID=$(curl -s -X POST http://localhost:8080/branches \
  -H "Content-Type: application/json" \
  -d "{
    \"merchant_id\": \"$MERCHANT_ID\",
    \"branch_name\": \"Test Branch\",
    \"address\": \"123 Test St\",
    \"phone_number\": \"+251911234567\"
  }" | jq -r '.data.id')
echo "Branch ID: $BRANCH_ID"
```

### Step 3: Create a User
```bash
USER_ID=$(curl -s -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d "{
    \"phone_number\": \"+251911234567\",
    \"full_name\": \"Test User\",
    \"branch_id\": \"$BRANCH_ID\",
    \"merchant_id\": \"$MERCHANT_ID\"
  }" | jq -r '.data.id')
echo "User ID: $USER_ID"
```

### Step 4: Set Password
```bash
curl -X POST http://localhost:8080/set-password \
  -H "Content-Type: application/json" \
  -d "{
    \"phone_number\": \"+251911234567\",
    \"password\": \"test123\"
  }"
```

### Step 5: Login
```bash
TOKEN=$(curl -s -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d "{
    \"phone_number\": \"+251911234567\",
    \"password\": \"test123\"
  }" | jq -r '.data.access_token')
echo "Access Token: $TOKEN"
```

---

## Using HTTPie (Alternative to cURL)

If you prefer HTTPie, here are some examples:

```bash
# Install HTTPie: pip install httpie

# Create User
http POST localhost:8080/users \
  phone_number="+251911234567" \
  full_name="John Doe" \
  branch_id="branch-uuid" \
  merchant_id="merchant-uuid"

# Login
http POST localhost:8080/login \
  phone_number="+251911234567" \
  password="test123"
```

---

## Notes

1. **Base URL**: All endpoints use `http://localhost:8080` (or your configured PORT)
2. **Content-Type**: JSON endpoints require `Content-Type: application/json`
3. **File Uploads**: Merchant and Menu endpoints use `multipart/form-data`
4. **IDs**: Replace `{id}` placeholders with actual UUIDs from previous responses
5. **Phone Numbers**: Use E.164 format (e.g., `+251911234567`)
6. **Error Responses**: All errors follow the standard error format defined in `internal/common/error.go`

---

## Quick Test Script

Save this as `test-api.sh`:

```bash
#!/bin/bash

BASE_URL="http://localhost:8080"

echo "Testing API endpoints..."

# Test health/connection
echo -e "\n1. Testing server connection..."
curl -s -o /dev/null -w "Status: %{http_code}\n" $BASE_URL/users || echo "Server not running!"

# Create merchant
echo -e "\n2. Creating merchant..."
MERCHANT_RESPONSE=$(curl -s -X POST $BASE_URL/merchants -F "name=Test Merchant")
echo $MERCHANT_RESPONSE | jq '.'

# Get all merchants
echo -e "\n3. Getting all merchants..."
curl -s $BASE_URL/merchants | jq '.'

echo -e "\n✅ API tests completed!"
```

Make it executable: `chmod +x test-api.sh` and run: `./test-api.sh`

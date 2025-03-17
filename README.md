# Student Attendance System

A simple API for tracking student attendance with location data.

## Running the Application

### Docker Setup

1. Clone the repository and navigate to the project directory

2. Make sure you have a `.env` file with your database credentials

3. Start the application using Docker Compose:

```bash
docker-compose up -d
```

This starts:

- PostgreSQL database container (`student_db`)
- Application backend container (`zkp_backend`)

The API will be available at `http://localhost:8080`

## API Endpoints

### 1. Register a Student

Register a new student with their name and location.

**Endpoint:** `POST /register`

**Sample Request:**

```json
{
  "name": "Ram",
  "latitude": 37.774,
  "longitude": -122.414
}
```

**Sample Response:**

```json
{
  "id": "1",
  "name": "Ram",
  "location": {
    "latitude": 37.774,
    "longitude": -122.414
  },
  "login_token": 595445,
  "created_at": "2025-03-17T12:00:35.680577Z",
  "present_count": 0,
  "absent_count": 0
}
```

### 2. Get Student by Token

Retrieve a student's information using their login token.

**Endpoint:** `GET /user/{token}`

**Sample Response:**

```json
{
  "id": "2",
  "name": "Mann",
  "location": {
    "latitude": 13.00010999783522,
    "longitude": 77.60661453008652
  },
  "login_token": 893152,
  "created_at": "2025-03-17T13:28:54.163546Z",
  "present_count": 0,
  "absent_count": 0
}
```

### 3. Update Attendance

Update a student's attendance status and last checked time.

**Endpoint:** `PUT /students/{token}/attendance`

**Sample Request:**

```json
{
  "last_checked": "2025-03-17T12:30:00Z",
  "is_present": true
}
```

**Sample Response:**

```json
{
  "id": "1",
  "name": "Ram",
  "location": {
    "latitude": 37.774,
    "longitude": -122.414
  },
  "login_token": 595445,
  "created_at": "2025-03-17T12:00:35.680577Z",
  "last_checked": "2025-03-17T12:30:00Z",
  "present_count": 1,
  "absent_count": 0
}
```

# API Documentation

## Timesheet API Documentation

The Timesheet API provides endpoints to manage timesheet entries. Below are the
available endpoints with examples of how to use them.

### Health Check

**GET /health**

Check the health status of the API and its database connection.

```bash
curl -X GET http://localhost:8080/health
```

Response:

```json
{
  "status": "healthy"
}
```

### Create Timesheet Entry

**POST /api/entry**

Create a new timesheet entry.

```bash
curl -X POST http://localhost:8080/api/entry \
  -H "Content-Type: application/json" \
  -d '{
    "date": "2023-11-15",
    "client_name": "Client A",
    "client_hours": 6,
    "vacation_hours": 0,
    "idle_hours": 1,
    "training_hours": 1
  }'
```

Response:

```json
{
  "message": "Timesheet entry created successfully"
}
```

### Get All Entries

**GET /api/entries**

Fetch all timesheet entries or filter by year and month.

```bash
# Get all entries
curl -X GET http://localhost:8080/api/entries

# Filter by year and month (January 2023)
curl -X GET "http://localhost:8080/api/entries?year=2023&month=1"
```

Response:

```json
{
  "entries": [
    {
      "Id": 1,
      "Date": "2023-11-15",
      "Client_name": "Client A",
      "Client_hours": 6,
      "Vacation_hours": 0,
      "Idle_hours": 1,
      "Training_hours": 1,
      "Total_hours": 8,
      "Notes": ""
    }
    // more entries...
  ]
}
```

### Get Entry by Date

**GET /api/entry/:date**

Fetch a timesheet entry by date.

```bash
curl -X GET http://localhost:8080/api/entry/2023-11-15
```

Response:

```json
{
  "Id": 1,
  "Date": "2023-11-15",
  "Client_name": "Client A",
  "Client_hours": 6,
  "Vacation_hours": 0,
  "Idle_hours": 1,
  "Training_hours": 1,
  "Total_hours": 8,
  "Notes": ""
}
```

### Update Entry by ID

**PUT /api/entry/:id**

Update specific fields of a timesheet entry by ID.

```bash
curl -X PUT http://localhost:8080/api/entry/1 \
  -H "Content-Type: application/json" \
  -d '{
    "client_hours": 7,
    "training_hours": 1
  }'
```

Response:

```json
{
  "message": "Timesheet entry updated successfully"
}
```

### Update Entry by Date

**PUT /api/entry/date/:date**

Update an entire timesheet entry by date.

```bash
curl -X PUT http://localhost:8080/api/entry/date/2023-11-15 \
  -H "Content-Type: application/json" \
  -d '{
    "client_name": "Client B",
    "client_hours": 8,
    "vacation_hours": 0,
    "idle_hours": 0,
    "training_hours": 0
  }'
```

Response:

```json
{
  "message": "Timesheet entry updated successfully"
}
```

### Delete Entry

**DELETE /api/entry/:id**

Delete a timesheet entry by ID.

```bash
curl -X DELETE http://localhost:8080/api/entry/1
```

Response:

```json
{
  "message": "Timesheet entry deleted successfully"
}
```

### API Info

**GET /api**

Get API information including the client's IP address.

```bash
curl -X GET http://localhost:8080/api
```

Response:

```json
{
  "response": {
    "ip": "127.0.0.1"
  }
}
```

## Error Responses

All endpoints return appropriate HTTP status codes:

- `400 Bad Request` - Invalid input data
- `404 Not Found` - Requested resource not found
- `500 Internal Server Error` - Server-side error

Example error response:

```json
{
  "error": "Entry not found for date: 2023-11-16"
}
```

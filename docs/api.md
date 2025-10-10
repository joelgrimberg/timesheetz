# Timesheetz API Documentation

This document provides comprehensive documentation for the Timesheetz REST API. All examples use `curl` and assume the API is running on `http://localhost:8080`.

## Table of Contents

- [Base URL](#base-url)
- [Health Check](#health-check)
- [Timesheet Endpoints](#timesheet-endpoints)
- [Training Budget Endpoints](#training-budget-endpoints)
- [Training Hours Endpoints](#training-hours-endpoints)
- [Vacation Hours Endpoints](#vacation-hours-endpoints)
- [Overview Endpoints](#overview-endpoints)
- [Utility Endpoints](#utility-endpoints)
- [Export Endpoints](#export-endpoints)
- [Error Responses](#error-responses)

## Base URL

```
http://localhost:8080
```

**Note:** The default port is 8080, but this can be configured in your `config.json` or via the `--port` flag.

---

## Health Check

### Check API Health

Check if the API server is running and healthy.

**Endpoint:** `GET /health`

**Example:**
```bash
curl http://localhost:8080/health
```

**Response:**
```json
{
  "status": "ok"
}
```

---

## Timesheet Endpoints

### Get All Timesheet Entries

Retrieve all timesheet entries from the database.

**Endpoint:** `GET /api/timesheet`

**Example:**
```bash
curl http://localhost:8080/api/timesheet
```

**Response:**
```json
[
  {
    "Id": 1,
    "Date": "2024-10-10",
    "Client_name": "Acme Corp",
    "Client_hours": 8,
    "Vacation_hours": 0,
    "Idle_hours": 0,
    "Training_hours": 0,
    "Total_hours": 8,
    "Sick_hours": 0,
    "Holiday_hours": 0
  },
  {
    "Id": 2,
    "Date": "2024-10-11",
    "Client_name": "Acme Corp",
    "Client_hours": 6,
    "Vacation_hours": 0,
    "Idle_hours": 1,
    "Training_hours": 2,
    "Total_hours": 9,
    "Sick_hours": 0,
    "Holiday_hours": 0
  }
]
```

### Create Timesheet Entry

Create a new timesheet entry.

**Endpoint:** `POST /api/timesheet`

**Example:**
```bash
curl -X POST http://localhost:8080/api/timesheet \
  -H "Content-Type: application/json" \
  -d '{
    "Date": "2024-10-12",
    "Client_name": "Acme Corp",
    "Client_hours": 7,
    "Vacation_hours": 0,
    "Idle_hours": 1,
    "Training_hours": 1,
    "Sick_hours": 0,
    "Holiday_hours": 0
  }'
```

**Request Body:**
```json
{
  "Date": "2024-10-12",
  "Client_name": "Acme Corp",
  "Client_hours": 7,
  "Vacation_hours": 0,
  "Idle_hours": 1,
  "Training_hours": 1,
  "Sick_hours": 0,
  "Holiday_hours": 0
}
```

**Response:**
```json
{
  "Id": 3,
  "Date": "2024-10-12",
  "Client_name": "Acme Corp",
  "Client_hours": 7,
  "Vacation_hours": 0,
  "Idle_hours": 1,
  "Training_hours": 1,
  "Total_hours": 9,
  "Sick_hours": 0,
  "Holiday_hours": 0
}
```

### Update Timesheet Entry

Update an existing timesheet entry by ID.

**Endpoint:** `PUT /api/timesheet/:id`

**Example:**
```bash
curl -X PUT http://localhost:8080/api/timesheet/3 \
  -H "Content-Type: application/json" \
  -d '{
    "Client_name": "New Client",
    "Client_hours": 8,
    "Vacation_hours": 0,
    "Idle_hours": 0,
    "Training_hours": 0,
    "Sick_hours": 0,
    "Holiday_hours": 0
  }'
```

**Request Body:**
```json
{
  "Client_name": "New Client",
  "Client_hours": 8,
  "Vacation_hours": 0,
  "Idle_hours": 0,
  "Training_hours": 0,
  "Sick_hours": 0,
  "Holiday_hours": 0
}
```

**Response:**
```json
{
  "Id": 3,
  "Date": "2024-10-12",
  "Client_name": "New Client",
  "Client_hours": 8,
  "Vacation_hours": 0,
  "Idle_hours": 0,
  "Training_hours": 0,
  "Total_hours": 8,
  "Sick_hours": 0,
  "Holiday_hours": 0
}
```

### Delete Timesheet Entry

Delete a timesheet entry by ID.

**Endpoint:** `DELETE /api/timesheet/:id`

**Example:**
```bash
curl -X DELETE http://localhost:8080/api/timesheet/3
```

**Response:**
```json
{
  "message": "Entry deleted successfully"
}
```

---

## Training Budget Endpoints

### Get Training Budget Entries

Retrieve all training budget entries for a specific year.

**Endpoint:** `GET /api/training-budget?year={year}`

**Parameters:**
- `year` (required): The year to retrieve entries for

**Example:**
```bash
curl "http://localhost:8080/api/training-budget?year=2024"
```

**Response:**
```json
[
  {
    "Id": 1,
    "Date": "2024-03-15",
    "Training_name": "Go Programming Workshop",
    "Hours": 8,
    "Cost_without_vat": 1500.00
  },
  {
    "Id": 2,
    "Date": "2024-06-20",
    "Training_name": "API Design Course",
    "Hours": 16,
    "Cost_without_vat": 2000.00
  }
]
```

### Create Training Budget Entry

Create a new training budget entry.

**Endpoint:** `POST /api/training-budget`

**Example:**
```bash
curl -X POST http://localhost:8080/api/training-budget \
  -H "Content-Type: application/json" \
  -d '{
    "Date": "2024-10-15",
    "Training_name": "Cloud Architecture Certification",
    "Hours": 40,
    "Cost_without_vat": 3500.00
  }'
```

**Request Body:**
```json
{
  "Date": "2024-10-15",
  "Training_name": "Cloud Architecture Certification",
  "Hours": 40,
  "Cost_without_vat": 3500.00
}
```

**Response:**
```json
{
  "Id": 3,
  "Date": "2024-10-15",
  "Training_name": "Cloud Architecture Certification",
  "Hours": 40,
  "Cost_without_vat": 3500.00
}
```

### Update Training Budget Entry

Update an existing training budget entry.

**Endpoint:** `PUT /api/training-budget`

**Example:**
```bash
curl -X PUT http://localhost:8080/api/training-budget \
  -H "Content-Type: application/json" \
  -d '{
    "Id": 3,
    "Date": "2024-10-15",
    "Training_name": "Updated Cloud Architecture Certification",
    "Hours": 48,
    "Cost_without_vat": 4000.00
  }'
```

**Request Body:**
```json
{
  "Id": 3,
  "Date": "2024-10-15",
  "Training_name": "Updated Cloud Architecture Certification",
  "Hours": 48,
  "Cost_without_vat": 4000.00
}
```

**Response:**
```json
{
  "Id": 3,
  "Date": "2024-10-15",
  "Training_name": "Updated Cloud Architecture Certification",
  "Hours": 48,
  "Cost_without_vat": 4000.00
}
```

### Delete Training Budget Entry

Delete a training budget entry by ID.

**Endpoint:** `DELETE /api/training-budget?id={id}`

**Parameters:**
- `id` (required): The ID of the entry to delete

**Example:**
```bash
curl -X DELETE "http://localhost:8080/api/training-budget?id=3"
```

**Response:**
```json
{
  "message": "Entry deleted successfully"
}
```

---

## Training Hours Endpoints

### Get Training Hours Summary

Get a summary of training hours for a specific year, including total, used, and available hours.

**Endpoint:** `GET /api/training-hours?year={year}`

**Parameters:**
- `year` (required): The year to get training hours for

**Example:**
```bash
curl "http://localhost:8080/api/training-hours?year=2024"
```

**Response:**
```json
{
  "year": 2024,
  "total_hours": 36,
  "used_hours": 18,
  "available_hours": 18
}
```

**Response Fields:**
- `year`: The year for which hours are calculated
- `total_hours`: Total training hours allocated per year (from config)
- `used_hours`: Training hours already used in timesheet entries
- `available_hours`: Remaining training hours (total - used)

---

## Vacation Hours Endpoints

### Get Vacation Hours Summary

Get a summary of vacation hours for a specific year, including total, used, and available hours.

**Endpoint:** `GET /api/vacation-hours?year={year}`

**Parameters:**
- `year` (required): The year to get vacation hours for

**Example:**
```bash
curl "http://localhost:8080/api/vacation-hours?year=2024"
```

**Response:**
```json
{
  "year": 2024,
  "total_hours": 180,
  "used_hours": 45,
  "available_hours": 135
}
```

**Response Fields:**
- `year`: The year for which hours are calculated
- `total_hours`: Total vacation hours allocated per year (from config)
- `used_hours`: Vacation hours already used in timesheet entries
- `available_hours`: Remaining vacation hours (total - used)

---

## Overview Endpoints

### Get Overview

Get a comprehensive overview of training and vacation days left for a specific year. This endpoint combines training and vacation data and calculates days remaining (based on 9 hours per day).

**Endpoint:** `GET /api/overview?year={year}`

**Parameters:**
- `year` (optional): The year to get overview data for. Defaults to current year if not specified.

**Example (current year):**
```bash
curl http://localhost:8080/api/overview
```

**Example (specific year):**
```bash
curl "http://localhost:8080/api/overview?year=2024"
```

**Response:**
```json
{
  "year": 2024,
  "training": {
    "total_hours": 36,
    "used_hours": 18,
    "available_hours": 18,
    "days_left": 2.0
  },
  "vacation": {
    "total_hours": 180,
    "used_hours": 90,
    "available_hours": 90,
    "days_left": 10.0
  }
}
```

**Response Fields:**
- `year`: The year for which the overview is calculated
- `training.total_hours`: Total training hours allocated per year (from config)
- `training.used_hours`: Training hours already used
- `training.available_hours`: Remaining training hours
- `training.days_left`: Remaining training days (calculated as available_hours / 9)
- `vacation.total_hours`: Total vacation hours allocated per year (from config)
- `vacation.used_hours`: Vacation hours already used
- `vacation.available_hours`: Remaining vacation hours
- `vacation.days_left`: Remaining vacation days (calculated as available_hours / 9)

---

## Utility Endpoints

### Get Last Client Name

Retrieve the client name from the most recent timesheet entry. Useful for pre-filling forms.

**Endpoint:** `GET /api/last-client`

**Example:**
```bash
curl http://localhost:8080/api/last-client
```

**Response:**
```json
{
  "client_name": "Acme Corp"
}
```

---

## Export Endpoints

### Export to PDF

Export timesheet data to PDF format.

**Endpoint:** `GET /api/export/pdf`

**Example:**
```bash
curl http://localhost:8080/api/export/pdf
```

**Response:**
```json
{
  "error": "PDF export not implemented yet"
}
```

**Status:** Not yet implemented

### Export to Excel

Export timesheet data to Excel format.

**Endpoint:** `GET /api/export/excel`

**Example:**
```bash
curl http://localhost:8080/api/export/excel
```

**Response:**
```json
{
  "error": "Excel export not implemented yet"
}
```

**Status:** Not yet implemented

---

## Error Responses

All endpoints return appropriate HTTP status codes and error messages:

### Status Codes

- `200 OK` - Successful request
- `201 Created` - Resource created successfully
- `400 Bad Request` - Invalid request parameters or body
- `404 Not Found` - Resource not found
- `500 Internal Server Error` - Server-side error

### Error Response Format

```json
{
  "error": "Error message describing what went wrong"
}
```

### Common Errors

**Missing Required Parameter:**
```bash
curl "http://localhost:8080/api/training-hours"
```
```json
{
  "error": "Year parameter is required"
}
```

**Invalid Parameter:**
```bash
curl "http://localhost:8080/api/training-hours?year=abc"
```
```json
{
  "error": "Invalid year parameter"
}
```

**Resource Not Found:**
```bash
curl -X DELETE http://localhost:8080/api/timesheet/999
```
```json
{
  "error": "no entry found with id 999"
}
```

**Invalid JSON Body:**
```bash
curl -X POST http://localhost:8080/api/timesheet \
  -H "Content-Type: application/json" \
  -d '{"invalid json'
```
```json
{
  "error": "invalid character 'i' in literal null (expecting 'u')"
}
```

---

## Testing the API

### Quick Test Sequence

Here's a sequence of commands to test the API:

```bash
# 1. Check if API is running
curl http://localhost:8080/health

# 2. Get all timesheet entries
curl http://localhost:8080/api/timesheet

# 3. Create a new entry
curl -X POST http://localhost:8080/api/timesheet \
  -H "Content-Type: application/json" \
  -d '{
    "Date": "2024-10-12",
    "Client_name": "Test Client",
    "Client_hours": 8,
    "Vacation_hours": 0,
    "Idle_hours": 0,
    "Training_hours": 0,
    "Sick_hours": 0,
    "Holiday_hours": 0
  }'

# 4. Get overview
curl http://localhost:8080/api/overview

# 5. Get training hours for current year
curl "http://localhost:8080/api/training-hours?year=2024"

# 6. Get vacation hours for current year
curl "http://localhost:8080/api/vacation-hours?year=2024"

# 7. Get last client name
curl http://localhost:8080/api/last-client
```

### Using jq for Pretty Output

Install `jq` and pipe responses for better formatting:

```bash
curl http://localhost:8080/api/overview | jq
```

---

## Notes

- All datetime fields use ISO 8601 format: `YYYY-MM-DD`
- The API automatically refreshes the TUI when data is modified via POST, PUT, or DELETE operations
- All numeric fields for hours are integers
- Cost fields in training budget are floating-point numbers with 2 decimal places
- The API uses SQLite as its database backend
- Configuration is loaded from `~/.config/timesheetz/config.json` (or equivalent on your OS)

For more information about keyboard shortcuts in the TUI, see [docs/shortcuts.md](shortcuts.md).

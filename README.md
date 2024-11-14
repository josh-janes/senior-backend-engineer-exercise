## Features

1. The API accepts job data in JSON format and updates the SQLite database
2. Uses transactions to ensure data consistency
3. Handles both insertions and updates (UPSERT) using SQLite's ON CONFLICT clause
4. Validates foreign key constraints against the employees table
5. Configurable port through environment variable
6. Dockerized server

## To run the server

The server is fully Dockerized, so all you need to do is build the image and run.

```
export PORT=8080

# Build the Docker image
docker build -t syndio-backend .

# Run the container
docker run -p $PORT:$PORT -v $(pwd)/employees.db:/app/employees.db syndio-backend
```

## Usage

To ingest or update job data, send a POST request to /api/jobs endpoint:

### POST /api/jobs

```
# Test the API endpoint
curl -X POST http://localhost:8080/api/jobs \
  -H "Content-Type: application/json" \
  -d '[
    {"employee_id": 1, "department": "Engineering", "job_title": "Senior Engineer"},
    {"employee_id": 2, "department": "Engineering", "job_title": "Super Senior Engineer"}
  ]'
```

To get a list of employees in the database:

### GET /api/employees
```
# List all employees and their job data
curl http://localhost:8080/api/employees
```


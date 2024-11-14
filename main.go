package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

// joined data schema
type EmployeeJobData struct {
	EmployeeID int     `json:"employee_id"`
	Gender     *string `json:"gender"`
	Department *string `json:"department"`
	JobTitle   *string `json:"job_title"`
}

var db *sql.DB
var port string = ""

func main() {
	port = os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port if not specified
	}

	// Open database connection
	var err error
	db, err = sql.Open("sqlite3", "employees.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create jobs table if it doesn't exist
	createTable := `
    CREATE TABLE IF NOT EXISTS jobs (
        employee_id INTEGER PRIMARY KEY,
        department TEXT NOT NULL,
        job_title TEXT NOT NULL,
        FOREIGN KEY(employee_id) REFERENCES employees(id)
    );`

	_, err = db.Exec(createTable)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/api/employees", http.HandlerFunc(getEmployees))
	http.HandleFunc("/api/jobs", http.HandlerFunc(updateJobs))

	log.Printf("Server starting")
	log.Fatal(http.ListenAndServe(":"+port, nil))
	log.Printf("Server runing on port %s...\n", port)
}

// List endpoint - GET /api/employees
func getEmployees(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Query to join employees and jobs tables
	query := `
        SELECT 
            e.id,
            e.gender,
            j.department,
            j.job_title
        FROM employees e
        LEFT JOIN jobs j ON e.id = j.employee_id
        ORDER BY e.id;
    `

	rows, err := db.Query(query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var employees []EmployeeJobData
	for rows.Next() {
		var emp EmployeeJobData
		var dept, title sql.NullString

		err := rows.Scan(&emp.EmployeeID, &emp.Gender, &dept, &title)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Handle NULL values from the LEFT JOIN
		if dept.Valid {
			emp.Department = &dept.String
		}
		if title.Valid {
			emp.JobTitle = &title.String
		}

		employees = append(employees, emp)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(employees)
}

// Update endpoint - POST /api/jobs
func updateJobs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var jobs []EmployeeJobData
	if err := json.NewDecoder(r.Body).Decode(&jobs); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Begin transaction
	tx, err := db.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Prepare statement for inserting/updating jobs
	stmt, err := tx.Prepare(`
        INSERT INTO jobs (employee_id, department, job_title) 
        VALUES (?, ?, ?)
        ON CONFLICT(employee_id) 
        DO UPDATE SET department=excluded.department, job_title=excluded.job_title
    `)
	if err != nil {
		tx.Rollback()
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	// Execute for each job
	for _, job := range jobs {
		_, err = stmt.Exec(job.EmployeeID, job.Department, job.JobTitle)
		if err != nil {
			tx.Rollback()
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

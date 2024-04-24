package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"io"
	"encoding/json"
	"encoding/csv"
	"github.com/joho/godotenv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/rs/cors"

	"database/sql"

	_ "github.com/lib/pq"
)

type Employee struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func main() {
	err := godotenv.Load()
	if err != nil {
	  log.Fatal("Error loading .env file")
	}

	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASS")
	dbname := os.Getenv("DB_DATABASE")

	db := connectPostgres(host, port, user, password, dbname)
	defer db.Close()

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("meow"))
	})

	r.Route("/employee", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			getEmployee(db, w, r)
		})
		r.Put("/update", func(w http.ResponseWriter, r *http.Request) {
			updateEmployee(db, w, r)
		})

		r.Put("/update-csv", func(w http.ResponseWriter, r *http.Request) {
			updateEmployeeFromCSV(db, w, r)
		})
	})

	c := cors.New(cors.Options{
		AllowedOrigins: []string{"http://localhost:5173"},
		AllowedMethods: []string{http.MethodGet},
	})

	handler := c.Handler(r)

	// Start server
	go func() {
		log.Println("Starting server on :3000")
		if err := http.ListenAndServe(":3000", handler); err != nil {
			log.Fatal(err)
		}
	}()

	// Read GET and pass CSV

	response, err := http.Get("http://localhost:3000/employee")

    if err != nil {
        fmt.Print(err.Error())
        os.Exit(1)
    }

    responseData, err := io.ReadAll(response.Body)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(string(responseData))

	csvData, err := marshalCSV(responseData)
    if err != nil {
        log.Fatal(err)
        return
    }

	// Write CSV data to file
	err = os.WriteFile("example2.csv", csvData, 0644)
	if err != nil {
		log.Fatal(err)
		return
	}

	fmt.Println("CSV file 'example2.csv' generated successfully")

	// Keep the main goroutine running
	select {}
}

func JSONMiddleWare(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func updateEmployeeFromCSV(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	file, err := os.Open("test_update.csv")
	if err != nil {
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	// Skip header
	if _, err := reader.Read(); err != nil {
		return
	}

	for {
		record, err := reader.Read()
		fmt.Println("record")
		fmt.Println(record)
		fmt.Println(len(record))
		if err == io.EOF {
			break
		}
		if err != nil {
			return
		}

		// Assuming CSV format: ID,Name,Email
		if len(record) != 3 {
			return
		}

		// Convert ID to int
		id, err := strconv.Atoi(record[0])
		if err != nil {
			return
		}
		fmt.Println("record array")
		fmt.Println(record[0])
		fmt.Println(record[1])
		// Update employee record in the database
		_, err = db.Exec("UPDATE employee SET name=?, email=?, WHERE id=?", record[1], record[2], record[0], id)
		if err != nil {
			return
		}
		fmt.Println(err)
	}

	return
}





func connectPostgres(host, port, user, password, dbname string) *sql.DB {
	db, err := sql.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname))

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Successfully connected to PostgreSQL!")

	return db
}

func getEmployee(db *sql.DB, w http.ResponseWriter, r *http.Request) {
    // Query the database to fetch employees
    rows, err := db.Query("SELECT id, name, email FROM employee")
    if err != nil {
        http.Error(w, "Failed to query database", http.StatusInternalServerError)
        log.Fatal(err)
        return
    }
    defer rows.Close()

    // Create a slice to store the retrieved employees
    var employees []Employee

    // Iterate over the rows
    for rows.Next() {
        var emp Employee
        // Scan the values from the row into the struct
        err := rows.Scan(&emp.ID, &emp.Name, &emp.Email)
        if err != nil {
            http.Error(w, "Failed to scan rows", http.StatusInternalServerError)
            log.Fatal(err)
            return
        }
        // Append the employee to the slice
        employees = append(employees, emp)
    }

    // Check for errors during iteration
    if err = rows.Err(); err != nil {
        http.Error(w, "Error during iteration", http.StatusInternalServerError)
        log.Fatal(err)
        return
    }

	json.NewEncoder(w).Encode(employees)

    // // Marshal employees slice into CSV format
    // csvData, err := marshalCSV(employees)
    // if err != nil {
    //     http.Error(w, "Failed to marshal CSV data", http.StatusInternalServerError)
    //     log.Fatal(err)
    //     return
    // }

	// // Write CSV data to file
	// err = os.WriteFile("example.csv", csvData, 0644)
	// if err != nil {
	// 	http.Error(w, "Failed to write CSV file", http.StatusInternalServerError)
	// 	log.Fatal(err)
	// 	return
	// }

    // // Respond with a success message
    // w.WriteHeader(http.StatusOK)
    // fmt.Fprintln(w, "CSV file 'example.csv' generated successfully")
}

func marshalCSV(jsonData []byte) ([]byte, error) {
    // Unmarshal JSON data into slice of Employee structs
    var employees []Employee
    if err := json.Unmarshal(jsonData, &employees); err != nil {
        return nil, err
    }

    // Initialize CSV buffer
    var csvData bytes.Buffer

    // Create CSV writer
    writer := csv.NewWriter(&csvData)

    // Write CSV header
    if err := writer.Write([]string{"ID", "Name", "Email"}); err != nil {
        return nil, err
    }

    // Write employee data to CSV
    for _, emp := range employees {
        if err := writer.Write([]string{strconv.Itoa(emp.ID), emp.Name, emp.Email}); err != nil {
            return nil, err
        }
    }

    // Flush the writer
    writer.Flush()

    // Check for errors
    if err := writer.Error(); err != nil {
        return nil, err
    }

    // Return CSV data as bytes
    return csvData.Bytes(), nil
}


func updateEmployee(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	var employee Employee
	err := json.NewDecoder(r.Body).Decode(&employee)
	if err != nil {
		http.Error(w, "Failed to decode", http.StatusBadRequest)
		return
	}

	// Check if the employee with the given ID exists
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM employee WHERE id = $1", employee.ID).Scan(&count)
	if err != nil {
		http.Error(w, "Failed to check if employee ID exists", http.StatusInternalServerError)
		log.Fatal(err)
		return
	}
	if count == 0 {
		http.Error(w, fmt.Sprintf("Employee with ID %d does not exist", employee.ID), http.StatusNotFound)
		return
	}

	_, err = db.Exec("UPDATE employee SET name = $1, email = $2 WHERE id = $3", employee.Name, employee.Email, employee.ID)
	if err != nil {
		http.Error(w, "Failed to update employee", http.StatusInternalServerError)
		log.Fatal(err)
		return
	}

	// Respond with a success message
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Employee with ID %d updated successfully", employee.ID)

}

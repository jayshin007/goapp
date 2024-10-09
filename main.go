package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

const (
	dbDriver = "mysql"
	dbUser   = "root"
	dbPass   = "root"
	dbName   = "tcp(localhost:3306)/testdb"
)

func GetConnector() *sql.DB {
	cfg := mysql.Config{
		User:                 "root",
		Passwd:               "root",
		Net:                  "tcp",
		Addr:                 "localhost:3306",
		Collation:            "utf8mb4_general_ci",
		Loc:                  time.UTC,
		MaxAllowedPacket:     4 << 20.,
		AllowNativePasswords: true,
		CheckConnLiveness:    true,
		DBName:               "testdb",
	}
	connector, err := mysql.NewConnector(&cfg)
	if err != nil {
		panic(err)
	}
	db := sql.OpenDB(connector)
	return db
}

func main() {

	// Create a new router
	r := mux.NewRouter()

	// Define your HTTP routes using the router
	r.HandleFunc("/user", createUserHandler).Methods("POST")
	r.HandleFunc("/user/{id}", getUserHandler).Methods("GET")
	r.HandleFunc("/user/{id}", updateUserHandler).Methods("PUT")
	r.HandleFunc("/user/{id}", deleteUserHandler).Methods("DELETE")
	r.HandleFunc("/", homeHandler)

	// Start the HTTP server on port 8090
	log.Println("Server listening on :8090")
	log.Fatal(http.ListenAndServe(":8090", r))
}

type User struct {
	ID    int    `json:"ID"`
	Name  string `json:"Name"`
	Email string `json:"Email"`
}

func createUserHandler(w http.ResponseWriter, r *http.Request) {
	db := GetConnector()
	err := db.Ping()
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// Parse JSON data from the request body
	var user User

	json.NewDecoder(r.Body).Decode(&user)

	fmt.Fprintln(w, user.Email)

	CreateUser(db, user.Name, user.Email)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintln(w, "User created successfully")
}

func CreateUser(db *sql.DB, name, email string) error {
	query := "INSERT INTO users (name, email) VALUES (?, ?)"
	_, err := db.Exec(query, name, email)
	if err != nil {
		return err
	}
	return nil
}

func getUserHandler(w http.ResponseWriter, r *http.Request) {
	db := GetConnector()
	err := db.Ping()
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// Get the 'id' parameter from the URL
	vars := mux.Vars(r)
	idStr := vars["id"]

	// Convert 'id' to an integer
	userID, err := strconv.Atoi(idStr)

	// Call the GetUser function to fetch the user data from the database
	user, err := GetUser(db, userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Convert the user object to JSON and send it in the response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func GetUser(db *sql.DB, id int) (*User, error) {
	query := "SELECT * FROM users WHERE id = ?"
	row := db.QueryRow(query, id)

	user := &User{}
	err := row.Scan(&user.ID, &user.Name, &user.Email)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func updateUserHandler(w http.ResponseWriter, r *http.Request) {
	db := GetConnector()
	err := db.Ping()
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// Get the 'id' parameter from the URL
	vars := mux.Vars(r)
	idStr := vars["id"]

	// Convert 'id' to an integer
	userID, err := strconv.Atoi(idStr)

	var user User
	err = json.NewDecoder(r.Body).Decode(&user)

	// Call the GetUser function to fetch the user data from the database
	UpdateUser(db, userID, user.Name, user.Email)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	fmt.Fprintln(w, "User updated successfully")
}

func UpdateUser(db *sql.DB, id int, name, email string) error {
	query := "UPDATE users SET name = ?, email = ? WHERE id = ?"
	_, err := db.Exec(query, name, email, id)
	if err != nil {
		return err
	}
	return nil
}

func deleteUserHandler(w http.ResponseWriter, r *http.Request) {
	db := GetConnector()
	err := db.Ping()
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// Get the 'id' parameter from the URL
	vars := mux.Vars(r)
	idStr := vars["id"]

	// Convert 'id' to an integer
	userID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid 'id' parameter", http.StatusBadRequest)
		return
	}

	user := DeleteUser(db, userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	fmt.Fprintln(w, "User deleted successfully")

	// Convert the user object to JSON and send it in the response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func DeleteUser(db *sql.DB, id int) error {
	query := "DELETE FROM users WHERE id = ?"
	_, err := db.Exec(query, id)
	if err != nil {
		return err
	}
	return nil
}

// 핸들러 함수 - HTML 페이지 응답
func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `
<!DOCTYPE html>
<html lang="en">
<head>
   <meta name="referrer" content="no-referrer-when-downgrade" />
   <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Submit User Info as JSON</title>
</head>
<body>
    <h2>Submit User Information</h2>

    <form id="postUserForm">
        <label for="name">Name:</label><br>
        <input type="text" id="name" name="name" required><br><br>

        <label for="email">Email:</label><br>
        <input type="email" id="email" name="email" required><br><br>

        <button type="submit">Submit</button>
    </form>

    <script>



        document.getElementById('postUserForm').addEventListener('submit', function(event) {
            event.preventDefault();

            // 폼 입력 값 가져오기
            const name = document.getElementById('name').value;
            const email = document.getElementById('email').value;

             // Fetch API를 사용해 JSON 형식으로 데이터 전송
            fetch('http://localhost:8090/user', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json' // Content-Type 설정
                },
                body: JSON.stringify({
                    Name: name,
                    Email: email
                })
            })
            .then(response => response.json());
        });
    </script>
</body>
</html>
	`)
}

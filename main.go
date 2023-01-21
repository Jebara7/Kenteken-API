package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

// JSON structure
type Kenteken struct {
	Kenteken string `json:"Kenteken"`
}

func init() {
	logFile, err := os.OpenFile(".\\trace.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal("Logfile kan niet aangemaakt worden")
	}
	log.SetOutput(logFile)
}

// Read JSON configuration file
func ReadJSON(fileName string) (map[string]string, error) {
	configFile, _ := ioutil.ReadFile(fileName)

	// Unmarshal JSON data
	var config map[string]string
	err := json.Unmarshal(configFile, &config)
	if err != nil {
		log.Fatal(err)
	}
	return config, err
}

// Get environment variables from JSON
func GetEnvVars(config map[string]string) (host string, user string, password string, dbname string, port string) {
	host = config["MYSQL_HOST"]
	user = config["MYSQL_USER"]
	password = config["MYSQL_PASSWORD"]
	dbname = config["MYSQL_DATABASE"]
	port = config["MYSQL_PORT"]
	return host, user, password, dbname, port
}

// Create connection string for MySQL
func ConnectToDB(user string, password string, host string, port string, dbname string) (db *sql.DB, err error) {
	db, err = sql.Open("mysql", user+":"+password+"@tcp("+host+":"+port+")/"+dbname)
	if err != nil {
		log.Fatal(err)
	}
	return db, err
}

func main() {
	fmt.Println("API Running...")
	fmt.Println("Connecting to database...")
	fmt.Println("Reading JSON!")

	// Read JSON configuration file
	config, err := ReadJSON("config.json")
	if err != nil {
		log.Fatal(err)
	}

	// Get environment variables from JSON
	host, user, password, dbname, port := GetEnvVars(config)

	// Create connection string for MySQL
	db, err := ConnectToDB(user, password, host, port, dbname)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("JSON File Read & Connected!")
	http.HandleFunc("/storeKenteken", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "Error reading request body", http.StatusInternalServerError)
				return
			}
			kenteken := Kenteken{}
			err = json.Unmarshal(body, &kenteken)
			if err != nil {
				http.Error(w, "Error Unmarshalling JSON", http.StatusInternalServerError)
				return
			}

			// Insert kenteken into database
			err = InsertKenteken(db, kenteken.Kenteken)
			if err != nil {
				http.Error(w, "Error inserting Kenteken into database", http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusCreated)
			return
		}
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	})
	http.ListenAndServe(":8080", nil)
}

// Check to see if kenteken is already in database
func InsertKenteken(db *sql.DB, kenteken string) error {
	// Check to see if kenteken is already in database
	rows, err := db.Query("SELECT * FROM kenteken WHERE Kentekennummer = ?", kenteken)
	if err != nil {
		log.Println("Error checking duplicate kenteken")
		return err
	}
	defer rows.Close()

	// Kenteken is already in database
	if rows.Next() {
		log.Println("Kenteken already in database", http.StatusConflict)
		return err
	}

	// Insert kenteken into database
	_, err = db.Query("INSERT INTO kenteken (Kentekennummer) VALUES (?)", kenteken)
	if err != nil {
		log.Println("Error inserting Kenteken into database")
		return err
	}
	return err
}

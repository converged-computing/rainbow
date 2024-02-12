package server

import (
	"fmt"

	pb "github.com/converged-computing/rainbow/pkg/api/v1"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"

	"database/sql"
	"log"
	"os"
)

type Database struct {
	filepath string
}

// cleanup removes the filepath
func (db *Database) cleanup() {
	// Delete a previous database that exists
	// Note that in the future we might not want to do this
	log.Printf("🧹️ cleaning up %s...", db.filepath)
	os.Remove(db.filepath)
}

// create the database
func (db *Database) create() error {

	log.Printf("✨️ creating %s...", db.filepath)

	// Create SQLite file (ensures that we can!)
	file, err := os.Create(db.filepath)
	if err != nil {
		return err
	}
	file.Close()
	log.Printf("   %s file created", db.filepath)

	// Open the created SQLite File (to test)
	conn, err := db.connect()
	defer conn.Close()
	return err
}

// Connect to the database - the caller is responsible for closing
func (db *Database) connect() (*sql.DB, error) {
	conn, err := sql.Open("sqlite3", db.filepath)
	if err != nil {
		return nil, err
	}
	return conn, err
}

// RegisterCluster registers a cluster or returns another status
// REGISTER_SUCCESS = 1;
// REGISTER_ERROR = 2;
// REGISTER_DENIED = 3;
// REGISTER_EXISTS = 4;
func (db *Database) RegisterCluster(name string) (pb.RegisterResponse_ResultType, string, error) {

	// Connect!
	conn, err := db.connect()
	if err != nil {
		return 0, "", err
	}
	defer conn.Close()

	// First determine if it exists
	query := fmt.Sprintf("SELECT count(*) from clusters WHERE name LIKE \"%s\"", name)
	result, err := conn.Exec(query)
	count, err := result.RowsAffected()

	// Debugging extra for now
	fmt.Printf("%s: (%d)\n", query, count)

	// Case 1: already exists
	if count > 0 {
		return 4, "", nil
	}

	// Generate a "secret" token, lol
	token := uuid.New().String()
	query = fmt.Sprintf("INSERT into clusters VALUES (\"%s\", \"%s\")", name, token)
	result, err = conn.Exec(query)
	if err != nil {
		return 2, "", err
	}
	count, err = result.RowsAffected()
	fmt.Printf("%s: (%d)\n", query, count)

	// REGISTER_SUCCESS
	if count > 0 {
		return 1, token, nil
	}

	// REGISTER_ERROR
	return 2, "", err
}

// create the database
func (db *Database) createTables() error {

	conn, err := db.connect()
	if err != nil {
		return err
	}
	defer conn.Close()

	// Create the clusters table, where we store the name and secret
	// obviously the secret should not be stored in plain text - it's fine for now
	createClusterTableSQL := `
	CREATE TABLE clusters (
		name TEXT NOT NULL PRIMARY KEY,		
		secret TEXT
	  );`

	createJobsTableSQL := `
	  CREATE TABLE jobs (
		  idJob integer NOT NULL PRIMARY KEY AUTOINCREMENT,		
		  jobspec TEXT,
		  cluster TEXT,
		  FOREIGN KEY(cluster) REFERENCES clusters(name)
		);`

	for table, statement := range map[string]string{"cluster": createClusterTableSQL, "jobs": createJobsTableSQL} {
		log.Printf("   create %s table...\n", table)
		query, err := conn.Prepare(statement) // Prepare SQL Statement
		if err != nil {
			return err
		}
		// Execute SQL query
		_, err = query.Exec()
		if err != nil {
			return err
		}
		log.Printf("   %s table created\n", table)
	}
	return nil
}

func initDatabase(filepath string, cleanup bool) (*Database, error) {

	// Create a new database (todo, add cleanupc check)
	db := Database{filepath: filepath}

	if cleanup {
		db.cleanup()
	}

	// Create the database
	err := db.create()
	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}
	err = db.createTables()
	return &db, err
}

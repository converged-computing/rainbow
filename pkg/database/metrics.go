package database

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"
)

// WithTime calls a function interface and records a time
// Since we are using it for different scheduling functions, a function name is required.
// We aren't using this because the reflect.Value return needs to be typed!
// This would work nicely for a function that does not need output checked, etc.
func (d *Database) WithTime(fn interface{}, fnName string, params ...interface{}) (result []reflect.Value) {
	f := reflect.ValueOf(fn)
	if f.Type().NumIn() != len(params) {
		panic("incorrect number of parameters!")
	}
	inputs := make([]reflect.Value, len(params))
	for k, in := range params {
		inputs[k] = reflect.ValueOf(in)
	}

	start := time.Now()
	response := f.Call(inputs)
	duration := time.Since(start)

	// Save duration with function name in database, and return response
	fmt.Printf("Time for %s took %s", fnName, duration)
	d.SaveMetric(fnName, duration.String(), map[string]string{})
	return response
}

// SaveMetric saves a named metric for a job, optionally with metadata
// the calling function is free to define metadata as they please (format)
func (db *Database) SaveMetric(name, value string, metadata map[string]string) error {

	conn, err := db.connect()
	if err != nil {
		return err
	}
	defer conn.Close()

	fmt.Printf("metadata %s\n", metadata)
	var fields, values string
	if len(metadata) == 0 {
		fields = "(name, value)"
		values = fmt.Sprintf("(\"%s\", \"%s\")", name, value)
	} else {

		// Here we serialize the metadata to bytes
		meta, err := json.Marshal(metadata)
		if err != nil {
			return err
		}
		fields = "(name, value, metadata)"
		values = fmt.Sprintf("(\"%s\", \"%s\", '%s')", name, value, string(meta))
	}
	query := fmt.Sprintf("INSERT into metrics %s VALUES %s", fields, values)
	fmt.Println(query)

	// Execute SQL query
	_, err = conn.Exec(query)
	if err != nil {
		return err
	}
	return nil
}

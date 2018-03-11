package schema

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
)

// Server :
type Server struct {
	DriverName string `json:"driver_name"`
	Host       string `json:"host"`
	Port       int    `json:"port"`
	User       string `json:"user"`
	Password   string `json:"password"`
}

// Table :
type Table struct {
	Name       string   `json:"name"`
	PrimaryKey string   `json:"primary_key"`
	UniqueKey  []string `json:"unique_key"`
}

// Column :
type Column struct {
	Name          string `json:"name"`
	Type          string `json:"type"`
	NotNull       bool   `json:"not_null"`
	AutoIncrement bool   `json:"auto_increment"`
}

// Schema :
type Schema struct {
	DatabaseName string   `json:"database"`
	Description  string   `json:"description"`
	Table        Table    `json:"table"`
	Columns      []Column `json:"properties"`
}

// NewJSONSchema :
func NewJSONSchema(r io.Reader) (*Schema, error) {
	var sc *Schema
	if err := json.NewDecoder(r).Decode(&sc); err != nil {
		return nil, err
	}
	return sc, nil
}

// NewJSONFileSchema :
func NewJSONFileSchema(path string) (*Schema, error) {
	r, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return NewJSONSchema(r)
}

// Output :
func (sc *Schema) Output(path string) error {
	// Set default path and default file name.
	if path == "" {
		path = "resources/schema/" + sc.DatabaseName + "_" + sc.Table.Name + ".json"
	}

	// Output with indentation
	jsonBytes, err := json.MarshalIndent(sc, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, jsonBytes, 0644)
}

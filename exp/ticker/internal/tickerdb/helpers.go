package tickerdb

import (
	"fmt"
	"reflect"
)

// getDBFieldTags returns all "db" tags for a given struct, optionally excluding the "id".
func getDBFieldTags(model interface{}, excludeID bool) (fields []string) {
	r := reflect.ValueOf(model)
	for i := 0; i < r.Type().NumField(); i++ {
		dbField := r.Type().Field(i).Tag.Get("db")
		if excludeID && dbField == "id" {
			continue
		}
		fields = append(fields, dbField)
	}
	fields = sanitizeFieldNames(fields)
	return
}

// sanitizeFieldNames adds double quotes to each entry on a slice of field names.
func sanitizeFieldNames(fieldNames []string) (sanitizedFields []string) {
	for _, v := range fieldNames {
		quotedField := fmt.Sprintf("\"%s\"", v)
		sanitizedFields = append(sanitizedFields, quotedField)
	}
	return
}

// getDBFieldValues returns all "db"-tagged values, optionally excluding the "id".
func getDBFieldValues(model interface{}, excludeID bool) (values []interface{}) {
	r := reflect.ValueOf(model)
	for i := 0; i < r.Type().NumField(); i++ {
		dbField := r.Type().Field(i).Tag.Get("db")
		dbVal := r.Field(i).Interface()
		if excludeID && dbField == "id" {
			continue
		}
		values = append(values, dbVal)
	}
	return
}

// createOnConflictFragment generates a ON CONFLICT sql clause for a given constraint,
// preserving the fields listed in the fields param.
func createOnConflictFragment(constraint string, fields []string) (fragment string) {
	fragment = fmt.Sprintf("ON CONFLICT ON CONSTRAINT %s DO UPDATE SET ", constraint)
	for i, field := range fields {
		fragment += fmt.Sprintf("%s = EXCLUDED.%s", field, field)

		if i != len(fields)-1 {
			fragment += ","
		}
	}

	return
}

// generatePlaceholders generates a string formatted as (?, ?, ?, ?) of length
// equal to the size of the fields param
func generatePlaceholders(fields []interface{}) (placeholder string) {
	for i := range fields {
		placeholder += "?"

		if i != len(fields)-1 {
			placeholder += ", "
		}
	}
	return
}

// optionalVar is a simple struct to represent a query variable that should
// only be used in a statement if its value is not null
type optionalVar struct {
	name string
	val  *string
}

// generateWhereClause generates a WHERE clause in the format:
// "WHERE x = ? AND y = ? AND ..." where the number of conditions is equal
// to the number of optVars with val != nil. It also returns the valid vals
// in the args param. This function was created to take advantage of go/sql's
// sanitization and to prevent possible SQL injections.
func generateWhereClause(optVars []optionalVar) (clause string, args []string) {
	for _, ov := range optVars {
		if ov.val != nil {
			if clause == "" {
				clause = fmt.Sprintf("WHERE %s = ?", ov.name)
			} else {
				clause += fmt.Sprintf(" AND %s = ?", ov.name)
			}
			args = append(args, *ov.val)
		}
	}
	return
}

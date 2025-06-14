package db

import (
	"context"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Record interface {
	Info() (table string, fields []FieldI, updated map[string]struct{})
}

type FieldI interface {
	Name() string
	Value() any
	Pointer() any
}

type fieldT[T any] struct {
	name       string
	value      *T
	primaryKey bool
}

// Returns name of the field
func (f *fieldT[T]) Name() string {
	return f.name
}

// Returns value of the field
func (f *fieldT[T]) Value() any {
	return *f.value
}

// Returns pointer to the field
func (f *fieldT[T]) Pointer() any {
	return f.value
}

// Use as element in []FieldI
func Field[T any](name string, value *T) *fieldT[T] {
	return &fieldT[T]{name, value, false}
}

// Inserts the given record in the specified table.
func Insert(ctx context.Context, r Record) error {
	tbl, fields, _ := r.Info()
	fieldL := len(fields)
	if fieldL == 0 {
		return e("no fields for %T", r)
	}

	// build up cols, placeholders and args
	columns := make([]string, 0, fieldL)
	placeHolders := Placeholders(fieldL)
	args := make([]any, 0, fieldL)
	for _, field := range fields {
		columns = append(columns, field.Name())
		args = append(args, field.Value())
	}

	// exec query
	query := f("insert into %s (%s) values (%s)",
		tbl, strings.Join(columns, ", "), placeHolders)
	if _, err := Exec(ctx, query, args...); err != nil {
		if status.Code(err) == codes.AlreadyExists {
			return alreadyExist(r)
		}
		return e("inserting %s: %w", recordIdentifier(r), err)
	}
	return nil
}

// Gets the given record from the specified table. Only uses the primary key
// fields to build the needed query.
func Get(ctx context.Context, r Record) error {
	tbl, fields, _ := r.Info()
	if len(fields) == 0 {
		return e("getting %T: empty field map", r)
	}
	fieldL := len(fields)

	// build up cols, condition, args and destinations
	cols := make([]string, 0, fieldL)
	conditions := make([]string, 0, fieldL)
	args := make([]any, 0, fieldL)
	destinations := make([]any, 0, len(fields))
	for _, field := range fields {
		if strings.HasSuffix(field.Name(), "_pk") {
			conditions = append(conditions, field.Name()+" = ?")
			args = append(args, field.Value())
		} else {
			cols = append(cols, field.Name())
			destinations = append(destinations, field.Pointer())
		}
	}
	if len(conditions) == 0 {
		return e("getting %T: no primary key columns found", r)
	}

	// if no non-pk cols, use first pk col
	if len(cols) == 0 {
		cols = append(cols, fields[0].Name())
		destinations = append(destinations, fields[0].Value())
	}
	condition := strings.Join(conditions, " and ")

	// exec query
	query := f("select %s from %s where %s",
		strings.Join(cols, ","), tbl, condition)
	rows, err := Query(ctx, query, args...)
	if err != nil {
		return e("getting %T: %w", r, err)
	}

	// process rows
	defer rows.Close()
	if !rows.Next() {
		return notFound("getting", r)
	}
	if err := rows.Scan(destinations...); err != nil {
		return e("getting %s: %w", recordIdentifier(r), err)
	}
	return nil
}

// Updates the given record in the specified table. At least one fields must be
// specified, but more than one may be specified.
func Update(ctx context.Context, r Record) error {
	tbl, fields, updated := r.Info()
	if len(fields) == 0 {
		return e("updating %T: empty col map", r)
	}
	fieldL := len(fields)
	updatedL := len(updated)

	// build up conditions
	conditions := make([]string, 0, fieldL)
	conditionArgs := make([]any, 0, fieldL)
	for _, field := range fields {
		if strings.HasSuffix(field.Name(), "_pk") {
			conditions = append(conditions, field.Name()+" = ?")
			conditionArgs = append(conditionArgs, field.Value())
		}
	}
	if len(conditions) == 0 {
		return e("updating %T: no primary key columns found", r)
	}
	condition := strings.Join(conditions, " and ")

	// build up set clauses and args
	setClauses := make([]string, 0, updatedL)
	setArgs := make([]any, 0, updatedL)
	for _, field := range fields {
		if _, ok := updated[field.Name()]; ok {
			setClauses = append(setClauses, field.Name()+" = ?")
			setArgs = append(setArgs, field.Value())
		}
	}
	if len(setClauses) == 0 {
		return e("updating %s: no fields to update", recordIdentifier(r))
	}

	// exec query
	query := f("update %s set %s where %s",
		tbl, strings.Join(setClauses, ", "), condition)
	setArgs = append(setArgs, conditionArgs...)
	result, err := Exec(ctx, query, setArgs...)
	if err != nil {
		return e("updating %s: %w", recordIdentifier(r), err)
	}
	if affectedRows, err := result.RowsAffected(); err != nil {
		return e("updating %s: %w", recordIdentifier(r), err)
	} else if affectedRows == 0 {
		return notFound("updating", r)
	} else if affectedRows > 1 {
		return e("updating %s: expected 1 row affected, got %d", recordIdentifier(r), affectedRows)
	}

	// clear updated set
	for k := range updated {
		delete(updated, k)
	}
	return nil
}

// Deletes the given record from the specified table. Only uses the primary key
// columns (ending with "_pk") to build the query.
func Delete(ctx context.Context, r Record) error {
	tbl, fields, _ := r.Info()
	fieldL := len(fields)
	conditions := make([]string, 0, fieldL)
	args := make([]any, 0, fieldL)
	for _, field := range fields {
		if strings.HasSuffix(field.Name(), "_pk") {
			conditions = append(conditions, field.Name()+" = ?")
			args = append(args, field.Value())
		}
	}
	if len(conditions) == 0 {
		return e("deleting %T: no primary keys found", r)
	}
	condition := strings.Join(conditions, "and")
	query := f("delete from %s where %s", tbl, condition)
	if result, err := Exec(ctx, query, args...); err != nil {
		return e("deleting %T: %w", r, err)
	} else {
		if rowsAffected, err := result.RowsAffected(); err != nil {
			return e("deleting %s: %w", recordIdentifier(r), err)
		} else if rowsAffected == 0 {
			return notFound("deleting", r)
		} else if rowsAffected > 1 {
			return e("deleting %s: expected 1 row affected, got %d", recordIdentifier(r), rowsAffected)
		}
	}
	return nil
}

type NotFoundError struct {
	err string
}

func (e *NotFoundError) Error() string {
	return e.err
}

func notFound(action string, r Record) *NotFoundError {
	return &NotFoundError{
		err: f("%s %s: not found", action, recordIdentifier(r)),
	}
}

type AlreadyExistsError struct {
	err string
}

func (e *AlreadyExistsError) Error() string {
	return e.err
}

func alreadyExist(r Record) *AlreadyExistsError {
	return &AlreadyExistsError{
		err: f("inserting %s: already exists", recordIdentifier(r)),
	}
}

func recordIdentifier(r Record) string {
	_, fields, _ := r.Info()
	keys := make([]string, 0, len(fields))
	for _, field := range fields {
		if strings.HasSuffix(field.Name(), "_pk") {
			keys = append(keys, f("%+v", field.Value()))
		}
	}
	return f("%T(%s)", r, strings.Join(keys, ", "))
}

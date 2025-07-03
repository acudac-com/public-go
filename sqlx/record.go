package sqlx

import (
	"context"
	"fmt"
	"slices"
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

type field[T any] struct {
	name       string
	value      *T
	primaryKey bool
}

// Returns name of the field
func (f *field[T]) Name() string {
	return f.name
}

// Returns value of the field
func (f *field[T]) Value() any {
	return *f.value
}

// Returns pointer to the field
func (f *field[T]) Pointer() any {
	return f.value
}

// Use as element in []FieldI
func Field[T any](name string, value *T) *field[T] {
	return &field[T]{name, value, false}
}

// Inserts the given record in the specified table.
func (d *Db) Insert(ctx context.Context, r Record) error {
	tbl, fields, _ := r.Info()
	fieldL := len(fields)
	if fieldL == 0 {
		panic(fmt.Errorf("sqlx.Db.Insert(): no fields for %T", r))
	}

	// build up cols and args
	columns := make([]string, 0, fieldL)
	args := make([]any, 0, fieldL)
	for _, field := range fields {
		columns = append(columns, field.Name())
		args = append(args, field.Value())
	}

	// exec query
	query := fmt.Sprintf("insert into %s (%s) values (%s)",
		tbl, strings.Join(columns, ", "), Placeholders(fieldL))
	if _, err := d.Exec(ctx, query, args...); err != nil {
		if status.Code(err) == codes.AlreadyExists {
			return alreadyExist(r)
		}
		return fmt.Errorf("inserting %s: %w", recordIdentifier(r), err)
	}
	return nil
}

// Gets the given record from the specified table. Only uses the primary key
// fields to build the needed query.
func (d *Db) Get(ctx context.Context, r Record, columns ...string) error {
	tbl, fields, _ := r.Info()
	fieldL := len(fields)
	if fieldL == 0 {
		panic(fmt.Errorf("sqlx.Db.Get(): no fields for %T", r))
	}

	// build up conditions, args, cols & destinations
	conditions := make([]string, 0, fieldL)
	args := make([]any, 0, fieldL)
	cols := make([]string, 0, fieldL)
	destinations := make([]any, 0, fieldL)
	for _, field := range fields {
		if strings.HasSuffix(field.Name(), "_pk") {
			conditions = append(conditions, field.Name()+" = ?")
			args = append(args, field.Value())
		} else {
			if columns != nil && !slices.Contains(columns, field.Name()) {
				continue
			}
			cols = append(cols, field.Name())
			destinations = append(destinations, field.Pointer())
		}
	}
	if len(conditions) == 0 {
		panic(fmt.Errorf("sqlx.Db.Get(): no primary key columns for %T", r))
	}

	// if no non-pk cols, use first pk col
	if len(cols) == 0 {
		cols = append(cols, fields[0].Name())
		destinations = append(destinations, fields[0].Pointer())
	}
	condition := strings.Join(conditions, " and ")

	// exec query
	query := fmt.Sprintf("select %s from %s where %s",
		strings.Join(cols, ","), tbl, condition)
	rows, err := d.Query(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("getting %T: %w", r, err)
	}

	// process rows
	defer rows.Close()
	if !rows.Next() {
		return notFound("getting", r)
	}
	if err := rows.Scan(destinations...); err != nil {
		return fmt.Errorf("getting %s: %w", recordIdentifier(r), err)
	}
	return nil
}

// Updates the given record in the specified table. At least one fields must be
// specified, but more than one may be specified.
func (d *Db) Update(ctx context.Context, r Record) error {
	tbl, fields, updated := r.Info()
	fieldL := len(fields)
	updatedL := len(updated)
	if fieldL == 0 {
		panic(fmt.Errorf("sqlx.Db.Update(): no fields for %T", r))
	}
	if updatedL == 0 {
		panic(fmt.Errorf("sqlx.Db.Update(): no fields to update for %T", r))
	}

	// build up conditions and set clauses
	conditions := make([]string, 0, fieldL)
	conditionArgs := make([]any, 0, fieldL)
	setClauses := make([]string, 0, updatedL)
	setArgs := make([]any, 0, updatedL)
	for _, field := range fields {
		if strings.HasSuffix(field.Name(), "_pk") {
			conditions = append(conditions, field.Name()+" = ?")
			conditionArgs = append(conditionArgs, field.Value())
		} else if _, ok := updated[field.Name()]; ok {
			setClauses = append(setClauses, field.Name()+" = ?")
			setArgs = append(setArgs, field.Value())
		}
	}
	if len(conditions) == 0 {
		panic(fmt.Errorf("sqlx.Db.Update(): no primary key columns for %T", r))
	}
	if len(setClauses) == 0 {
		return fmt.Errorf("updating %s: no fields to update", recordIdentifier(r))
	}

	// exec query
	query := fmt.Sprintf("update %s set %s where %s",
		tbl, strings.Join(setClauses, ", "), strings.Join(conditions, " and "))
	setArgs = append(setArgs, conditionArgs...)
	result, err := d.Exec(ctx, query, setArgs...)
	if err != nil {
		return fmt.Errorf("updating %s: %w", recordIdentifier(r), err)
	}

	// process result
	if affectedRows, err := result.RowsAffected(); err != nil {
		return fmt.Errorf("updating %s: %w", recordIdentifier(r), err)
	} else if affectedRows == 0 {
		return notFound("updating", r)
	} else if affectedRows > 1 {
		panic(fmt.Errorf("sqlx.Db.Update(): more than one (%d) row affected for %s", affectedRows, recordIdentifier(r)))
	}

	// clear updated set
	for k := range updated {
		delete(updated, k)
	}
	return nil
}

// Deletes the given record from the specified table. Only uses the primary key
// columns (ending with "_pk") to build the query.
func (d *Db) Delete(ctx context.Context, r Record) error {
	tbl, fields, _ := r.Info()
	fieldL := len(fields)
	if fieldL == 0 {
		panic(fmt.Errorf("sqlx.Db.Delete(): no fields for %T", r))
	}

	// build up conditions and args
	conditions := make([]string, 0, fieldL)
	args := make([]any, 0, fieldL)
	for _, field := range fields {
		if strings.HasSuffix(field.Name(), "_pk") {
			conditions = append(conditions, field.Name()+" = ?")
			args = append(args, field.Value())
		}
	}
	if len(conditions) == 0 {
		panic(fmt.Errorf("sqlx.Db.Delete(): no primary keys columns found for %T", r))
	}

	// execute query
	query := fmt.Sprintf("delete from %s where %s",
		tbl, strings.Join(conditions, " and "))
	result, err := d.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("deleting %T: %w", r, err)
	}

	// process result
	if rowsAffected, err := result.RowsAffected(); err != nil {
		return fmt.Errorf("deleting %s: %w", recordIdentifier(r), err)
	} else if rowsAffected == 0 {
		return notFound("deleting", r)
	} else if rowsAffected > 1 {
		panic(fmt.Errorf("sqlx.Db.Delete(): more than one (%d) row affected for %s", rowsAffected, recordIdentifier(r)))
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
		err: fmt.Sprintf("%s %s: not found", action, recordIdentifier(r)),
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
		err: fmt.Sprintf("inserting %s: already exists", recordIdentifier(r)),
	}
}

func recordIdentifier(r Record) string {
	_, fields, _ := r.Info()
	keys := make([]string, 0, len(fields))
	for _, field := range fields {
		if strings.HasSuffix(field.Name(), "_pk") {
			keys = append(keys, fmt.Sprintf("%+v", field.Value()))
		}
	}
	return fmt.Sprintf("%T(%s)", r, strings.Join(keys, ", "))
}

// Returns a string with a number of placeholders (?) for SQL queries.
func Placeholders(nr int) string {
	placeholders := make([]string, 0, nr)
	for range nr {
		placeholders = append(placeholders, "?")
	}
	return strings.Join(placeholders, ", ")
}

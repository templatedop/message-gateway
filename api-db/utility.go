package db

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	apierrors "MgApplication/api-errors"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stephenafamo/bob"
	//l "MgApplication/api-log"
)

type TimedBatch struct {
	*pgx.Batch
	timeoutSet bool
	timeoutVal int
	//queries    *BatchQueries
}

func NewTimedBatch(timeoutMs int) *TimedBatch {
	return &TimedBatch{
		Batch:      &pgx.Batch{},
		timeoutSet: false,
		timeoutVal: timeoutMs,
		//queries:    &BatchQueries{Queries: make([]string, 0)},
	}
}

// func (b *TimedBatch) addQuery(query string) {
// 	b.queries.Lock()
// 	defer b.queries.Unlock()
// 	b.queries.Queries = append(b.queries.Queries, query)
// }

// // Get all queries in the batch
// func (b *TimedBatch) GetQueries() []string {
// 	b.queries.Lock()
// 	defer b.queries.Unlock()
// 	return append([]string{}, b.queries.Queries...)
// }

type BatchQueries struct {
	Queries []string
	sync.Mutex
}

func interpolateQuery(query string, args []interface{}) string {
	for i, arg := range args {
		var value string
		switch v := arg.(type) {
		case string:
			value = fmt.Sprintf("'%s'", v)
		case time.Time:
			value = fmt.Sprintf("'%s'", v.Format(time.RFC3339))
		default:
			value = fmt.Sprintf("%v", v)
		}
		// Replace $1, $2 etc with actual values
		query = strings.Replace(query, fmt.Sprintf("$%d", i+1), value, 1)
	}
	return query
}

type SQLValue string

func execReturn[T any](ctx context.Context, db *DB, sql string, args []any, scanFn pgx.RowToFunc[T]) (T, error) {
	var result T
	rows, err := db.Query(ctx, sql, args...)
	if err != nil {
		//	//l.Error(ctx, err)
		return result, err
	}
	defer rows.Close()
	collectedRow, err := pgx.CollectOneRow(rows, scanFn)

	if err != nil {
		return result, err
	}

	return collectedRow, nil
}

func UpdateReturning[T any](ctx context.Context, db *DB, query bob.Query, scanFn pgx.RowToFunc[T]) (T, error) {
	var result T
	sql, args, err := query.BuildN(ctx, 1)
	if err != nil {
		//	//l.Error(ctx, err)
		return result, err
	}
	collectedrows, err := execReturn(ctx, db, sql, args, scanFn)

	if err != nil {
		//	//l.Error(ctx, err)
		return result, err
	}
	return collectedrows, nil

}

func execinsert(ctx context.Context, db *DB, sql string, args []any) (pgconn.CommandTag, error) {

	rows, err := db.Exec(ctx, sql, args...)

	if err != nil {
		//	//l.Error(ctx, err)
		return rows, err
	}

	return rows, err
}

func execupdate(ctx context.Context, db *DB, sql string, args []any) (pgconn.CommandTag, error) {

	rows, err := db.Exec(ctx, sql, args...)

	if err != nil {
		//	//l.Error(ctx, err)
		return rows, err
	}

	return rows, err
}

func execdelete(ctx context.Context, db *DB, sql string, args []any) (pgconn.CommandTag, error) {

	rows, err := db.Exec(ctx, sql, args...)

	if err != nil {
		//	//l.Error(ctx, err)
		return rows, err
	}

	return rows, err
}

func Exec(ctx context.Context, db *DB, sql string, args []any) (pgconn.CommandTag, error) {

	rows, err := db.Query(ctx, sql, args...)

	if err != nil {
		//	//l.Error(ctx, err)
		return pgconn.CommandTag{}, err
	}
	defer rows.Close()
	return rows.CommandTag(), rows.Err()
}

func Update(ctx context.Context, db *DB, query bob.Query) (pgconn.CommandTag, error) {
	sql, args, err := query.BuildN(ctx, 1)
	if err != nil {
		//l.Error(ctx, err)
		return pgconn.CommandTag{}, err
	}

	return execupdate(ctx, db, sql, args)
}

func Delete(ctx context.Context, db *DB, query bob.Query) (pgconn.CommandTag, error) {
	sql, args, err := query.BuildN(ctx, 1)
	if err != nil {
		//l.Error(ctx, err)
		return pgconn.CommandTag{}, err
	}
	return execdelete(ctx, db, sql, args)
}

func Insert(ctx context.Context, db *DB, query bob.Query) (pgconn.CommandTag, error) {
	sql, args, err := query.BuildN(ctx, 1)
	if err != nil {
		//l.Error(ctx, err)
		return pgconn.CommandTag{}, err
	}

	return execinsert(ctx, db, sql, args)
}

func ExecRow(ctx context.Context, db *DB, sql string, args ...any) (pgconn.CommandTag, error) {
	ct, err := Exec(ctx, db, sql, args)
	if err != nil {
		//l.Error(ctx, err)
		return ct, err
	}
	rowsAffected := ct.RowsAffected()
	if rowsAffected == 0 {
		return ct, pgx.ErrNoRows
	}
	return ct, nil
}

func SelectOneOK[T any](ctx context.Context, db *DB, builder bob.Query, scanFn pgx.RowToFunc[T]) (T, bool, error) {

	var zero T
	sql, args, err := builder.BuildN(ctx, 1)
	if err != nil {
		//l.Error(ctx, err)
		return zero, false, err
	}
	rows, err := db.Query(ctx, sql, args...)
	if err != nil {
		if err == pgx.ErrNoRows {
			//l.Error(ctx, err)
			return zero, false, nil
		}
		return zero, false, err
	}
	defer rows.Close()
	collectedRow, b, err := CollectOneRowOK(rows, scanFn)
	if err != nil {
		//l.Error(ctx, err)
		return zero, false, err
	}

	return collectedRow, b, nil
}

func SelectOne[T any](ctx context.Context, db *DB, builder bob.Query, scanFn pgx.RowToFunc[T]) (T, error) {
	var zero T
	sql, args, err := builder.BuildN(ctx, 1)
	if err != nil {
		//l.Error(ctx, err)
		return zero, err
	}
	rows, err := db.Query(ctx, sql, args...)
	if err != nil {
		//l.Error(ctx, err)
		return zero, err
	}
	defer rows.Close()

	collectedRow, err := pgx.CollectOneRow(rows, scanFn)
	if err != nil {
		//l.Error(ctx, err)
		return zero, err
	}

	return collectedRow, nil
}

func InsertReturning[T any](ctx context.Context, db *DB, builder bob.Query, scanFn pgx.RowToFunc[T]) (T, error) {
	var zero T
	sql, args, err := builder.BuildN(ctx, 1)
	if err != nil {
		//l.Error(ctx, err)
		return zero, err
	}
	collectedRow, err := execReturn(ctx, db, sql, args, scanFn)
	if err != nil {
		//l.Error(ctx, err)
		return zero, err
	}
	return collectedRow, nil

}

func SelectRows[T any](ctx context.Context, db *DB, builder bob.Query, scanFn pgx.RowToFunc[T]) ([]T, error) {

	sql, args, err := builder.BuildN(ctx, 1)
	if err != nil {
		//l.Error(ctx, err)
		return nil, err
	}
	rows, err := db.Query(ctx, sql, args...)
	if err != nil {
		//l.Error(ctx, err)
		return nil, err
	}

	defer rows.Close()
	collectedRows, err := pgx.CollectRows(rows, scanFn)
	if err != nil {
		//l.Error(ctx, err)
		return nil, err
	}

	return collectedRows, nil
}

func SelectRowsOK[T any](ctx context.Context, db *DB, builder bob.Query, scanFn pgx.RowToFunc[T]) ([]T, bool, error) {
	sql, args, err := builder.BuildN(ctx, 1)
	if err != nil {
		//l.Error(ctx, err)
		return nil, false, err
	}
	rows, err := db.Query(ctx, sql, args...)
	if err != nil {
		//l.Error(ctx, err)
		if err == pgx.ErrNoRows {
			return nil, false, nil
		}
		return nil, false, err
	}

	defer rows.Close()
	collectedRows, b, err := CollectRowsOK(rows, scanFn)

	if err != nil {
		//l.Error(ctx, err)
		return nil, false, err
	}

	return collectedRows, b, nil
}

func SelectRowsTag[T any](ctx context.Context, db *DB, builder bob.Query, tag string) ([]T, error) {

	sql, args, err := builder.BuildN(ctx, 1)
	if err != nil {
		//l.Error(ctx, err)
		return nil, err
	}
	rows, err := db.Query(ctx, sql, args...)
	if err != nil {
		//l.Error(ctx, err)
		return nil, err
	}
	defer rows.Close()
	collectedRows, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (T, error) {
		return RowToStructByTag[T](row, tag)
	})
	if err != nil {
		//l.Error(ctx, err)
		return nil, err
	}

	return collectedRows, nil
}

func RowToStructByTag[T any](row pgx.CollectableRow, tag string) (T, error) {

	var value T
	err := row.Scan(&tagStructRowScanner{ptrToStruct: &value, lax: true, tag: tag})
	return value, err
}

type tagStructRowScanner struct {
	ptrToStruct any
	lax         bool
	tag         string
}

func (ts *tagStructRowScanner) ScanRow(rows pgx.Rows) error {

	dst := ts.ptrToStruct
	dstValue := reflect.ValueOf(dst)
	if dstValue.Kind() != reflect.Ptr {
		return fmt.Errorf("dst not a pointer")
	}

	dstElemValue := dstValue.Elem()

	scanTargets, err := ts.appendScanTargets(dstElemValue, nil, rows.FieldDescriptions(), ts.tag)

	if err != nil {
		//l.Error(nil, err)
		return err
	}

	for i, t := range scanTargets {

		if t == nil {
			return fmt.Errorf("struct doesn't have corresponding field to match returned column %s", rows.FieldDescriptions()[i].Name)
		}
	}

	return rows.Scan(scanTargets...)
}

func (rs *tagStructRowScanner) appendScanTargets(dstElemValue reflect.Value, scanTargets []any, fldDescs []pgconn.FieldDescription, tagkey string) ([]any, error) {
	var err error
	dstElemType := dstElemValue.Type()

	if scanTargets == nil {
		scanTargets = make([]any, len(fldDescs))
	}

	for i := 0; i < dstElemType.NumField(); i++ {
		sf := dstElemType.Field(i)

		if sf.PkgPath != "" && !sf.Anonymous {

			// Field is unexported, skip it.
			continue
		}

		if sf.Anonymous && sf.Type.Kind() == reflect.Struct {

			scanTargets, err = rs.appendScanTargets(dstElemValue.Field(i), scanTargets, fldDescs, tagkey)
			if err != nil {
				return nil, err
			}
		} else {

			dbTag, dbTagPresent := sf.Tag.Lookup(tagkey)
			if dbTagPresent {

				dbTag = strings.Split(dbTag, ",")[0]
			}
			if dbTag == "-" {

				// Field is ignored, skip it.
				continue
			}

			colName := dbTag

			if !dbTagPresent {

				colName = sf.Name
			}

			fpos := fieldPosByName(fldDescs, colName)

			if fpos == -1 {
				if rs.lax {

					continue
				}
				return nil, fmt.Errorf("cannot find field %s in returned row", colName)
			}
			if fpos >= len(scanTargets) && !rs.lax {
				return nil, fmt.Errorf("cannot find field %s in returned row", colName)
			}

			scanTargets[fpos] = dstElemValue.Field(i).Addr().Interface()
		}
	}

	return scanTargets, err
}

func fieldPosByName(fldDescs []pgconn.FieldDescription, field string) (i int) {
	i = -1
	for i, desc := range fldDescs {
		if strings.EqualFold(desc.Name, field) {
			return i
		}
	}
	return
}

func StructToSetMap(article interface{}) map[string]interface{} {

	setMap := make(map[string]interface{})

	val := reflect.ValueOf(article).Elem()

	for i := 0; i < val.NumField(); i++ {
		field := val.Type().Field(i)
		tag := field.Tag.Get("json")

		// Skip fields without the "db" tag
		if tag == "" {
			continue
		}

		// Check if the value is the zero value for its type
		switch val.Field(i).Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if val.Field(i).Int() == 0 {
				continue
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if val.Field(i).Uint() == 0 {
				continue
			}
		case reflect.Float32, reflect.Float64:
			if val.Field(i).Float() == 0 {
				continue
			}
		case reflect.String:
			if val.Field(i).String() == "" {
				continue
			}
		case reflect.Bool:
			if !val.Field(i).Bool() {
				continue
			}

		case reflect.Struct:
			if val.Field(i).Type() == reflect.TypeOf(time.Time{}) && val.Field(i).Interface().(time.Time).IsZero() {
				continue
			}

		default:
			// Handle other types as needed
		}

		setMap[tag] = val.Field(i).Interface()
	}

	return setMap
}

func Buildertostring(d time.Duration) string {

	stringbuilder := strings.Builder{}
	stringbuilder.WriteString("SET LOCAL statement_timeout = ")
	stringbuilder.WriteString(fmt.Sprintf("%d", d.Milliseconds()))
	return stringbuilder.String()

}
func QueueExecRow(batch *pgx.Batch, builder bob.Query) error {
	var qErr error

	sql, args, err := builder.BuildN(context.Background(), 1)
	if err != nil {
		//l.Error(nil, err)
		return err
	}

	// if !batch.timeoutSet {
	// 	batch.Queue("SET LOCAL statement_timeout = $1", batch.timeoutVal)
	// 	batch.timeoutSet = true
	// }

	//batch.Queue(Buildertostring(d))
	batch.Queue(sql, args...).Exec(func(ct pgconn.CommandTag) error {
		rowsAffected := ct.RowsAffected()
		if rowsAffected == 0 {
			qErr = pgx.ErrNoRows
			return nil
		}
		return nil
	})

	return qErr
}

func QueueReturn[T any](batch *pgx.Batch, builder bob.Query, scanFn pgx.RowToFunc[T], result *[]T) error {
	if err := validateOutputVariable(result); err != nil {
		return err
	}
	var qErr error

	sql, args, err := builder.BuildN(context.Background(), 1)
	if err != nil {
		//l.Error(nil, err)
		return err
	}
	// if !batch.timeoutSet {
	// 	batch.Queue("SET LOCAL statement_timeout = $1", batch.timeoutVal)
	// 	batch.timeoutSet = true
	// }
	//batch.Queue(Buildertostring(d))
	batch.Queue(sql, args...).Query(func(rows pgx.Rows) error {
		collectedRows, err := pgx.CollectRows(rows, scanFn)
		if err != nil {
			//l.Error(nil, err)
			qErr = err
			return nil
		}
		*result = collectedRows
		return nil
	})

	return qErr
}

func QueueReturnRow[T any](batch *pgx.Batch, builder bob.Query, scanFn pgx.RowToFunc[T], result *T) error {

	if err := validateOutputVariable(result); err != nil {
		return err
	}
	var qErr error

	sql, args, err := builder.BuildN(context.Background(), 1)
	if err != nil {
		//l.Error(nil, err)
		return err
	}
	// if !batch.timeoutSet {
	// 	batch.Queue("SET LOCAL statement_timeout = $1", batch.timeoutVal)
	// 	batch.timeoutSet = true
	// }
	batch.Queue(sql, args...).Query(func(rows pgx.Rows) error {
		collectedRow, err := pgx.CollectOneRow(rows, scanFn)
		if err != nil {
			//l.Error(nil, err)
			qErr = err
			return nil
		}

		*result = collectedRow
		return nil
	})

	return qErr
}

func TimedQueueExecRow(batch *TimedBatch, builder bob.Query) error {
	var qErr error

	sql, args, err := builder.BuildN(context.Background(), 1)
	if err != nil {
		//l.Error(nil, err)
		return err
	}

	if !batch.timeoutSet {
		timeoutSQL := fmt.Sprintf("SET LOCAL statement_timeout = %d", batch.timeoutVal)
		batch.Queue(timeoutSQL)
		//batch.addQuery(timeoutSQL)
		//batch.Queue("SET LOCAL statement_timeout = %d", batch.timeoutVal)
		batch.timeoutSet = true
	}
	// interpolatedSQL := interpolateQuery(sql, args) // You'll need to implement this
	// batch.addQuery(interpolatedSQL)
	//batch.Queue(Buildertostring(d))
	batch.Queue(sql, args...).Exec(func(ct pgconn.CommandTag) error {
		rowsAffected := ct.RowsAffected()
		if rowsAffected == 0 {
			qErr = pgx.ErrNoRows
			return nil
		}
		return nil
	})

	return qErr
}

func TimedQueueReturn[T any](batch *TimedBatch, builder bob.Query, scanFn pgx.RowToFunc[T], result *[]T) error {
	if err := validateOutputVariable(result); err != nil {
		return err
	}
	var qErr error

	sql, args, err := builder.BuildN(context.Background(), 1)
	if err != nil {
		//l.Error(nil, err)
		return err
	}
	if !batch.timeoutSet {
		//batch.Queue(fmt.Sprintf("SET LOCAL statement_timeout = %d", batch.timeoutVal))
		timeoutSQL := fmt.Sprintf("SET LOCAL statement_timeout = %d", batch.timeoutVal)
		batch.Queue(timeoutSQL)
		//batch.addQuery(timeoutSQL)
		//		batch.Queue("SET LOCAL statement_timeout = %d", batch.timeoutVal)
		batch.timeoutSet = true
	}
	// interpolatedSQL := interpolateQuery(sql, args) // You'll need to implement this
	// batch.addQuery(interpolatedSQL)
	//batch.Queue(Buildertostring(d))
	batch.Queue(sql, args...).Query(func(rows pgx.Rows) error {
		collectedRows, err := pgx.CollectRows(rows, scanFn)
		if err != nil {
			//l.Error(nil, err)
			qErr = err
			return nil
		}
		*result = collectedRows
		return nil
	})

	return qErr
}

func TimedQueueReturnRow[T any](batch *TimedBatch, builder bob.Query, scanFn pgx.RowToFunc[T], result *T) error {

	if err := validateOutputVariable(result); err != nil {
		return err
	}
	var qErr error

	sql, args, err := builder.BuildN(context.Background(), 1)
	if err != nil {
		//l.Error(nil, err)
		return err
	}
	if !batch.timeoutSet {
		//batch.Queue(fmt.Sprintf("SET LOCAL statement_timeout = %d", batch.timeoutVal))
		timeoutSQL := fmt.Sprintf("SET LOCAL statement_timeout = %d", batch.timeoutVal)
		batch.Queue(timeoutSQL)
		//batch.addQuery(timeoutSQL)

		//batch.Queue("SET LOCAL statement_timeout = %d", batch.timeoutVal)
		batch.timeoutSet = true
	}
	// interpolatedSQL := interpolateQuery(sql, args) // You'll need to implement this
	// batch.addQuery(interpolatedSQL)
	batch.Queue(sql, args...).Query(func(rows pgx.Rows) error {
		collectedRow, err := pgx.CollectOneRow(rows, scanFn)
		if err != nil {
			//l.Error(nil, err)
			qErr = err
			return nil
		}

		*result = collectedRow
		return nil
	})

	return qErr
}

func TxReturnRow[T any](ctx context.Context, tx pgx.Tx, builder bob.Query, scanFn pgx.RowToFunc[T], result *T) error {
	if err := validateOutputVariable(result); err != nil {
		return err
	}

	sql, args, err := builder.BuildN(ctx, 1)
	if err != nil {
		return err
	}
	rows, err := tx.Query(ctx, sql, args...)
	if err != nil {
		//l.Error(ctx, err)
		return err
	}
	defer rows.Close()

	collectedRow, err := pgx.CollectOneRow(rows, scanFn)
	if err != nil {
		return err
	}
	*result = collectedRow
	return nil
}

func TxRows[T any](ctx context.Context, tx pgx.Tx, builder bob.Query, scanFn pgx.RowToFunc[T], result *[]T) error {
	if err := validateOutputVariable(result); err != nil {
		return err
	}

	sql, args, err := builder.BuildN(ctx, 1)
	if err != nil {
		//l.Error(ctx, err)
		return err
	}
	rows, err := tx.Query(ctx, sql, args...)
	if err != nil {
		//l.Error(ctx, err)
		return err
	}
	defer rows.Close()

	collectedRows, err := pgx.CollectRows(rows, scanFn)
	if err != nil {
		//l.Error(ctx, err)
		return err
	}

	*result = collectedRows
	return nil
}

func TxExec(ctx context.Context, tx pgx.Tx, builder bob.Query) error {
	sql, args, err := builder.BuildN(ctx, 1)
	if err != nil {
		//l.Error(ctx, err)
		return err
	}
	_, err = tx.Exec(ctx, sql, args...)
	if err != nil {
		//l.Error(ctx, err)
		return err
	}
	return nil
}

func GenerateMapFromStruct(instance interface{}, tag string) map[string]interface{} {
	result := make(map[string]interface{})

	val := reflect.Indirect(reflect.ValueOf(instance))
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		tag := field.Tag.Get(tag)
		if tag != "" {
			result[tag] = val.Field(i).Interface()
		}
	}
	return result
}

func GenerateColumnsFromStruct(instance interface{}, tag string) []string {
	var columns []string

	val := reflect.Indirect(reflect.ValueOf(instance))
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		tag := field.Tag.Get(tag)
		if tag != "" {
			columns = append(columns, tag)
		}
	}

	return columns
}

func CollectRowsOK[T any](rows pgx.Rows, fn pgx.RowToFunc[T]) ([]T, bool, error) {
	var value []T
	var err error
	value, err = pgx.CollectRows(rows, fn)
	if err != nil {
		//l.Error(nil, err)
		if err == pgx.ErrNoRows {
			return value, false, nil
		}
		return value, false, err
	}
	return value, true, nil
}
func CollectOneRowOK[T any](rows pgx.Rows, fn pgx.RowToFunc[T]) (T, bool, error) {
	var value T
	var err error
	value, err = pgx.CollectOneRow(rows, fn)
	if err != nil {
		//l.Error(nil, err)
		if err == pgx.ErrNoRows {
			return value, false, nil
		}
		return value, false, err
	}
	return value, true, nil
}

func DBQueryMultipleRows(ctx context.Context, query bob.Query, dbs *DB, str interface{}) ([]interface{}, error) {
	// Generate SQL query and arguments
	sql, args, err := query.BuildN(ctx, 1)
	if err != nil {
		return nil, err
	}

	// Execute the query and get the result set
	rows, err := dbs.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Get the type of the provided struct
	structType := reflect.TypeOf(str)

	// Create a slice to store the results
	results := make([]interface{}, 0)

	// Iterate over the rows
	for rows.Next() {
		// Create pointers to each field in the struct
		scanArgs := make([]interface{}, structType.NumField())
		for i := 0; i < structType.NumField(); i++ {
			// Create a pointer to the zero value of the field type
			fieldPtr := reflect.New(structType.Field(i).Type).Interface()
			scanArgs[i] = fieldPtr
		}

		// Scan the row into the pointers of each field in the struct
		if err := rows.Scan(scanArgs...); err != nil {
			fmt.Println("Error scanning row:", err)
			continue
		}

		// Create a new instance of the struct
		newStruct := reflect.New(structType).Elem()

		// Assign the scanned values to the fields of the new struct
		for i := 0; i < len(scanArgs); i++ {
			fieldVal := reflect.ValueOf(scanArgs[i]).Elem()
			newStruct.Field(i).Set(fieldVal)
		}

		// Append the new struct to the results slice
		results = append(results, newStruct.Interface())
	}

	// Check for errors during iteration
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func Tx(gctx *gin.Context, dbPool *DB, f func(ctx context.Context, gctx *gin.Context, tx pgx.Tx, params ...interface{}) error, params ...interface{}) error {
	//func withTx1(ctx context.Context, dbPool *pgxpool.Pool, f func(ctx context.Context, tx pgx.Tx, params ...interface{}) params ...interface{},error) error {
	//var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := dbPool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) // Rollback if not committed

	if err := f(ctx, gctx, tx, params...); err != nil {
		// If an error occurred during the transactional logic, rollback
		return err
	}

	// Commit the transaction
	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

func InsertReturningrows[T any](ctx context.Context, db *DB, builder bob.Query, scanFn pgx.RowToFunc[T]) ([]T, error) {

	sql, args, err := builder.BuildN(ctx, 1)
	if err != nil {
		//l.Error(ctx, err)
		return nil, err
	}
	rows, _ := db.Query(ctx, sql, args...)

	defer rows.Close()
	collectedRows, err := pgx.CollectRows(rows, scanFn)
	if err != nil {
		//l.Error(ctx, err)
		return nil, err
	}

	return collectedRows, nil

}

func QueueReturnBulk[T any](batch *pgx.Batch, builder bob.Query, scanFn pgx.RowToFunc[T], result *[]T) error {
	if err := validateOutputVariable(result); err != nil {
		return err
	}

	var qErr error
	// Build the SQL query and arguments
	sql, args, err := builder.BuildN(context.Background(), 1)
	if err != nil {

		return err
	}
	// if !batch.timeoutSet {
	// 	batch.Queue("SET LOCAL statement_timeout = $1", batch.timeoutVal)
	// 	batch.timeoutSet = true
	// }
	//batch.Queue(Buildertostring(d))
	// Queue the query in the batch
	batch.Queue(sql, args...).Query(func(rows pgx.Rows) error {
		// Collect rows into the result slice
		collectedRows, err := pgx.CollectRows(rows, scanFn)
		if err != nil {
			qErr = err

			return nil // Returning nil to continue processing other queries in the batch
		}

		// Append the collected rows to the result slice
		*result = append(*result, collectedRows...)
		return nil
	})

	return qErr
}

func TimedQueueReturnBulk[T any](batch *TimedBatch, builder bob.Query, scanFn pgx.RowToFunc[T], result *[]T) error {
	if err := validateOutputVariable(result); err != nil {
		return err
	}

	var qErr error
	// Build the SQL query and arguments
	sql, args, err := builder.BuildN(context.Background(), 1)
	if err != nil {

		return err
	}
	if !batch.timeoutSet {
		timeoutSQL := fmt.Sprintf("SET LOCAL statement_timeout = %d", batch.timeoutVal)
		batch.Queue(timeoutSQL)
		//batch.addQuery(timeoutSQL)
		//		batch.Queue("SET LOCAL statement_timeout = %d", batch.timeoutVal)
		batch.timeoutSet = true
	}
	// interpolatedSQL := interpolateQuery(sql, args) // You'll need to implement this
	// batch.addQuery(interpolatedSQL)
	//batch.Queue(Buildertostring(d))
	// Queue the query in the batch
	batch.Queue(sql, args...).Query(func(rows pgx.Rows) error {
		// Collect rows into the result slice
		collectedRows, err := pgx.CollectRows(rows, scanFn)
		if err != nil {
			qErr = err

			return nil // Returning nil to continue processing other queries in the batch
		}

		// Append the collected rows to the result slice
		*result = append(*result, collectedRows...)
		return nil
	})

	return qErr
}

func validateOutputVariable[T any](output *T) error {
	if output == nil {
		err := fmt.Errorf("the output variable cannot be nil. Please provide a valid reference")
		appError := apierrors.NewAppError("Error occurred while validating the output variable", "400", err)
		return &appError
	}
	return nil
}

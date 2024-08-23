package gorm

import (
	"context"
	"database/sql"
	"fmt"
)

// Define callbacks for row query
func init() {
	DefaultCallback.RowQuery().Register("gorm:row_query", rowQueryCallback)
}

type RowQueryResult struct {
	Row *sql.Row
}

type RowsQueryResult struct {
	Rows  *sql.Rows
	Error error
}

// queryCallback used to query data from database
func rowQueryCallback(scope *Scope) {
	if result, ok := scope.InstanceGet("row_query_result"); ok {
		scope.prepareQuerySQL()
		if str, ok := scope.Get("gorm:query_option"); ok {
			scope.SQL += addExtraSpaceIfExist(fmt.Sprint(str))
		}

		sqldb := scope.SQLDB()
		var querierRow func(query string, args ...any) *sql.Row

		querierRow = sqldb.QueryRow

		if tmp, ok := sqldb.(interface {
			QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
		}); ok {
			if scope.context != nil {
				querierRow = func(query string, args ...any) *sql.Row {
					return tmp.QueryRowContext(scope.context, query, args...)
				}
			}
		}

		var querier func(query string, args ...any) (*sql.Rows, error)

		querier = sqldb.Query

		if tmp, ok := sqldb.(interface {
			QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
		}); ok {
			if scope.context != nil {
				querier = func(query string, args ...any) (*sql.Rows, error) {
					return tmp.QueryContext(scope.context, query, args...)
				}
			}
		}

		if rowResult, ok := result.(*RowQueryResult); ok {
			rowResult.Row = querierRow(scope.SQL, scope.SQLVars...)
		} else if rowsResult, ok := result.(*RowsQueryResult); ok {
			rowsResult.Rows, rowsResult.Error = querier(scope.SQL, scope.SQLVars...)
		}
	}
}

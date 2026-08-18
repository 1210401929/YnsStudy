package repository

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"sort"
	"strings"
)

var identifierPattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

type SQLRepository struct {
	db *sql.DB
}

func NewSQLRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{db: db}
}

func (r *SQLRepository) DB() *sql.DB {
	return r.db
}

func (r *SQLRepository) Query(ctx context.Context, query string, args ...any) ([]map[string]any, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRows(rows)
}

func (r *SQLRepository) Exec(ctx context.Context, query string, args ...any) (int64, error) {
	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// Execute supports both query and update statements, matching the generic Java SQL endpoint.
func (r *SQLRepository) Execute(ctx context.Context, query string, args ...any) (any, error) {
	if returnsRows(query) {
		return r.Query(ctx, query, args...)
	}
	count, err := r.Exec(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return count, nil
}

func (r *SQLRepository) ExecuteBatch(ctx context.Context, queries []string, args [][]any, composite bool) ([]any, int64, error) {
	if len(queries) != len(args) {
		return nil, 0, fmt.Errorf("SQL 语句与参数列表数量不匹配")
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, 0, err
	}
	defer tx.Rollback()

	results := make([]any, 0, len(queries))
	var total int64
	for i, query := range queries {
		if composite && returnsRows(query) {
			rows, queryErr := tx.QueryContext(ctx, query, args[i]...)
			if queryErr != nil {
				return nil, 0, queryErr
			}
			data, scanErr := scanRows(rows)
			rows.Close()
			if scanErr != nil {
				return nil, 0, scanErr
			}
			results = append(results, data)
			continue
		}
		result, execErr := tx.ExecContext(ctx, query, args[i]...)
		if execErr != nil {
			return nil, 0, execErr
		}
		count, _ := result.RowsAffected()
		total += count
		if composite {
			results = append(results, count)
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, 0, err
	}
	return results, total, nil
}

func (r *SQLRepository) SaveAll(ctx context.Context, saveType, table string, data []map[string]any, key string) ([]map[string]any, error) {
	if err := validIdentifier(table); err != nil {
		return nil, err
	}
	if err := validIdentifier(key); err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return []map[string]any{}, nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	resultRows := make([]map[string]any, 0, len(data))
	for _, item := range data {
		var keyValue any
		switch strings.ToLower(saveType) {
		case "add":
			keyValue, err = insertMap(ctx, tx, table, item, key)
		case "edit":
			keyValue, err = updateMap(ctx, tx, table, item, key)
		default:
			return nil, fmt.Errorf("不支持的操作类型: %s", saveType)
		}
		if err != nil {
			return nil, err
		}

		if keyValue == nil {
			resultRows = append(resultRows, item)
			continue
		}
		rows, queryErr := tx.QueryContext(ctx, "SELECT * FROM `"+table+"` WHERE `"+key+"` = ?", keyValue)
		if queryErr != nil {
			return nil, queryErr
		}
		found, scanErr := scanRows(rows)
		rows.Close()
		if scanErr != nil {
			return nil, scanErr
		}
		if len(found) == 0 {
			resultRows = append(resultRows, item)
		} else {
			resultRows = append(resultRows, found[0])
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return resultRows, nil
}

func insertMap(ctx context.Context, tx *sql.Tx, table string, item map[string]any, key string) (any, error) {
	columns, err := sortedColumns(item)
	if err != nil {
		return nil, err
	}
	if len(columns) == 0 {
		return nil, fmt.Errorf("新增数据不能为空")
	}
	quoted := make([]string, len(columns))
	marks := make([]string, len(columns))
	values := make([]any, len(columns))
	for i, column := range columns {
		quoted[i] = "`" + column + "`"
		marks[i] = "?"
		values[i] = item[column]
	}
	query := "INSERT INTO `" + table + "` (" + strings.Join(quoted, ", ") + ") VALUES (" + strings.Join(marks, ", ") + ")"
	result, err := tx.ExecContext(ctx, query, values...)
	if err != nil {
		return nil, err
	}
	if value, ok := caseInsensitiveValue(item, key); ok && value != nil && fmt.Sprint(value) != "" {
		return value, nil
	}
	if value, err := result.LastInsertId(); err == nil && value > 0 {
		return value, nil
	}
	return nil, nil
}

func updateMap(ctx context.Context, tx *sql.Tx, table string, item map[string]any, key string) (any, error) {
	keyValue, ok := caseInsensitiveValue(item, key)
	if !ok || keyValue == nil || fmt.Sprint(keyValue) == "" {
		return nil, fmt.Errorf("修改数据缺少主键 %s", key)
	}
	columns, err := sortedColumns(item)
	if err != nil {
		return nil, err
	}
	sets := make([]string, 0, len(columns)-1)
	values := make([]any, 0, len(columns))
	for _, column := range columns {
		if strings.EqualFold(column, key) {
			continue
		}
		sets = append(sets, "`"+column+"` = ?")
		values = append(values, item[column])
	}
	if len(sets) == 0 {
		return nil, fmt.Errorf("没有可修改的字段")
	}
	values = append(values, keyValue)
	query := "UPDATE `" + table + "` SET " + strings.Join(sets, ", ") + " WHERE `" + key + "` = ?"
	if _, err := tx.ExecContext(ctx, query, values...); err != nil {
		return nil, err
	}
	return keyValue, nil
}

func sortedColumns(item map[string]any) ([]string, error) {
	columns := make([]string, 0, len(item))
	for column := range item {
		if err := validIdentifier(column); err != nil {
			return nil, err
		}
		columns = append(columns, column)
	}
	sort.Strings(columns)
	return columns, nil
}

func validIdentifier(value string) error {
	if !identifierPattern.MatchString(value) {
		return fmt.Errorf("非法数据库标识符: %q", value)
	}
	return nil
}

func caseInsensitiveValue(item map[string]any, key string) (any, bool) {
	if value, ok := item[key]; ok {
		return value, true
	}
	for candidate, value := range item {
		if strings.EqualFold(candidate, key) {
			return value, true
		}
	}
	return nil, false
}

func returnsRows(query string) bool {
	fields := strings.Fields(strings.TrimSpace(strings.TrimLeft(query, "(\ufeff")))
	if len(fields) == 0 {
		return false
	}
	switch strings.ToLower(fields[0]) {
	case "select", "show", "describe", "desc", "explain", "with":
		return true
	default:
		return false
	}
}

func scanRows(rows *sql.Rows) ([]map[string]any, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	result := make([]map[string]any, 0)
	for rows.Next() {
		values := make([]any, len(columns))
		destinations := make([]any, len(columns))
		for i := range values {
			destinations[i] = &values[i]
		}
		if err := rows.Scan(destinations...); err != nil {
			return nil, err
		}
		row := make(map[string]any, len(columns))
		for i, column := range columns {
			if raw, ok := values[i].([]byte); ok {
				row[column] = string(raw)
			} else {
				row[column] = values[i]
			}
		}
		result = append(result, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

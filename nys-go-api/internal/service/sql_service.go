package service

import (
	"context"
	"fmt"
	"strings"

	"nys-go-api/internal/model"
)

func (s *Service) SelectList(ctx context.Context, query string, args []any) model.Result {
	if err := validateSQL(query, "select"); err != nil {
		return model.Failure(err.Error())
	}
	rows, err := s.Repo.Query(ctx, query, args...)
	if err != nil {
		return dbFailure("SQL查询", err)
	}
	return model.Success(rows)
}

func (s *Service) ExecuteSQL(ctx context.Context, query string, args []any) model.Result {
	if err := validateSQL(query, "execute"); err != nil {
		return model.Failure(err.Error())
	}
	value, err := s.Repo.Execute(ctx, query, args...)
	if err != nil {
		return dbFailure("SQL执行", err)
	}
	if count, ok := value.(int64); ok && len(args) > 0 {
		return model.Success(fmt.Sprintf("执行成功，影响行数：%d", count))
	}
	return model.Success(value)
}

func (s *Service) ExecuteSQLBatch(ctx context.Context, queries []string, args [][]any, composite bool) model.Result {
	for _, query := range queries {
		if err := validateSQL(query, "execute"); err != nil {
			return model.Failure(err.Error())
		}
	}
	results, total, err := s.Repo.ExecuteBatch(ctx, queries, args, composite)
	if err != nil {
		return dbFailure("批量 SQL 执行", err)
	}
	if composite {
		return model.Success(results)
	}
	return model.Success(fmt.Sprintf("批量执行成功，总影响行数：%d", total))
}

func (s *Service) SaveAll(ctx context.Context, saveType, table string, data []map[string]any, key string) model.Result {
	rows, err := s.Repo.SaveAll(ctx, saveType, table, data, key)
	if err != nil {
		return dbFailure("保存数据", err)
	}
	return model.Success(rows)
}

func validateSQL(query, expected string) error {
	normalized := strings.ToLower(strings.TrimSpace(query))
	if normalized == "" {
		return fmt.Errorf("SQL 不能为空")
	}
	if strings.Contains(normalized, "drop ") || strings.Contains(normalized, "truncate ") {
		return fmt.Errorf("禁止执行此类危险SQL")
	}
	if expected == "select" && !strings.HasPrefix(normalized, "select") && !strings.HasPrefix(normalized, "with") {
		return fmt.Errorf("仅允许执行 SELECT 查询")
	}
	return nil
}

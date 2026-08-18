package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/go-sql-driver/mysql"

	"nys-go-api/internal/config"
)

func ConnectMySQL(cfg config.DatabaseConfig) (*sql.DB, error) {
	charset := cfg.Charset
	if charset == "" {
		charset = "utf8mb4"
	}
	location := time.Local
	if cfg.Location != "" && cfg.Location != "Local" {
		loaded, err := time.LoadLocation(cfg.Location)
		if err != nil {
			return nil, fmt.Errorf("加载数据库时区 %q: %w", cfg.Location, err)
		}
		location = loaded
	}

	driverConfig := mysql.NewConfig()
	driverConfig.User = cfg.Username
	driverConfig.Passwd = cfg.Password
	driverConfig.Net = "tcp"
	driverConfig.Addr = fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	driverConfig.DBName = cfg.Name
	driverConfig.Params = map[string]string{"charset": charset}
	driverConfig.ParseTime = true
	driverConfig.Loc = location
	driverConfig.MultiStatements = true
	db, err := sql.Open("mysql", driverConfig.FormatDSN())
	if err != nil {
		return nil, fmt.Errorf("创建 MySQL 连接池: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConnections)
	db.SetMaxIdleConns(cfg.MaxIdleConnections)
	db.SetConnMaxLifetime(time.Duration(cfg.ConnectionMaxLifetimeMinutes) * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.ConnectTimeoutSeconds)*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("连接 MySQL: %w", err)
	}
	return db, nil
}

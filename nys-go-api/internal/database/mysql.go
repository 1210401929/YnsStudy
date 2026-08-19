package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"
	_ "time/tzdata"

	"github.com/go-sql-driver/mysql"

	"nys-go-api/internal/config"
)

// ResolveLocation 将配置中的时区名称解析成 Go 时区。
// 项目内置 time/tzdata，因此即使 Docker 镜像没有安装 tzdata，Asia/Shanghai 仍然可用。
func ResolveLocation(value string) (*time.Location, error) {
	if value == "" || value == "Local" {
		return time.Local, nil
	}
	location, err := time.LoadLocation(value)
	if err != nil {
		return nil, fmt.Errorf("加载数据库时区 %q: %w", value, err)
	}
	return location, nil
}

func ConnectMySQL(cfg config.DatabaseConfig) (*sql.DB, error) {
	charset := cfg.Charset
	if charset == "" {
		charset = "utf8mb4"
	}
	location, err := ResolveLocation(cfg.Location)
	if err != nil {
		return nil, err
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

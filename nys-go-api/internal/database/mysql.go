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

// 数据库读写超时：一次查询的结果超过这个时间还没读到，就判定连接已坏，丢弃并报错。
// 没有这个超时时，一条已经断掉但两端都不知道的连接会让请求永远等下去，
// 这样的连接越积越多，占满连接池后整个网站的数据库页面都会卡死。
const queryIOTimeout = 15 * time.Second

// connMaxIdleTime 空闲超过这个时间的连接直接关闭，下次用时重新建立，避免复用长时间空闲后可能已失效的连接
const connMaxIdleTime = 3 * time.Minute

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
	connectTimeout := time.Duration(cfg.ConnectTimeoutSeconds) * time.Second
	if connectTimeout <= 0 {
		connectTimeout = 5 * time.Second
	}
	driverConfig.Timeout = connectTimeout
	driverConfig.ReadTimeout = queryIOTimeout
	driverConfig.WriteTimeout = queryIOTimeout
	db, err := sql.Open("mysql", driverConfig.FormatDSN())
	if err != nil {
		return nil, fmt.Errorf("创建 MySQL 连接池: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConnections)
	db.SetMaxIdleConns(cfg.MaxIdleConnections)
	db.SetConnMaxLifetime(time.Duration(cfg.ConnectionMaxLifetimeMinutes) * time.Minute)
	db.SetConnMaxIdleTime(connMaxIdleTime)

	ctx, cancel := context.WithTimeout(context.Background(), connectTimeout)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("连接 MySQL: %w", err)
	}
	return db, nil
}

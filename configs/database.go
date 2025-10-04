package configs

import (
	"errors"
	"fmt"
	"github.com/seelly/gorm-oracle"
	"goDict/utils"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log/slog"
	_ "log/slog"
	"net/url"
)

// DatabaseConfig 数据库配置结构
type DatabaseConfig struct {
	Type     string `yaml:"type"` // mysql, postgres, sqlserver, etc.
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Service  string `yaml:"service"` // oracle
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
	Charset  string `yaml:"charset"`
}

// DB 全局数据库实例
var DB *gorm.DB

// InitDatabase 初始化数据库连接
func InitDatabase(config *DatabaseConfig) (*gorm.DB, error) {
	var dialector gorm.Dialector

	// 初始化数据库配置
	if nil == config {
		return nil, errors.New("数据库连接错误")
	}

	// 如果是 SQLite 类型，需要判断文件是否存在
	if "sqlite" == config.Type {
		if !utils.FileExists(config.Database) {
			err := errors.New("数据库文件不存在")
			slog.Error("数据库文件不存在", "error", err)
			return nil, err
		}
	}

	// 对密码进行编码，确保它能安全地放在DSN或URL中
	encodedPassword := url.QueryEscape(config.Password)

	switch config.Type {
	case "MySQL":
		dsn := "%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local"
		dsn = fmt.Sprintf(dsn, config.Username, config.Password, config.Host, config.Port, config.Database, config.Charset)
		dialector = mysql.Open(dsn)
	case "PostgresSQL":
		dsn := "host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Asia/Shanghai"
		dsn = fmt.Sprintf(dsn, config.Host, config.Username, encodedPassword, config.Database, config.Port)
		dialector = postgres.Open(dsn)
	case "SQLServer":
		dsn := "server=%s;user id=%s;password=%s;port=%d;database=%s;TrustServerCertificate=true;encrypt=disable"
		dsn = fmt.Sprintf(dsn, config.Host, config.Username, config.Password, config.Port, config.Database)
		dialector = sqlserver.Open(dsn)
	case "Oracle":
		dsn := oracle.BuildUrl(config.Host, config.Port, config.Service, config.Username, config.Password, nil)
		dialector = oracle.Open(dsn)
	case "SQLite":
		dsn := config.Host
		dialector = sqlite.Open(dsn)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", config.Type)
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: &SqlLogger{
			logLevel: logger.Info, // 设置默认日志级别
		},
	})

	if err != nil {
		return nil, err
	}

	DB = db
	return db, nil
}

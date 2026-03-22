package config

import (
	"gorm.io/gorm/logger"
)

// GORM_LOG_LEVEL_MAP
var gormLogLevelMap = map[string]logger.LogLevel{
	"silent": logger.Silent,
	"error":  logger.Error,
	"warn":   logger.Warn,
	"info":   logger.Info,
}

func ValidateGormLogLevel(level string) bool {
	_, ok := gormLogLevelMap[level]
	return ok
}

func GetGormLogLevel(level string) logger.LogLevel {
	return gormLogLevelMap[level]
}

package loggers

import (
	"secret-svc/api/dtos"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var req = &dtos.RequestLog{}
var logger *zap.Logger

func init() {
	logger = zap.Must(getDefaultConfig().Build())
}

func getDefaultConfig() zap.Config {
	encoderCfg := zap.NewProductionConfig()
	encoderCfg.EncoderConfig.TimeKey = "timestamp"
	encoderCfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	config := zap.Config{
		Level:             zap.NewAtomicLevelAt(zap.InfoLevel),
		Development:       false,
		DisableCaller:     false,
		DisableStacktrace: true,
		Sampling:          nil,
		Encoding:          "json",
		EncoderConfig:     encoderCfg.EncoderConfig,
		OutputPaths: []string{
			"stderr",
		},
		ErrorOutputPaths: []string{
			"stderr",
		},
	}
	return config
}

func getExtendedConfig() zap.Config {
	encoderCfg := zap.NewProductionConfig()
	encoderCfg.EncoderConfig.TimeKey = "timestamp"
	encoderCfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	config := zap.Config{
		Level:             zap.NewAtomicLevelAt(zap.InfoLevel),
		Development:       false,
		DisableCaller:     false,
		DisableStacktrace: true,
		Sampling:          nil,
		Encoding:          "json",
		EncoderConfig:     encoderCfg.EncoderConfig,
		OutputPaths: []string{
			"stderr",
		},
		ErrorOutputPaths: []string{
			"stderr",
		},
		InitialFields: map[string]interface{}{
			"traceId":        req.TraceId,
			"organizationId": req.OrgId,
			"projectId":      req.ProjectId,
			"scope":          req.Scope,
		},
	}
	return config
}

func GetCustomLogger() *zap.Logger {
	return logger
}

func UpdateLoggerData(newReq *dtos.RequestLog) {
	req = newReq
	logger = zap.Must(getExtendedConfig().Build())
	zap.ReplaceGlobals(logger)
}

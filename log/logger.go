package log

import (
	"os"
	"time"

	"github.com/easynet-cn/file-service/util"
	"github.com/natefinch/lumberjack"
	"github.com/spf13/viper"
	"go.elastic.co/ecszap"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	Logger *zap.Logger
)

func InitLogger(viper *viper.Viper) {
	hook := lumberjack.Logger{
		Filename:   viper.GetString("logging.file"),
		MaxSize:    10,
		MaxBackups: 30,
		MaxAge:     7,
		Compress:   true,
	}
	write := zapcore.AddSync(&hook)

	var level zapcore.Level

	switch viper.GetString("logging.level") {
	case "debug":
		level = zap.DebugLevel
	case "info":
		level = zap.InfoLevel
	case "error":
		level = zap.ErrorLevel
	case "warn":
		level = zap.WarnLevel
	default:
		level = zap.InfoLevel
	}

	encoderConfig := ecszap.NewDefaultEncoderConfig()

	var writes = []zapcore.WriteSyncer{write}

	if level == zap.DebugLevel {
		writes = append(writes, zapcore.AddSync(os.Stdout))
	}

	core := ecszap.NewCore(
		encoderConfig,
		zapcore.NewMultiWriteSyncer(writes...),
		level,
	)

	Logger = zap.New(
		core, zap.AddCaller(),
		zap.Fields(
			zap.String("application", viper.GetString("spring.application.name")),
			zap.String("serverIp", util.LocalIp()),
			zap.Int("port", viper.GetInt("server.port")),
			zap.String("profile", viper.GetString("spring.profiles.active")),
			zap.String("logTime", time.Now().Format(viper.GetString("logging.date-time-format"))),
		),
	)
}

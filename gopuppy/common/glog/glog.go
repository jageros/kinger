package glog

import (
	"encoding/json"
	"runtime/debug"
	"strings"

	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"sync"
)

var (
	// DebugLevel level
	DebugLevel Level = Level(zap.DebugLevel)
	// InfoLevel level
	InfoLevel Level = Level(zap.InfoLevel)
	// WarnLevel level
	WarnLevel Level = Level(zap.WarnLevel)
	// ErrorLevel level
	ErrorLevel Level = Level(zap.ErrorLevel)
	// PanicLevel level
	PanicLevel Level = Level(zap.PanicLevel)
	// FatalLevel level
	FatalLevel Level = Level(zap.FatalLevel)
)

type (
	logFormatFunc   func(format string, args ...interface{})
	Level           = zapcore.Level
	Field           = zap.Field
	ObjectMarshaler = zapcore.ObjectMarshaler
	ObjectEncoder   = zapcore.ObjectEncoder
)

var (
	cfg          zap.Config
	logger       *zap.Logger
	sugar        *zap.SugaredLogger
	errSugar     *zap.SugaredLogger
	source       string
	currentLevel Level

	guard    sync.Mutex
	type2Log = map[string]*zap.Logger{}

	String  = zap.String
	Strings = zap.Strings
	Bool    = zap.Bool
	Int     = zap.Int
	Int32   = zap.Int32
	Uint32  = zap.Uint32
	Int64   = zap.Int64
	Uint64  = zap.Uint64
	Ints    = zap.Ints
	Int32s  = zap.Int32s
	Uint32s = zap.Uint32s
	Int64s  = zap.Int64s
	Float64 = zap.Float64
)

type objects []zapcore.ObjectMarshaler

func (os objects) MarshalLogArray(arr zapcore.ArrayEncoder) error {
	for i := range os {
		arr.AppendObject(os[i])
	}
	return nil
}

func Objects(key string, os []ObjectMarshaler) Field {
	return zap.Array(key, objects(os))
}

func init() {
	var err error
	cfgJson := []byte(`{
		"level": "debug",
		"outputPaths": ["stderr"],
		"errorOutputPaths": ["stderr"],
		"encoding": "console",
		"encoderConfig": {
			"messageKey": "message",
			"levelKey": "level",
			"levelEncoder": "lowercase"
		}
	}`)
	currentLevel = InfoLevel

	if err = json.Unmarshal(cfgJson, &cfg); err != nil {
		panic(err)
	}
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	rebuildLoggerFromCfg()
}

//@param app (gate|center|game|...)
func SetSource(app string) {
	source = app
	rebuildLoggerFromCfg()
}

func SetLevel(lv Level) {
	currentLevel = lv
	cfg.Level.SetLevel(lv)
}

func GetLevel() Level {
	return currentLevel
}

func TraceError(format string, args ...interface{}) {
	Error(string(debug.Stack()))
	Errorf(format, args...)
}

func SetOutput(outputs []string) {
	cfg.OutputPaths = outputs
	rebuildLoggerFromCfg()
}

func ParseLevel(s string) Level {
	if strings.ToLower(s) == "debug" {
		return DebugLevel
	} else if strings.ToLower(s) == "info" {
		return InfoLevel
	} else if strings.ToLower(s) == "warn" || strings.ToLower(s) == "warning" {
		return WarnLevel
	} else if strings.ToLower(s) == "error" {
		return ErrorLevel
	} else if strings.ToLower(s) == "panic" {
		return PanicLevel
	} else if strings.ToLower(s) == "fatal" {
		return FatalLevel
	}
	Errorf("StringToLevel: unknown level: %s", s)
	return DebugLevel
}

func rebuildLoggerFromCfg() {
	if sugar != nil {
		sugar.Sync()
	}
	if errSugar != nil {
		errSugar.Sync()
	}

	cfg := zap.NewProductionEncoderConfig()
	cfg.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.MessageKey = "message"
	cfg.LevelKey = "level"
	cfg.EncodeLevel = zapcore.LowercaseLevelEncoder

	errCore := zapcore.NewCore(
		zapcore.NewConsoleEncoder(cfg),
		os.Stdout,
		zapcore.Level(currentLevel),
	)

	var w zapcore.WriteSyncer
	w = &esWriter{type_: infoLogType}
	w = zapcore.NewMultiWriteSyncer(w, os.Stdout)
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(cfg),
		w,
		zapcore.Level(currentLevel),
	)

	errSugar = zap.New(errCore).Sugar()
	errSugar = errSugar.Named("log")
	sugar = zap.New(core).Sugar()
	if source != "" {
		sugar = sugar.With(zap.String("source", source))
		errSugar = errSugar.With(zap.String("source", source))
	}
}

func SetupGLog(app string, appID uint16, logLevel string) {
	SetSource(fmt.Sprintf("%s%d", app, appID))
	Infof("Set log level to %s", logLevel)
	SetLevel(ParseLevel(logLevel))
}

func Debugf(format string, args ...interface{}) {
	sugar.Debugf(format, args...)
}

func Infof(format string, args ...interface{}) {
	sugar.Infof(format, args...)
}

func Warnf(format string, args ...interface{}) {
	sugar.Warnf(format, args...)
}

func Errorf(format string, args ...interface{}) {
	sugar.Errorf(format, args...)
}

func Panicf(format string, args ...interface{}) {
	sugar.Panicf(format, args...)
}

func Fatalf(format string, args ...interface{}) {
	debug.PrintStack()
	sugar.Fatalf(format, args...)
}

func Error(args ...interface{}) {
	sugar.Error(args...)
}

func Panic(args ...interface{}) {
	sugar.Panic(args...)
}

func Fatal(args ...interface{}) {
	sugar.Fatal(args...)
}

func setSugar(sugar_ *zap.SugaredLogger) {
	sugar = sugar_
}

func JsonInfo(type_ string, fields ...Field) {
	log, ok := type2Log[type_]
	if !ok {
		guard.Lock()
		log, ok = type2Log[type_]
		if !ok {

			var w zapcore.WriteSyncer
			w = &esWriter{type_: type_}
			// TODO
			if type_ == "logout" || type_ == "loginext" {
				w = zapcore.NewMultiWriteSyncer(os.Stdout)
			} else {
				w = zapcore.NewMultiWriteSyncer(w, os.Stdout)
			}

			cfg := zap.NewProductionEncoderConfig()
			cfg.EncodeTime = zapcore.ISO8601TimeEncoder

			core := zapcore.NewCore(
				zapcore.NewJSONEncoder(cfg),
				w,
				zap.InfoLevel,
			)
			log = zap.New(core)
			log = log.Named(type_)
			if source != "" {
				log = log.With(zap.String("source", source))
			}

			type2Log[type_] = log

		}

		guard.Unlock()
	}

	log.Info("", fields...)
}

func Close() {
	if sugar != nil {
		sugar.Sync()
	}
	if errSugar != nil {
		errSugar.Sync()
	}

	for _, log := range type2Log {
		log.Sync()
	}
	closeEs()
}

func Flush() error {
	var err error
	if sugar != nil {
		err = sugar.Sync()
		if err != nil {
			return err
		}
	}
	if errSugar != nil {
		err = errSugar.Sync()
		if err != nil {
			return err
		}
	}

	for _, log := range type2Log {
		err = log.Sync()
		if err != nil {
			return err
		}
	}
	return nil
}

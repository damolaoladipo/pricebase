package logger

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	maxSizeMB = 5
	maxFiles  = 5
	service   = "pricebase-api"
)

// UserLogger writes rotating JSON to user.log.
var UserLogger *zap.Logger

// DBLogger writes rotating JSON to db.log.
var DBLogger *zap.Logger

// SystemLogger writes rotating JSON to system.log.
var SystemLogger *zap.Logger

// MonitorLogger ships to Better Stack with LOGTAIL_MONITOR_TOKEN.
var MonitorLogger *zap.Logger

// APILogger ships to Better Stack with LOGTAIL_AUDIT_TOKEN.
var APILogger *zap.Logger

var exceptionsWriter zapcore.WriteSyncer

func init() {
	dir := logsDir()
	_ = os.MkdirAll(dir, 0o755)

	exceptionsWriter = zapcore.AddSync(rotatingFile(filepath.Join(dir, "exceptions.log")))

	UserLogger = newFileLogger("user", filepath.Join(dir, "user.log"))
	DBLogger = newFileLogger("db", filepath.Join(dir, "db.log"))
	SystemLogger = newFileLogger("system", filepath.Join(dir, "system.log"))
	MonitorLogger = newRemoteLogger("monitor", os.Getenv("LOGTAIL_MONITOR_TOKEN"), os.Getenv("LOGTAIL_INGESTION_URL"))
	APILogger = newRemoteLogger("api", os.Getenv("LOGTAIL_AUDIT_TOKEN"), os.Getenv("LOGTAIL_INGESTION_URL"))
}

// Sync flushes buffered zap cores. Call on process shutdown.
func Sync() {
	_ = UserLogger.Sync()
	_ = DBLogger.Sync()
	_ = SystemLogger.Sync()
	_ = MonitorLogger.Sync()
	_ = APILogger.Sync()
}

func logsDir() string {
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "."
	}

	logsPath := os.Getenv("LOGS_PATH")
	if logsPath == "" {
		logsPath = "internal/_data/logs"
	}

	if filepath.IsAbs(logsPath) {
		return logsPath
	}

	return filepath.Join(cwd, logsPath)
}

func jsonEncoder() zapcore.Encoder {
	cfg := zap.NewProductionEncoderConfig()
	cfg.TimeKey = "timestamp"
	cfg.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05")
	cfg.MessageKey = "message"
	cfg.LevelKey = "level"
	cfg.EncodeLevel = zapcore.LowercaseLevelEncoder
	cfg.CallerKey = zapcore.OmitKey
	cfg.StacktraceKey = "stacktrace"
	return zapcore.NewJSONEncoder(cfg)
}

func rotatingFile(filename string) *lumberjack.Logger {
	return &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    maxSizeMB,
		MaxBackups: maxFiles,
		LocalTime:  true,
	}
}

func newFileLogger(name, filename string) *zap.Logger {
	fileCore := zapcore.NewCore(
		jsonEncoder(),
		zapcore.AddSync(rotatingFile(filename)),
		zap.DebugLevel,
	)
	exceptionCore := zapcore.NewCore(
		jsonEncoder(),
		exceptionsWriter,
		zap.ErrorLevel,
	)

	return zap.New(
		zapcore.NewTee(fileCore, exceptionCore),
		zap.AddStacktrace(zap.ErrorLevel),
	).With(
		zap.String("service", service),
		zap.String("logger", name),
	)
}

func newRemoteLogger(name, token, endpoint string) *zap.Logger {
	if token == "" || endpoint == "" {
		return zap.NewNop().With(
			zap.String("service", service),
			zap.String("logger", name),
		)
	}

	syncer := zapcore.AddSync(&logtailSyncer{
		token:    token,
		endpoint: endpoint,
		client:   &http.Client{Timeout: 5 * time.Second},
	})

	core := zapcore.NewCore(jsonEncoder(), syncer, zap.DebugLevel)

	return zap.New(
		core,
		zap.AddStacktrace(zap.ErrorLevel),
	).With(
		zap.String("service", service),
		zap.String("logger", name),
	)
}

type logtailSyncer struct {
	token    string
	endpoint string
	client   *http.Client
}

func (w *logtailSyncer) Write(p []byte) (int, error) {
	body := bytes.Clone(p)
	go w.post(body)
	return len(p), nil
}

func (w *logtailSyncer) Sync() error {
	return nil
}

func (w *logtailSyncer) post(body []byte) {
	req, err := http.NewRequest(http.MethodPost, w.endpoint, bytes.NewReader(body))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+w.token)

	resp, err := w.client.Do(req)
	if err != nil {
		return
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()
}

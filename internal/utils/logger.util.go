package utils

import (
	"encoding/json"
	"fmt"

	commonutils "github.com/damolaoladipo/common-go/utils"
	"github.com/damolaoladipo/pricebase/internal/common/logger"
	"github.com/fatih/color"
	"go.uber.org/zap"
)

type Logger struct{}

// NewLogger allocates a new Logger value for structured and console logging.
//
// The returned Logger has no internal mutable state beyond method-local buffers used
// during serialization. Multiple instances behave identically because configuration is
// read from the environment singleton at log emission time.
//
// Side effects: allocates a struct only; no I/O or environment loading occurs here.
//
// Returns a non-nil *Logger. The package also exposes a default instance as Log.
func NewLogger() *Logger {
	return &Logger{}
}

var Log = NewLogger()

// buildLogReplacer constructs a json.Marshal replacer function for safe log serialization.
//
// Each call creates a fresh seen map keyed by pointer address strings. error values are
// converted to {"name": type, "message": err.Error()}. nil passes through unchanged.
// Primitive types (strings, numeric types, bool) pass through unchanged. Other values
// are tracked by fmt.Sprintf("%p", value); repeated pointers yield "[Circular]".
//
// Side effects: the returned closure mutates its captured seen map during marshaling.
// A new replacer should be created per top-level serialize operation.
//
// Returns a func(string, any) any suitable for json.Encoder or custom marshal paths.
// The replacer is not currently wired into safeSerialize but supports future use.
func (l *Logger) buildLogReplacer() func(string, any) any {
	seen := map[any]bool{}

	return func(_ string, value any) any {
		if err, ok := value.(error); ok {
			return map[string]any{
				"name":    fmt.Sprintf("%T", err),
				"message": err.Error(),
			}
		}

		if value == nil {
			return value
		}

		switch value.(type) {
		case string, bool, float32, float64, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			return value
		}

		ptr := fmt.Sprintf("%p", value)
		if seen[ptr] {
			return "[Circular]"
		}
		seen[ptr] = true

		return value
	}
}

// safeSerialize JSON-encodes value with indentation for human-readable log output.
//
// The value parameter is marshaled with json.MarshalIndent using two-space indentation.
// When value is nil, no encoding is attempted. When marshaling fails (circular refs,
// unsupported types), a fallback string is produced via fmt.Sprintf("%v", value).
//
// Side effects: none beyond temporary allocation during marshal; does not write to stdout
// or remote log sinks.
//
// Returns a pointer to the encoded or fallback string on success paths. Returns nil when
// value is nil.
func (l *Logger) safeSerialize(value any) *string {
	if value == nil {
		return nil
	}

	encoded, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		fallback := fmt.Sprintf("%v", value)
		return &fallback
	}

	text := string(encoded)
	return &text
}

// safeClone produces a deep copy of value suitable for passing to structured loggers.
//
// Cloning is implemented as JSON marshal (via safeSerialize) followed by unmarshal into
// map[string]any or generic any, stripping non-JSON types and breaking shared references.
// This reduces risk of mutating original maps/slices after logging and avoids circular
// reference issues in downstream log pipelines.
//
// Side effects: temporary allocations during round-trip encoding; does not modify value.
//
// Returns nil when value is nil, safeSerialize returns nil, or json.Unmarshal fails.
// Returns the cloned any value (typically map[string]any) on success.
func (l *Logger) safeClone(value any) any {
	serialized := l.safeSerialize(value)
	if serialized == nil {
		return nil
	}

	var cloned any
	if err := json.Unmarshal([]byte(*serialized), &cloned); err != nil {
		return nil
	}

	return cloned
}

// env resolves the active ENV configuration for log routing decisions.
//
// When utils.Current returns a non-nil singleton from a prior NewENV call, that value
// is used. Otherwise NewENV is invoked once. If NewENV fails (missing required vars),
// an empty ENV struct is returned so logging can continue with conservative defaults
// (no dev/staging/prod flags set).
//
// Side effects: may trigger NewENV and .env loading on first log when bootstrap was skipped.
// Subsequent calls reuse the cached singleton when available.
//
// Returns a non-nil *ENV, either the loaded configuration or an empty fallback.
func (l *Logger) env() *ENV {
	if current := Current(); current != nil {
		return current
	}

	env, err := NewENV()
	if err != nil {
		return &ENV{}
	}

	return env
}

// stamp formats today's date as a console log prefix segment.
//
// The current calendar date is obtained via commonutils.DateToday. Year, month, and day
// are rendered with LeadingNum for zero-padded month and day when needed. The result
// ends with ": " for concatenation with class name and message in printColored.
//
// Side effects: none beyond reading the system clock through DateToday.
//
// Returns a string like "2026/09/10: " used as the dtb segment in colored output.
func (l *Logger) stamp() string {
	today := commonutils.DateToday()
	return fmt.Sprintf("%d/%s/%s: ", today.Year, commonutils.LeadingNum(today.Month), commonutils.LeadingNum(today.Date))
}

// displayDirect determines whether direct-level logs should be written to the console.
//
// The options parameter may be nil when callers omit LogDirectDTO. When options.Display
// is nil, console output is enabled by default to preserve backward-compatible verbose
// logging in development. When Display is non-nil, its boolean value is returned verbatim.
//
// Side effects: none; read-only inspection of options.
//
// Returns true when options or Display is nil, or when *options.Display is true.
// Returns false only when Display is explicitly set to false.
func displayDirect(options *LogDirectDTO) bool {
	if options == nil || options.Display == nil {
		return true
	}

	return *options.Display
}

// displayRequest determines whether request-style logs should be written to the console.
//
// The options parameter may be nil for bare Log calls. When options.Display is nil,
// console output defaults to true. When Display is set, its value controls visibility
// independently of the Monitor flag (remote shipping may still occur).
//
// Side effects: none; read-only inspection of options.
//
// Returns true when options or Display is nil, or when *options.Display is true.
// Returns false only when Display is explicitly set to false.
func displayRequest(options *LogRequestDTO) bool {
	if options == nil || options.Display == nil {
		return true
	}

	return *options.Display
}

// classNameDirect resolves the logical class or component name for direct log entries.
//
// When options is non-nil and options.ClassName is non-empty, that string is returned.
// Otherwise fallback is used—typically the log message itself when no explicit class
// name was provided by the caller.
//
// Side effects: none; read-only inspection of options.
//
// Returns the resolved class name string, never empty when fallback is non-empty.
func classNameDirect(options *LogDirectDTO, fallback string) string {
	if options != nil && options.ClassName != "" {
		return options.ClassName
	}

	return fallback
}

// printColored writes a colorized log line and optional payload to stdout.
//
// When className is non-empty, output format is "{className} - {dtb}: {colored message}".
// When className is empty, only the colored message line is printed. The paint function
// comes from fatih/color (BlueString, RedString, etc.) based on log level. When payload
// is non-nil, safeSerialize output is printed on subsequent lines.
//
// Side effects: writes to color.Output (typically stdout). May serialize payload in memory.
//
// Returns nothing. Does not ship logs to remote monitors; emitDirect handles that separately.
func (l *Logger) printColored(className string, dtb string, message string, paint func(format string, a ...interface{}) string, payload any) {
	if className != "" {
		fmt.Fprintf(color.Output, "%s - %s: %s\n", className, dtb, paint(message))
	} else {
		fmt.Fprintln(color.Output, paint(message))
	}

	if payload != nil {
		if serialized := l.safeSerialize(payload); serialized != nil {
			fmt.Println(*serialized)
		}
	}
}

// monitor ships a structured log entry to logger.MonitorLogger at the mapped severity level.
//
// Attributes include serialized payload (when present), message text, and metadata with
// className and loggedAt (dtb). LogInfo, LogWarning, and LogError map to Info, Warn, and
// Error on MonitorLogger respectively. LogSuccess maps to Debug. Unknown levels default
// to Info.
//
// Side effects: invokes the configured MonitorLogger backend (potentially remote Logtail
// or file sink). Payload is serialized via safeSerialize when non-nil.
//
// Returns nothing. Errors from the underlying logger are not propagated to callers.
func (l *Logger) monitor(level LogType, className string, message string, payload any, dtb string) {
	fields := []zap.Field{
		zap.Any("data", nil),
		zap.String("message", message),
		zap.Any("metadata", map[string]any{
			"className": className,
			"loggedAt":  dtb,
		}),
	}

	if payload != nil {
		fields[0] = zap.Any("data", l.safeSerialize(payload))
	}

	switch level {
	case LogInfo:
		logger.MonitorLogger.Info(className, fields...)
	case LogWarning:
		logger.MonitorLogger.Warn(className, fields...)
	case LogError:
		logger.MonitorLogger.Error(className, fields...)
	case LogSuccess:
		logger.MonitorLogger.Debug(className, fields...)
	default:
		logger.MonitorLogger.Info(className, fields...)
	}
}

// emitDirect routes a direct log entry to console and/or remote monitoring based on environment.
//
// Environment flags from env() control behavior: in development or staging, colored console
// output appears when displayDirect allows it, and remote monitor shipping occurs when
// options.Monitor is true. In production, only Monitor-enabled entries are forwarded to
// monitor; console output is suppressed regardless of Display.
//
// Level selects console color (blue info, yellow warning, red error, green success, white
// default). classNameDirect resolves the monitor class name; when ClassName is empty,
// the message itself is used as remoteName for remote logs.
//
// Side effects: may write to stdout via printColored and invoke monitor for remote sinks.
// Reads environment on each call through env().
//
// Returns nothing. options may be nil; payload and flags are read defensively.
func (l *Logger) emitDirect(level LogType, message string, options *LogDirectDTO) {
	env := l.env()
	dtb := l.stamp()
	monitor := options != nil && options.Monitor
	display := displayDirect(options)
	className := classNameDirect(options, message)
	var payload any
	if options != nil {
		payload = options.Payload
	}

	paint := color.WhiteString
	switch level {
	case LogInfo:
		paint = color.BlueString
	case LogWarning:
		paint = color.YellowString
	case LogError:
		paint = color.RedString
	case LogSuccess:
		paint = color.GreenString
	}

	if env.IsDev || env.IsStaging {
		if display {
			name := ""
			if options != nil {
				name = options.ClassName
			}
			l.printColored(name, dtb, message, paint, payload)
		}

		if monitor {
			remoteName := className
			if options == nil || options.ClassName == "" {
				remoteName = message
			}
			l.monitor(level, remoteName, message, payload, dtb)
		}
	}

	if env.IsProd && monitor {
		remoteName := className
		if options == nil || options.ClassName == "" {
			remoteName = message
		}
		l.monitor(level, remoteName, message, payload, dtb)
	}
}

// Log writes a general-purpose log message with optional request-style configuration.
//
// When the first options entry has a non-empty Type field, the call is redirected to
// emitDirect at that LogType with equivalent ClassName, Payload, Monitor, and Display
// fields copied into LogDirectDTO. Otherwise the request path applies: plain message
// to console in dev/staging when displayRequest allows, plus optional monitor shipping.
//
// Only the first variadic options pointer is considered when provided. ClassName from
// options overrides the default monitor class name (initially the message string).
//
// Side effects: may write to stdout and/or MonitorLogger depending on environment and
// flags. Invokes env() on each call.
//
// Returns nothing. Safe to call with zero options for a simple console line in dev/staging.
func (l *Logger) Log(message string, options ...*LogRequestDTO) {
	var opts *LogRequestDTO
	if len(options) > 0 {
		opts = options[0]
	}

	if opts != nil && opts.Type != "" {
		direct := &LogDirectDTO{
			ClassName: opts.ClassName,
			Payload:   opts.Payload,
			Monitor:   opts.Monitor,
			Display:   opts.Display,
		}
		l.emitDirect(LogType(opts.Type), message, direct)
		return
	}

	env := l.env()
	dtb := l.stamp()
	monitor := opts != nil && opts.Monitor
	display := displayRequest(opts)
	var payload any
	className := message
	if opts != nil {
		payload = opts.Payload
		if opts.ClassName != "" {
			className = opts.ClassName
		}
	}

	if env.IsDev || env.IsStaging {
		if display {
			fmt.Println(message)
			if payload != nil {
				if serialized := l.safeSerialize(payload); serialized != nil {
					fmt.Println(*serialized)
				}
			}
		}

		if monitor {
			l.monitor(LogInfo, className, message, payload, dtb)
		}
	}

	if env.IsProd && monitor {
		l.monitor(LogInfo, className, message, payload, dtb)
	}
}

// firstDirect extracts the first LogDirectDTO from a variadic options slice.
//
// Info, Warning, Error, Success, and Any pass their optional configuration through this
// helper before delegating to emitDirect. Additional options beyond the first are ignored.
//
// Side effects: none; read-only slice access.
//
// Returns options[0] when len(options) > 0, otherwise nil.
func firstDirect(options []*LogDirectDTO) *LogDirectDTO {
	if len(options) == 0 {
		return nil
	}

	return options[0]
}

// Info emits an informational direct log at LogInfo severity.
//
// The message parameter is the primary log text. Optional LogDirectDTO values control
// ClassName, Payload, Monitor (remote shipping), and Display (console visibility in
// dev/staging). Only the first option pointer is used.
//
// Side effects: delegates to emitDirect, which may write colored blue output to stdout
// and/or forward to MonitorLogger.Info when Monitor is enabled.
//
// Returns nothing.
func (l *Logger) Info(message string, options ...*LogDirectDTO) {
	l.emitDirect(LogInfo, message, firstDirect(options))
}

// Warning emits a warning direct log at LogWarning severity.
//
// The message parameter is the primary log text. Optional LogDirectDTO values control
// ClassName, Payload, Monitor, and Display. Only the first option pointer is used.
//
// Side effects: delegates to emitDirect, which may write yellow-colored output to stdout
// in dev/staging and/or forward to MonitorLogger.Warn when Monitor is enabled.
//
// Returns nothing.
func (l *Logger) Warning(message string, options ...*LogDirectDTO) {
	l.emitDirect(LogWarning, message, firstDirect(options))
}

// Error emits an error direct log at LogError severity.
//
// The message parameter is the primary log text. Optional LogDirectDTO values control
// ClassName, Payload, Monitor, and Display. Only the first option pointer is used.
//
// Side effects: delegates to emitDirect, which may write red-colored output to stdout
// in dev/staging and/or forward to MonitorLogger.Error when Monitor is enabled.
//
// Returns nothing. This method logs a message string; it does not accept a Go error value
// as a separate parameter—embed error details in message or Payload when needed.
func (l *Logger) Error(message string, options ...*LogDirectDTO) {
	l.emitDirect(LogError, message, firstDirect(options))
}

// Success emits a success direct log at LogSuccess severity.
//
// The message parameter is the primary log text. Optional LogDirectDTO values control
// ClassName, Payload, Monitor, and Display. Only the first option pointer is used.
//
// Side effects: delegates to emitDirect, which may write green-colored output to stdout
// in dev/staging and/or forward to MonitorLogger.Debug when Monitor is enabled.
//
// Returns nothing.
func (l *Logger) Success(message string, options ...*LogDirectDTO) {
	l.emitDirect(LogSuccess, message, firstDirect(options))
}

// Any emits a direct log at LogAny severity with default white console coloring.
//
// The message parameter is the primary log text. Optional LogDirectDTO values control
// ClassName, Payload, Monitor, and Display. Only the first option pointer is used.
// LogAny maps to white color in emitDirect and Info level on the monitor fallback path.
//
// Side effects: delegates to emitDirect for console and optional remote output.
//
// Returns nothing.
func (l *Logger) Any(message string, options ...*LogDirectDTO) {
	l.emitDirect(LogAny, message, firstDirect(options))
}

// Audit records a structured audit trail entry for security and compliance review.
//
// The data parameter carries LogAuditDTO fields: Name (audit event label), User (id and
// email, cloned with "not-defined" fallbacks via valueOr), Action, Entity (model name),
// and Request context. All nested values pass through safeClone before logging to avoid
// mutating caller-owned maps.
//
// Side effects: always writes to MonitorLogger.Info with the "audit-log: {Name}" prefix.
// Does not print to console regardless of environment flags.
//
// Returns nothing. Underlying logger errors are not propagated.
func (l *Logger) Audit(data LogAuditDTO) {
	logger.MonitorLogger.Info(fmt.Sprintf("audit-log: %s", data.Name),
		zap.Any("user", l.safeClone(map[string]any{
			"id":    valueOr(data.User.ID, "not-defined"),
			"email": valueOr(data.User.Email, "not-defined"),
		})),
		zap.Any("action", l.safeClone(data.Action)),
		zap.Any("entity", l.safeClone(map[string]any{"model": data.Entity.Model})),
		zap.Any("request", l.safeClone(data.Request)),
	)
}

// Request records a structured HTTP request log for observability and debugging.
//
// The data parameter supplies path, IP, device, headers, body, and metadata. Metadata is
// converted via mapFromAny and individual keys (requestId, correlationId, service,
// method, status, durationMs, channel, language, isAuthenticated, userId, businessId,
// branchId, userType, roles, sessionId) are promoted to top-level log attributes alongside
// the full metadata clone.
//
// Side effects: writes to MonitorLogger.Info with the "request-log: {Name}" prefix.
// Headers and body are cloned via safeClone. Does not write to console.
//
// Returns nothing. Underlying logger errors are not propagated.
func (l *Logger) Request(data LogAuditRequestDTO) {
	meta := mapFromAny(data.Metadata)

	logger.MonitorLogger.Info(fmt.Sprintf("request-log: %s", data.Name),
		zap.String("path", data.Path),
		zap.String("ip", data.IP),
		zap.Any("device", data.Device),
		zap.Any("headers", l.safeClone(data.Headers)),
		zap.Any("body", l.safeClone(data.Body)),
		zap.Any("requestId", meta["requestId"]),
		zap.Any("correlationId", meta["correlationId"]),
		zap.Any("service", meta["service"]),
		zap.Any("method", meta["method"]),
		zap.Any("status", meta["status"]),
		zap.Any("durationMs", meta["durationMs"]),
		zap.Any("channel", meta["channel"]),
		zap.Any("language", meta["language"]),
		zap.Any("isAuthenticated", meta["isAuthenticated"]),
		zap.Any("userId", meta["userId"]),
		zap.Any("businessId", meta["businessId"]),
		zap.Any("branchId", meta["branchId"]),
		zap.Any("userType", meta["userType"]),
		zap.Any("roles", meta["roles"]),
		zap.Any("sessionId", meta["sessionId"]),
		zap.Any("metadata", l.safeClone(data.Metadata)),
	)
}

// API records an outbound third-party or internal API call for integration monitoring.
//
// The data parameter includes Name (log label), Provider (service identifier), Payload
// (optional, serialized via safeSerialize), and IsError (selects severity). When IsError
// is true, APILogger.Error is used; otherwise APILogger.Info.
//
// Side effects: writes to logger.APILogger, separate from MonitorLogger used by Audit and
// Request. Payload serialization may allocate temporary strings.
//
// Returns nothing. Does not print to console. Underlying logger errors are not propagated.
func (l *Logger) API(data LogAPICallDTO) {
	var payload any
	if data.Payload != nil {
		payload = l.safeSerialize(data.Payload)
	}

	if data.IsError {
		logger.APILogger.Error(data.Name, zap.Any("provider", data.Provider), zap.Any("data", payload))
		return
	}

	logger.APILogger.Info(data.Name, zap.Any("provider", data.Provider), zap.Any("data", payload))
}

// valueOr selects a log-safe display value with a string fallback for missing data.
//
// When value is nil, fallback is returned. When value is a string and empty, fallback is
// returned. All other non-nil values—including non-string types—are returned unchanged.
//
// Side effects: none; pure selection logic used by Audit user field formatting.
//
// Returns value or fallback as any, never nil unless both value is nil and fallback is
// empty (fallback is always a non-empty string in current call sites).
func valueOr(value any, fallback string) any {
	if value == nil {
		return fallback
	}

	if text, ok := value.(string); ok && text == "" {
		return fallback
	}

	return value
}

// mapFromAny converts an arbitrary value into map[string]any via JSON encoding.
//
// The value parameter is marshaled to JSON and unmarshaled into map[string]any. This
// normalizes structs and nested types into a flat map suitable for keyed log attribute
// extraction in Request logging.
//
// Side effects: temporary allocations during marshal/unmarshal; does not modify value.
//
// Returns an empty non-nil map when value is nil, json.Marshal fails, or json.Unmarshal
// fails. Returns the decoded map on success.
func mapFromAny(value any) map[string]any {
	if value == nil {
		return map[string]any{}
	}

	encoded, err := json.Marshal(value)
	if err != nil {
		return map[string]any{}
	}

	out := map[string]any{}
	if err := json.Unmarshal(encoded, &out); err != nil {
		return map[string]any{}
	}

	return out
}

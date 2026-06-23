package gondola

import (
	"fmt"
	"strings"
	"time"
)

// Access log format identifiers used in the configuration (log_format).
const (
	logFormatJSON     = "json"
	logFormatCommon   = "common"
	logFormatCombined = "combined"
	logFormatCustom   = "custom"

	// accessLogMsg is the slog message used for access log records.
	accessLogMsg = "access_log"
)

// apacheTimeLayout is the timestamp layout used by the Apache common/combined
// log formats.
const apacheTimeLayout = "02/Jan/2006:15:04:05 -0700"

// formatAccessLog renders an access log line for the non-JSON formats from the
// collected slog attributes. JSON is handled by the slog JSON handler and is
// not produced here.
func formatAccessLog(format, customFormat string, attrs map[string]any, traceID string) string {
	get := func(k string) string {
		v, ok := attrs[k]
		if !ok || v == nil {
			return "-"
		}
		return fmt.Sprintf("%v", v)
	}

	switch format {
	case logFormatCommon:
		return fmt.Sprintf(`%s - - [%s] "%s %s %s" %s %s`,
			get("remote_addr"), time.Now().Format(apacheTimeLayout),
			get("method"), get("request_uri"), get("protocol"),
			get("status_code"), get("body_bytes_sent"))
	case logFormatCombined:
		return fmt.Sprintf(`%s - - [%s] "%s %s %s" %s %s "%s" "%s"`,
			get("remote_addr"), time.Now().Format(apacheTimeLayout),
			get("method"), get("request_uri"), get("protocol"),
			get("status_code"), get("body_bytes_sent"),
			get("referer"), get("user_agent"))
	case logFormatCustom:
		return expandCustomFormat(customFormat, attrs, traceID)
	default:
		return ""
	}
}

// expandCustomFormat replaces ${...} placeholders in the custom format string
// with values from the access log attributes.
func expandCustomFormat(tmpl string, attrs map[string]any, traceID string) string {
	get := func(k string) string {
		v, ok := attrs[k]
		if !ok || v == nil {
			return "-"
		}
		return fmt.Sprintf("%v", v)
	}
	return strings.NewReplacer(
		"${timestamp}", time.Now().Format(apacheTimeLayout),
		"${remote_ip}", get("remote_addr"),
		"${method}", get("method"),
		"${uri}", get("request_uri"),
		"${protocol}", get("protocol"),
		"${status}", get("status_code"),
		"${user_agent}", get("user_agent"),
		"${referer}", get("referer"),
		"${request_time}", get("request_time"),
		"${trace_id}", traceID,
	).Replace(tmpl)
}

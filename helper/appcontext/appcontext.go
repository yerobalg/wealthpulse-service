package appcontext

import (
	"context"
	"time"
)

type contextKey string

const (
	// Context keys
	requestId        contextKey = "RequestId"
	serviceVersion   contextKey = "ServiceVersion"
	userAgent        contextKey = "UserAgent"
	requestStartTime contextKey = "RequestStartTime"
	deviceType       contextKey = "DeviceType"
	sourceIP         contextKey = "SourceIP"

	// Header keys
	HeaderRequestId     = "x-request-id"
	HeaderCacheControl  = "cache-control"
	HeaderUserAgent     = "user-agent"
	HeaderDeviceType    = "x-device-type"
	HeaderXForwardedFor = "x-forwarded-for"
	HeaderXRealIP       = "x-real-ip"
)

type Metadata struct {
	RequestID      string `json:"requestId"`
	UserAgent      string `json:"userAgent"`
	DeviceType     string `json:"deviceType"`
	SourceIP       string `json:"sourceIp"`
	ServiceVersion string `json:"serviceVersion"`
}

// SetRequestId stores the request ID into the context.
func SetRequestId(ctx context.Context, rid string) context.Context {
	return context.WithValue(ctx, requestId, rid)
}

// GetRequestId retrieves the request ID from the context. Returns an empty string if not set.
func GetRequestId(ctx context.Context) string {
	rid, ok := ctx.Value(requestId).(string)
	if !ok {
		return ""
	}

	return rid
}

// SetServiceVersion stores the service version into the context.
func SetServiceVersion(ctx context.Context, version string) context.Context {
	return context.WithValue(ctx, serviceVersion, version)
}

// GetServiceVersion retrieves the service version from the context. Returns an empty string if not set.
func GetServiceVersion(ctx context.Context) string {
	version, ok := ctx.Value(serviceVersion).(string)
	if !ok {
		return ""
	}

	return version
}

// SetUserAgent stores the User-Agent string into the context.
func SetUserAgent(ctx context.Context, ua string) context.Context {
	return context.WithValue(ctx, userAgent, ua)
}

// GetUserAgent retrieves the User-Agent string from the context. Returns an empty string if not set.
func GetUserAgent(ctx context.Context) string {
	ua, ok := ctx.Value(userAgent).(string)
	if !ok {
		return ""
	}

	return ua
}

// SetRequestStartTime stores the request start time into the context.
func SetRequestStartTime(ctx context.Context, t time.Time) context.Context {
	return context.WithValue(ctx, requestStartTime, t)
}

// GetRequestStartTime retrieves the request start time from the context. Returns a zero time.Time if not set.
func GetRequestStartTime(ctx context.Context) time.Time {
	t, ok := ctx.Value(requestStartTime).(time.Time)
	if !ok {
		return time.Time{}
	}

	return t
}

// SetDeviceType stores the device type (platform) into the context.
func SetDeviceType(ctx context.Context, platform string) context.Context {
	return context.WithValue(ctx, deviceType, platform)
}

// GetDeviceType retrieves the device type from the context. Returns "web" if not set.
func GetDeviceType(ctx context.Context) string {
	platform, ok := ctx.Value(deviceType).(string)
	if !ok {
		return "web"
	}

	return platform
}

// GetMetadata builds a Metadata struct from the context values.
func GetMetadata(ctx context.Context) Metadata {
	return Metadata{
		RequestID:      GetRequestId(ctx),
		UserAgent:      GetUserAgent(ctx),
		DeviceType:     GetDeviceType(ctx),
		SourceIP:       GetSourceIP(ctx),
		ServiceVersion: GetServiceVersion(ctx),
	}
}

// SetSourceIP stores the source IP address into the context.
func SetSourceIP(ctx context.Context, ip string) context.Context {
	return context.WithValue(ctx, sourceIP, ip)
}

// GetSourceIP retrieves the source IP address from the context. Returns "127.0.0.1" if not set.
func GetSourceIP(ctx context.Context) string {
	ip, ok := ctx.Value(sourceIP).(string)
	if !ok {
		return "127.0.0.1"
	}

	return ip
}

package helper

import (
	"context"
	"path"
	"reflect"
	"strings"
	"time"

	"github.com/baepo-cloud/viscaufs-server/internal/helper/clog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UnaryServerInterceptorLogger returns a new unary server interceptor that logs requests using the clog package
func UnaryServerInterceptorLogger() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Extract method information
		service := path.Dir(info.FullMethod)[1:]
		method := path.Base(info.FullMethod)

		// Get or create a new clog Line
		ctx, line := clog.FromContext(ctx)
		line = line.Clone()

		// Add request metadata
		line.Add("service", service)
		line.Add("method", method)
		logRequestFields(line, req)

		// Record start time
		startTime := time.Now()

		// Process the RPC
		resp, err := handler(ctx, req)

		// Calculate duration
		duration := time.Since(startTime)

		// Add response metadata
		line.Add("duration_ms", duration.Milliseconds())

		// Handle error and status code
		if err != nil {
			st, ok := status.FromError(err)
			if ok {
				line.Add("code", st.Code().String())
				line.Add("status", st.Code())
			} else {
				line.Add("code", codes.Unknown.String())
				line.Add("status", codes.Unknown)
			}
			line.Error(err)
		} else {
			line.Add("code", codes.OK.String())
			line.Add("status", codes.OK)
			line.Add("response", resp)
		}

		// Log the request
		line.Log("processed gRPC request")

		return resp, err
	}
}

// StreamServerInterceptorLogger returns a new streaming server interceptor that logs streaming RPCs
func StreamServerInterceptorLogger() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		// Extract method information
		service := path.Dir(info.FullMethod)[1:]
		method := path.Base(info.FullMethod)

		// Get or create a new clog Line
		ctx, line := clog.FromContext(ss.Context())
		line = line.Clone()

		// Add request metadata
		line.Add("service", service)
		line.Add("method", method)
		line.Add("stream", true)
		line.Add("stream_type", streamType(info))

		// Record start time
		startTime := time.Now()

		// Process the RPC using a wrapped stream that will intercept the events
		wrappedStream := &serverStream{
			ServerStream: ss,
			ctx:          ctx,
		}
		err := handler(srv, wrappedStream)

		// Calculate duration
		duration := time.Since(startTime)

		// Add response metadata
		line.Add("duration_ms", duration.Milliseconds())
		line.Add("message_count", wrappedStream.messageCount)

		// Handle error and status code
		if err != nil {
			st, ok := status.FromError(err)
			if ok {
				line.Add("code", st.Code().String())
				line.Add("status", st.Code())
			} else {
				line.Add("code", codes.Unknown.String())
				line.Add("status", codes.Unknown)
			}
			line.Error(err)
		} else {
			line.Add("code", codes.OK.String())
			line.Add("status", codes.OK)
		}

		// Log the request
		line.Log("processed gRPC stream")

		return err
	}
}

// Helper to determine the stream type
func streamType(info *grpc.StreamServerInfo) string {
	if info.IsClientStream && info.IsServerStream {
		return "bidirectional"
	} else if info.IsClientStream {
		return "client"
	} else if info.IsServerStream {
		return "server"
	}
	return "unknown" // Should not happen
}

// serverStream wraps grpc.ServerStream to intercept events
type serverStream struct {
	grpc.ServerStream
	ctx          context.Context
	messageCount int
}

func (s *serverStream) Context() context.Context {
	return s.ctx
}

func (s *serverStream) SendMsg(m interface{}) error {
	s.messageCount++
	return s.ServerStream.SendMsg(m)
}

func (s *serverStream) RecvMsg(m interface{}) error {
	err := s.ServerStream.RecvMsg(m)
	if err == nil {
		s.messageCount++
	}
	return err
}

// logRequestFields extracts and logs individual fields from the request object
func logRequestFields(line *clog.Line, req interface{}) {
	if req == nil {
		return
	}

	v := reflect.ValueOf(req)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return
		}
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		// If not a struct, just log the whole thing
		line.Add("field.request", req)
		return
	}

	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		fieldValue := v.Field(i)
		fieldName := strings.ToLower(field.Name)

		// Handle nested structs if needed
		if fieldValue.Kind() == reflect.Struct && !isSimpleType(fieldValue.Type()) {
			// For complex nested structs, you might want to handle them specially
			// Here we're just logging the entire nested struct
			line.Add("field."+fieldName, fieldValue.Interface())
		} else {
			// For simple fields and values, log directly
			line.Add("field."+fieldName, fieldValue.Interface())
		}
	}
}

// isSimpleType determines if a type is a "simple" type that should be logged directly
func isSimpleType(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64, reflect.String:
		return true
	default:
		return false
	}
}

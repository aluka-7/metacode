package metacode

import (
	"fmt"
	"strconv"

	"github.com/aluka-7/metacode/types"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
)

// Error new status with code and message
func Error(code Code, message string) *Status {
	return &Status{s: &types.Status{Code: int32(code.Code()), Message: message}}
}

// Errorf new status with code and message
func Errorf(code Code, format string, args ...interface{}) *Status {
	return Error(code, fmt.Sprintf(format, args...))
}

var _ Codes = &Status{}

// Status statusError is an alias of a status proto
// implement metacode.Codes
type Status struct {
	s *types.Status
}

// Error implement error
func (s *Status) Error() string {
	return s.Message("")
}

// Code return error code
func (s *Status) Code() int {
	return int(s.s.Code)
}

// Message return error message for developer
func (s *Status) Message(_ string) string {
	if s.s.Message == "" {
		return strconv.Itoa(int(s.s.Code))
	}
	return s.s.Message
}

// Details return error details
func (s *Status) Details() []interface{} {
	if s == nil || s.s == nil {
		return nil
	}
	details := make([]interface{}, 0, len(s.s.Details))
	for _, a := range s.s.Details {
		detail := &ptypes.DynamicAny{}
		if err := ptypes.UnmarshalAny(a, detail); err != nil {
			details = append(details, err)
			continue
		}
		details = append(details, detail.Message)
	}
	return details
}

// WithDetails WithDetails
func (s *Status) WithDetails(pbs ...proto.Message) (*Status, error) {
	for _, pb := range pbs {
		anyMsg, err := ptypes.MarshalAny(pb)
		if err != nil {
			return s, err
		}
		s.s.Details = append(s.s.Details, anyMsg)
	}
	return s, nil
}

// Proto return origin protobuf message
func (s *Status) Proto() *types.Status {
	return s.s
}

// FromCode create status from code
func FromCode(code Code) *Status {
	return &Status{s: &types.Status{Code: int32(code), Message: code.Message("")}}
}

// FromProto new status from gRpc detail
func FromProto(pbMsg proto.Message) Codes {
	if msg, ok := pbMsg.(*types.Status); ok {
		if msg.Message == "" || msg.Message == strconv.FormatInt(int64(msg.Code), 10) {
			// NOTE: if message is empty convert to pure Code, will get message from config center.
			return Code(msg.Code)
		}
		return &Status{s: msg}
	}
	return Errorf(ServerErr, "invalid proto message get %v", pbMsg)
}

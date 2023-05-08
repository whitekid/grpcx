package grpcx

import (
	"time"

	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func UInt64[T int64 | uint64](v T) *wrapperspb.UInt64Value {
	return &wrapperspb.UInt64Value{Value: uint64(v)}
}

func Empty() *emptypb.Empty { return &emptypb.Empty{} }

func Timestamppb(in *time.Time) *timestamppb.Timestamp {
	if in == nil {
		return nil
	}

	return timestamppb.New(*in)
}

func TimestampToTimeP(in *timestamppb.Timestamp) *time.Time {
	if in == nil {
		return nil
	}

	t := in.AsTime()
	return &t
}

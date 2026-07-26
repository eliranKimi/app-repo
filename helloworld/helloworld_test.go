package helloworld_test

import (
	"testing"

	pb "app-repo/helloworld"

	"google.golang.org/protobuf/proto"
)

// TestProtoDescriptorValid verifies the raw descriptor bytes in helloworld.pb.go
// are valid. This catches the "slice bounds out of range" panic that occurs when
// the protobuf runtime tries to parse a corrupted or manually-written descriptor.
func TestProtoDescriptorValid(t *testing.T) {
	// This will panic if the raw descriptor is invalid — exactly the bug we saw in prod.
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("protobuf descriptor is invalid (would panic at runtime): %v", r)
		}
	}()

	// Instantiate both message types to trigger descriptor initialization
	req := &pb.HelloRequest{Name: "test"}
	reply := &pb.HelloReply{Message: "hello test"}

	if req.GetName() != "test" {
		t.Errorf("HelloRequest.GetName() = %q, want %q", req.GetName(), "test")
	}
	if reply.GetMessage() != "hello test" {
		t.Errorf("HelloReply.GetMessage() = %q, want %q", reply.GetMessage(), "hello test")
	}
}

// TestProtoMarshalUnmarshal verifies that messages can be serialized and
// deserialized correctly — a basic sanity check for the generated code.
func TestProtoMarshalUnmarshal(t *testing.T) {
	original := &pb.HelloRequest{Name: "eliran"}

	data, err := proto.Marshal(original)
	if err != nil {
		t.Fatalf("proto.Marshal failed: %v", err)
	}

	decoded := &pb.HelloRequest{}
	if err := proto.Unmarshal(data, decoded); err != nil {
		t.Fatalf("proto.Unmarshal failed: %v", err)
	}

	if decoded.GetName() != original.GetName() {
		t.Errorf("round-trip mismatch: got %q, want %q", decoded.GetName(), original.GetName())
	}
}

// TestHelloReplyRoundTrip verifies HelloReply serialization.
func TestHelloReplyRoundTrip(t *testing.T) {
	original := &pb.HelloReply{Message: "Hello eliran from pod-abc123"}

	data, err := proto.Marshal(original)
	if err != nil {
		t.Fatalf("proto.Marshal failed: %v", err)
	}

	decoded := &pb.HelloReply{}
	if err := proto.Unmarshal(data, decoded); err != nil {
		t.Fatalf("proto.Unmarshal failed: %v", err)
	}

	if decoded.GetMessage() != original.GetMessage() {
		t.Errorf("round-trip mismatch: got %q, want %q", decoded.GetMessage(), original.GetMessage())
	}
}

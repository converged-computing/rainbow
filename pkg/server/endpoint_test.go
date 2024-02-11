package server

import (
	"context"
	"log"
	"testing"
	"time"

	pb "github.com/converged-computing/rainbow/pkg/api/v1"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/test/bufconn"
	anypb "google.golang.org/protobuf/types/known/anypb"
	tspb "google.golang.org/protobuf/types/known/timestamppb"
	wrbp "google.golang.org/protobuf/types/known/wrapperspb"
)

var (
	maxTestRunDuration = 180 * time.Second // 3 minutes
)

func TestScalar(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), maxTestRunDuration)
	defer cancel()

	s := startTestServer(ctx, t)
	assert.NotNil(t, s)
	defer s.Stop()

	t.Run("scalar sans args", func(t *testing.T) {
		_, err := NewServer("", "", "")
		assert.Error(t, err)
		_, err = NewServer("test", "", "")
		assert.Error(t, err)
		_, err = NewServer("test", "test", "")
		assert.Error(t, err)
		_, err = NewServer("test", "", "test")
		assert.Error(t, err)
		_, err = NewServer("", "", "test")
		assert.Error(t, err)
	})

	t.Run("scalar sans args", func(t *testing.T) {
		if _, err := s.Serial(ctx, nil); err == nil {
			t.Fatalf("expected error on scalar without args")
		}
	})

	t.Run("scalar with args", func(t *testing.T) {
		data, err := anypb.New(wrbp.String("test"))
		if err != nil {
			t.Fatalf("error creating data: %v", err)
		}

		req := &pb.Request{
			Content: &pb.Content{
				Id:   uuid.New().String(),
				Data: data,
			},
			Sent: tspb.Now(),
		}

		// Scalar example
		res, err := s.Serial(ctx, req)
		if err != nil {
			t.Fatalf("error on scalar: %v", err)
		}

		assert.NotEmpty(t, res.GetRequestId())
		assert.Greater(t, res.GetMessageCount(), int64(0))
		assert.Equal(t, res.GetMessagesProcessed(), res.GetMessageCount())
		assert.Equal(t, success, res.GetProcessingDetails())
	})

	t.Run("stream sans args", func(t *testing.T) {
		if err := s.Stream(nil); err == nil {
			t.Fatalf("expected error on stream without args")
		}
	})
}

func startTestServer(ctx context.Context, t *testing.T) *Server {
	buf := 101024 * 1024
	lis := bufconn.Listen(buf)
	defer lis.Close()

	s, err := NewServer("test-server", "v0.0.0-test", "test")
	if err != nil {
		t.Fatalf("error while creating server: %v", err)
	}

	go func() {
		if err := s.serve(ctx, lis); err != nil && err.Error() != "closed" {
			log.Printf("error on server: %v", err)
		}
	}()

	return s
}

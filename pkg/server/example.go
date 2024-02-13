package server

import (
	"context"
	"log"

	pb "github.com/converged-computing/rainbow/pkg/api/v1"

	"github.com/pkg/errors"
)

// TEST ENDPOINTS ------------------------------------------------
// I'm leaving these here as examples for how to eventually test
// Stream implements the Stream method of the Service.
func (s *Server) Stream(stream pb.RainbowScheduler_StreamServer) error {
	if stream == nil {
		return errors.New("stream is required")
	}

	for {
		in, err := stream.Recv()
		if err != nil {
			return errors.Wrap(err, "failed to receive")
		}

		c := in.GetContent()
		log.Printf("received stream: %v", c.GetData())

		s.counter.Add(1)
		response := &pb.Response{
			RequestId:         c.GetId(),
			MessageCount:      s.GetCounter(),
			MessagesProcessed: s.GetCounter(),
			ProcessingDetails: success,
		}
		if err := stream.Send(response); err != nil {
			return errors.Wrap(err, "failed to send")
		}
	}
}

// Serialimplements the single method of the Service.
func (s *Server) Serial(_ context.Context, in *pb.Request) (*pb.Response, error) {
	if in == nil {
		return nil, errors.New("request is required")
	}

	c := in.GetContent()
	log.Printf("received serial: %v", c.GetData())

	// This is redundant, but for protobuf responses I like to see them directly
	return &pb.Response{
		RequestId:         c.GetId(),
		MessageCount:      s.GetCounter(),
		MessagesProcessed: s.GetCounter(),
		ProcessingDetails: success,
	}, nil
}

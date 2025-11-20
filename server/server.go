package main

import (
	proto "auction/grpc"
	"errors"
	"log"
	"net"
	"os"
	"os/signal"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

/*
	rpc Bid (BidAmount) returns (BidAck);
    rpc Result (Empty)  returns (AuctionResult);
    rpc UpdateReplica (ReplicaState) returns (Empty);
    rpc StartElection (ReplicaIdentity) returns (ElectionResponse);
    rpc ElectionFinished (Leader) returns (ReplicaState);
*/

type AuctionService struct {
	proto.UnimplementedAuctionServer
	servers            map[int64]proto.AuctionClient
	leader_id          int64
	id                 int64
	previous_responses map[int64]proto.AuctionResult
	highest_bid        int64
	highest_bidder     int64
	auction_running    bool
	timestamp          int64
}

func main() {
	/*f, err := os.OpenFile("log.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	log.SetOutput(f)*/
	server := &AuctionService{
		id:              1,
		leader_id:       1,
		timestamp:       0,
		highest_bid:     0,
		highest_bidder:  -1,
		auction_running: false,
		servers:         make(map[int64]proto.AuctionClient),
	}

	server.start_server()
}

func (s *AuctionService) start_server() {
	grpc_server := grpc.NewServer()
	listener, err := net.Listen("tcp", ":8080")

	if err != nil {
		log.Fatal(err)
	}

	proto.RegisterAuctionServer(grpc_server, s)
	log.Println("Server started on " + listener.Addr().String())
	go s.shutdown_logger(grpc_server)
	if s.id == s.leader_id {
		go s.start_auction()
	}
	err = grpc_server.Serve(listener)

	if err != nil {
		log.Fatal(err)
	}
}

func (s *AuctionService) start_auction() {
	for {
		s.highest_bid = 0
		s.highest_bidder = -1
		s.auction_running = true
		log.Println("Auction started")
		s.timestamp += 1
		time.Sleep(time.Millisecond * time.Duration(10000))
		log.Println("Auction over")
		s.auction_running = false
		s.timestamp += 1
		time.Sleep(time.Second * 5)
	}

}

func (s *AuctionService) shutdown_logger(grpc_server *grpc.Server) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	<-stop
	log.Printf("Server stopped with logical time stamp %d\n", s.timestamp)
	log.Println("---------------------------------------------------------------")
	grpc_server.GracefulStop()
}

func (s *AuctionService) update_timestamp(timestamp_in int64) {
	if s.timestamp < timestamp_in {
		s.timestamp = timestamp_in
	}
	s.timestamp += 1
}

func (s *AuctionService) Bid(context context.Context, bid *proto.BidAmount) (*proto.BidAck, error) {
	if s.id != s.leader_id {
		log.Println("Propogating request to leader")
		return s.servers[s.leader_id].Bid(context, bid)
	}
	if !s.auction_running {
		return &proto.BidAck{Accepted: false}, errors.New("auction not running")
	}
	s.update_timestamp(bid.Timestamp)

	if s.highest_bid < bid.BidAmount {
		log.Printf("Accepting bid from %d for %d", bid.Bidder, bid.BidAmount)
		s.highest_bid = bid.BidAmount
		s.highest_bidder = bid.Bidder
		for _, server := range s.servers {
			log.Printf("Updating server %d\n", server)
		}
		return &proto.BidAck{Accepted: true}, nil
	}
	log.Printf("Declining bid from %d for %d", bid.Bidder, bid.BidAmount)
	return &proto.BidAck{Accepted: false}, nil
}

func (s *AuctionService) Result(context context.Context, _ *proto.Empty) (*proto.AuctionResult, error) {
	return &proto.AuctionResult{
		HighestBid:    s.highest_bid,
		HighestBidder: s.highest_bidder,
		AuctionOver:   !s.auction_running,
		Timestamp:     s.timestamp,
	}, nil
}

func (s *AuctionService) UpdateReplica(context context.Context, state *proto.ReplicaState) (*proto.Empty, error) {

	return &proto.Empty{}, nil
}

func (s *AuctionService) StartElection(context context.Context, identity *proto.ReplicaIdentity) (*proto.ElectionResponse, error) {

	return &proto.ElectionResponse{}, nil
}

func (s *AuctionService) ElectionFinished(context context.Context, leader *proto.Leader) (*proto.ReplicaState, error) {

	return &proto.ReplicaState{}, nil
}

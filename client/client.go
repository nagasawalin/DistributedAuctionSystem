package main

import (
	proto "auction/grpc"
	"context"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	println("Starting client")
	conn, err := grpc.NewClient("localhost:8080", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	client := proto.NewAuctionClient(conn)
	bid_ack, err := client.Bid(context.Background(), &proto.BidAmount{BidAmount: 10, Bidder: 1})
	if err != nil {
		log.Println(err)
	} else {
		if bid_ack.Accepted {
			log.Println("Bid accepted")
		} else {
			log.Println("Bid denied")
		}
	}

	bid_ack, err = client.Bid(context.Background(), &proto.BidAmount{BidAmount: 10, Bidder: 2})
	if err != nil {
		log.Println(err)
	} else {
		if bid_ack.Accepted {
			log.Println("Bid accepted")
		} else {
			log.Println("Bid denied")
		}
	}

	auction_result, err := client.Result(context.Background(), &proto.Empty{})
	if !auction_result.AuctionOver {
		log.Printf("highest bidder %d with bid %d\n", auction_result.HighestBidder, auction_result.HighestBid)
	} else {
		log.Printf("Auction over. Highest bidder %d with bid %d\n", auction_result.HighestBidder, auction_result.HighestBid)
	}
}

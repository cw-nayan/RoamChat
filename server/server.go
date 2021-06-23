package main

import (
	"context"
	"fmt"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/joho/godotenv"
	"github.com/nayanmakasare/RoamChat/db_manager"
	pb "github.com/nayanmakasare/RoamChat/proto"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
	"log"
	"net"
	"net/http"
	"os"
)

var (
	mongoUrl     string
	mongoDbName      string
	grpcPort     string
	restPort     string
	msgCollName  string
	roomCollName string
	userCollName string
)

func loadEnv() {
	mongoUrl = os.Getenv("MONGO_DB_URL")
	log.Println(mongoUrl)

	mongoDbName = os.Getenv("CHAT_DB")
	log.Println(mongoDbName)

	roomCollName = os.Getenv("CHAT_ROOM_COLLECTION")
	log.Println(roomCollName)
	userCollName = os.Getenv("CHAT_USER_COLLECTION")
	log.Println(userCollName)
	msgCollName = os.Getenv("CHAT_MSG_COLLECTION")
	log.Println(msgCollName)

	grpcPort = os.Getenv("GRPC_PORT")
	log.Println(grpcPort)
	restPort = os.Getenv("REST_PORT")
	log.Println(restPort)
}

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Println(err.Error())
	}
	loadEnv()
}

func main() {
	go startGRPCServer(grpcPort)
	go startRESTServer(restPort, grpcPort)
	select {}
}

// starting a grpc server
func startGRPCServer(address string) error {
	lis, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	// Create an array of gRPC options with the credentials
	opts := []grpc.ServerOption{}

	// create a gRPC server object
	s := grpc.NewServer(opts...)
	roomColl := db_manager.GetMongoDbCollection(mongoUrl, mongoDbName, roomCollName)
	msgColl := db_manager.GetMongoDbCollection(mongoUrl, mongoDbName, msgCollName)
	userColl := db_manager.GetMongoDbCollection(mongoUrl, mongoDbName, userCollName)
	pb.RegisterRoamChatServer(s, &ChatServer{
		roomColl: roomColl,
		userColl: userColl,
		msgColl:  msgColl,
	})
	return s.Serve(lis)
}

// starting a rest server using grpc-rest.
func startRESTServer(address, grpcAddress string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	mux := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
			MarshalOptions: protojson.MarshalOptions{
				UseProtoNames:   true,
				UseEnumNumbers:  true,
				EmitUnpopulated: true,
			},
			UnmarshalOptions: protojson.UnmarshalOptions{},
		}))
	opts := []grpc.DialOption{grpc.WithInsecure()} // Register ping
	err := pb.RegisterRoamChatHandlerFromEndpoint(ctx, mux, grpcAddress, opts)
	if err != nil {
		return fmt.Errorf("could not register service Ping: %s", err)
	}
	log.Printf("starting HTTP/1.1 REST server on %s", address)
	log.Printf("starting HTTP/2 GRPC server on %s", grpcAddress)
	return http.ListenAndServe(address, mux)
}

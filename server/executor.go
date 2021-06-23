package main

import (
	"context"
	pb "github.com/nayanmakasare/RoamChat/proto"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ChatServer struct {
	roomColl *mongo.Collection
	userColl *mongo.Collection
	msgColl *mongo.Collection
}

//Room CURDL
func (c ChatServer) ListRooms(ctx context.Context, req *pb.ListRoomReq) (*pb.ListRoomResp, error) {
	if len(req.GetUserId()) == 0 {
		//send all the rooms in db
		cur, err := c.roomColl.Find(ctx, bson.D{})
		if err != nil {
			if err != mongo.ErrNoDocuments {
				return nil, status.Error(codes.NotFound, "No rooms in the DB")
			}
			return nil, err
		}
		rooms := []*pb.Room{}

		defer cur.Close(ctx)
		for cur.Next(ctx) {
			room := new(pb.Room)
			rooms = append(rooms, room)
		}
		return &pb.ListRoomResp{Rooms: rooms}, nil
	}
	//if userId is given fetch all the room where the user is present
	//TODO
	return nil, status.Error(codes.Unimplemented, "No implemented with respect to user Id")
}

func (c ChatServer) GetRoom(ctx context.Context, req *pb.GetRoomReq) (*pb.Room, error) {
	query := bson.D{
		{
			Key: "room_id",
			Value: req.GetRoomId(),
		},
	}
	result := c.roomColl.FindOne(ctx, query)
	if result.Err() != nil {
		return nil , result.Err()
	}
	resp := new(pb.Room)
	err := result.Decode(resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c ChatServer) UpdateRoom(ctx context.Context, room *pb.Room) (*pb.Room, error) {
	query := bson.D{
		{
			Key: "room_id",
			Value: room.GetRoomId(),
		},
	}
	//TODO
	result := c.roomColl.FindOneAndReplace(ctx, query, room)
	if result.Err() != nil {
		return nil, result.Err()
	}
	return room, nil
}

func (c ChatServer) DeleteRoom(ctx context.Context, req *pb.DeleteRoomReq) (*pb.DeleteRoomResp, error) {
	query := bson.D{
		{
			Key: "room_id",
			Value: req.GetRoomId(),
		},
	}
	result := c.roomColl.FindOneAndDelete(ctx, query)
	if result.Err() != nil {
		return nil, result.Err()
	}
	return &pb.DeleteRoomResp{IsDeleted: true}, nil
}

func (c ChatServer) CreateRoom(ctx context.Context, room *pb.Room) (*pb.Room, error) {
	// check if already preset
	query := bson.D{
		{
			Key: "room_id",
			Value: room.GetRoomId(),
		},
	}
	r := c.roomColl.FindOne(ctx, query)
	if r.Err() != nil {
		if r.Err() == mongo.ErrNoDocuments {
			//insert new room
			_, err := c.roomColl.InsertOne(ctx, room)
			if err != nil {
				return nil, err
			}
			return room, nil
		}
	}
	return nil, status.Error(codes.AlreadyExists, "Room already exist")
}












//User CURDL
func (c ChatServer) ListUsers(ctx context.Context, req *pb.ListUserReq) (*pb.ListUserResp, error) {
	//get the room id from where you want all the user
	if(len(req.GetRoomId()) == 0 ){
		return nil, status.Error(codes.InvalidArgument, "Room Id not provided")
	}
	query := bson.D{
		{
			Key: "room_id",
			Value: req.GetRoomId(),
		},
	}
	cur , err := c.userColl.Find(ctx, query)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	resp := new(pb.ListUserResp)
	for cur.Next(ctx) {
		user := new(pb.User)
		err = cur.Decode(user)
		if err != nil {
			return nil, err
		}
		resp.Users = append(resp.Users, user)
	}
	return resp, nil
}

func (c ChatServer) GetUser(ctx context.Context, req *pb.GetUserReq) (*pb.User, error) {
	query := bson.D{
		{
			Key: "user_id",
			Value: req.GetUserId(),
		},
	}
	r := c.userColl.FindOne(ctx, query)
	if r.Err() != nil {
		if r.Err() == mongo.ErrNoDocuments {
			return nil, status.Error(codes.NotFound, "User not found")
		}
		return nil, r.Err()
	}
	user := new(pb.User)
	err := r.Decode(user)
	if err != nil {
		return nil, err
	}
	return user, err
}

func (c ChatServer) UpdateUser(ctx context.Context, user *pb.User) (*pb.User, error) {
	query := bson.D{
		{
			Key: "user_id",
			Value: user.GetUserId(),
		},
	}
	result := c.userColl.FindOneAndReplace(ctx, query, user)
	if result.Err() != nil {
		return nil, result.Err()
	}
	return user, nil
}

func (c ChatServer) DeleteUser(ctx context.Context, req *pb.DeleteUserReq) (*pb.DeleteUserResp, error) {
	query := bson.D{
		{
			Key: "user_id",
			Value: req.GetUserId(),
		},
	}
	result := c.userColl.FindOneAndDelete(ctx, query)
	if result.Err() != nil {
		return nil, result.Err()
	}
	return &pb.DeleteUserResp{IsDeleted: true}, nil
}

func (c ChatServer) CreateUser(ctx context.Context, user *pb.User) (*pb.User, error) {
	// check if already preset
	query := bson.D{
		{
			Key: "user_id",
			Value: user.GetUserId(),
		},
	}
	r := c.userColl.FindOne(ctx, query)
	if r.Err() != nil {
		if r.Err() == mongo.ErrNoDocuments {
			//insert new room
			_, err := c.userColl.InsertOne(ctx, user)
			if err != nil {
				return nil, err
			}
			return user, err
		}
	}
	return nil, status.Error(codes.AlreadyExists, "User already exist")
}


//Message CURDL
func (c ChatServer) ListMessages(ctx context.Context, req *pb.ListMsgReq) (*pb.ListMsgResp, error) {
	//if room Id is provided
	var query bson.D
	if len(req.GetRoomId()) > 0 {
		query = bson.D{
			{
				Key: "room_id",
				Value: req.GetRoomId(),
			},
		}
	}else if len(req.GetUserId()) > 0 {
		query = bson.D{
			{
				Key: "user_id",
				Value: req.GetUserId(),
			},
		}
	}
	cur , err := c.msgColl.Find(ctx, query)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	resp := new(pb.ListMsgResp)
	for cur.Next(ctx) {
		msg := new(pb.Message)
		err = cur.Decode(msg)
		if err != nil {
			return nil, err
		}
		resp.Messages = append(resp.Messages, msg)
	}
	return resp, nil
}

func (c ChatServer) GetMsg(ctx context.Context, req *pb.GetMsgReq) (*pb.Message, error) {
	q := bson.D{
		{
			Key: "msg_id",
			Value: req.GetMsgId(),
		},
	}
	r := c.msgColl.FindOne(ctx, q)
	if r.Err() != nil {
		return nil, r.Err()
	}
	resp := new(pb.Message)
	err := r.Decode(resp)
	if err != nil {
		return nil, err
	}
	return resp , nil
}

func (c ChatServer) UpdateMsg(ctx context.Context, message *pb.Message) (*pb.Message, error) {
	q := bson.D{
		{
			Key: "msg_id",
			Value: message.GetMsgId(),
		},
	}
	r := c.msgColl.FindOneAndReplace(ctx, q, message)
	if r.Err() != nil {
		return nil, r.Err()
	}
	return message, nil
}

func (c ChatServer) DeleteMsg(ctx context.Context, req *pb.DeleteMsgReq) (*pb.DeleteMsgResp, error) {
	q := bson.D{
		{
			Key: "msg_id",
			Value: req.GetMsgId(),
		},
	}
	r := c.msgColl.FindOneAndDelete(ctx, q)
	if r.Err() != nil {
		return nil, r.Err()
	}
	return &pb.DeleteMsgResp{IsDeleted: true}, nil
}

func (c ChatServer) CreateMsg(ctx context.Context, msg *pb.Message) (*pb.Message, error) {
	q := bson.D{
		{
			Key: "msg_id",
			Value: msg.GetMsgId(),
		},
	}
	r := c.msgColl.FindOne(ctx, q)
	if r.Err() != nil {
		if r.Err() == mongo.ErrNoDocuments {
			//insert new msg
			_, err := c.msgColl.InsertOne(ctx, msg)
			if err != nil {
				return nil, err
			}
			return msg, err
		}
		return nil, r.Err()
	}
	return nil, status.Error(codes.AlreadyExists, "msg already present with msg id "+msg.GetMsgId())
}




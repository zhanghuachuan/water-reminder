package trpcservices

import (
	"context"

	"github.com/zhanghuachuan/water-reminder/api/proto"
	"github.com/zhanghuachuan/water-reminder/internal/services"
	"trpc.group/trpc-go/trpc-go/server"
)

type WaterRecordService struct{}

func RegisterWaterRecordService(s server.Service, svr *WaterRecordService) {
	proto.RegisterWaterRecordServiceService(s, svr)
}

func (s *WaterRecordService) CreateRecord(ctx context.Context, req *proto.CreateRecordRequest) (*proto.CreateRecordResponse, error) {
	record, err := services.CreateWaterRecord(uint(req.UserId), req.Amount, req.DrinkType)
	if err != nil {
		return nil, err
	}

	return &proto.CreateRecordResponse{
		RecordId:  uint32(record.ID),
		Amount:    record.Amount,
		DrinkType: record.DrinkType,
		Timestamp: record.Time,
	}, nil
}

func (s *WaterRecordService) GetRecords(ctx context.Context, req *proto.GetRecordsRequest) (*proto.GetRecordsResponse, error) {
	records, err := services.GetUserRecords(uint(req.UserId))
	if err != nil {
		return nil, err
	}

	var respRecords []*proto.Record
	for _, r := range records {
		respRecords = append(respRecords, &proto.Record{
			Id:        uint32(r.ID),
			Amount:    r.Amount,
			DrinkType: r.DrinkType,
			Timestamp: r.Time,
		})
	}

	return &proto.GetRecordsResponse{Records: respRecords}, nil
}

func (s *WaterRecordService) GetTodayIntake(ctx context.Context, req *proto.GetTodayIntakeRequest) (*proto.GetTodayIntakeResponse, error) {
	total, err := services.GetTodayIntake(uint(req.UserId))
	if err != nil {
		return nil, err
	}

	return &proto.GetTodayIntakeResponse{TotalAmount: total}, nil
}
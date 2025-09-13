package trpcservices

import (
	"context"

	"github.com/zhanghuachuan/water-reminder/api/proto"
	"github.com/zhanghuachuan/water-reminder/internal/services"
	"trpc.group/trpc-go/trpc-go/server"
)

type WaterRecordService struct {
	proto.UnimplementedWaterRecordServiceServer
}

func RegisterWaterRecordService(s server.Service, svr *WaterRecordService) {
	s.Register(&server.ServiceDesc{
		ServiceName: "water_reminder.water_record.WaterRecordService",
		HandlerType: (*WaterRecordService)(nil),
		Methods: []server.Method{
			{
				Name: "CreateRecord",
				Func: svr.CreateRecord,
			},
			{
				Name: "GetRecords",
				Func: svr.GetRecords,
			},
			{
				Name: "GetTodayIntake",
				Func: svr.GetTodayIntake,
			},
		},
	}, svr)
}

func (s *WaterRecordService) CreateRecord(svr interface{}, ctx context.Context, f server.FilterFunc) (interface{}, error) {
	req := &proto.CreateRecordRequest{}
	_, err := f(req)
	if err != nil {
		return nil, err
	}
	record, err := services.CreateWaterRecord(uint(req.UserId), req.Amount, req.DrinkType)
	if err != nil {
		return nil, err
	}

	resp := &proto.CreateRecordResponse{
		RecordId:  uint32(record.ID),
		Amount:    record.Amount,
		DrinkType: record.DrinkType,
		Timestamp: record.Time,
	}
	return resp, nil
}

func (s *WaterRecordService) GetRecords(svr interface{}, ctx context.Context, f server.FilterFunc) (interface{}, error) {
	req := &proto.GetRecordsRequest{}
	_, err := f(req)
	if err != nil {
		return nil, err
	}
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

	resp := &proto.GetRecordsResponse{Records: respRecords}
	return resp, nil
}

func (s *WaterRecordService) mustEmbedUnimplementedWaterRecordServiceServer() {}

func (s *WaterRecordService) GetTodayIntake(svr interface{}, ctx context.Context, f server.FilterFunc) (interface{}, error) {
	req := &proto.GetTodayIntakeRequest{}
	_, err := f(req)
	if err != nil {
		return nil, err
	}
	total, err := services.GetTodayIntake(uint(req.UserId))
	if err != nil {
		return nil, err
	}

	resp := &proto.GetTodayIntakeResponse{TotalAmount: total}
	return resp, nil
}

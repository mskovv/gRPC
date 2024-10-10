package grpc_service

import (
	"context"
	pb "coursera/hw7_microservice/gen"
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

type EventLogger interface {
	LogEvent(event *pb.Event)
	Subscribe() chan *pb.Event
	Unsubscribe(chan *pb.Event)
}

type StatsCounter interface {
	UpdateStatistics(consumer, method string)
	GetStats() (map[string]uint64, map[string]uint64)
}

type StatsNotifier interface {
	NotifyAll(*pb.Stat)
	Subscribe() chan *pb.Stat
	Unsubscribe(chan *pb.Stat)
}

type ACLChecker interface {
	CheckAccess(consumer, method string) bool
}

type AdminServer struct {
	pb.UnimplementedAdminServer
	Logger        EventLogger
	Counter       StatsCounter
	StatsNotifier StatsNotifier
	Acl           ACLChecker
	Ctx           context.Context
}

func (s *AdminServer) Logging(_ *pb.Nothing, stream pb.Admin_LoggingServer) error {
	method := "/main.Admin/Logging"
	consumer := GetConsumerFromContext(stream.Context())

	if !s.Acl.CheckAccess(consumer, method) {
		return status.Errorf(codes.Unauthenticated, "invalid token")
	}
	s.Logger.LogEvent(
		&pb.Event{
			Consumer: consumer,
			Method:   method,
			Host:     "127.0.0.1:8082",
		})

	logCh := s.Logger.Subscribe()
	defer s.Logger.Unsubscribe(logCh)

	for {
		select {
		case event := <-logCh:
			if err := stream.Send(event); err != nil {
				return err
			}
		case <-s.Ctx.Done():
			s.Logger.Unsubscribe(logCh)
			return stream.Context().Err()
		}
	}
}

/*
1. Определяем консьюмера из контекста
2. Проверяем его доступ к этому методу
3. Слушатель подписывается на уведомления по статистике
4. Обновляем статистику, хоть мы и не ожидаем в тестах, что информация о первом слушателе появится в статистике,
но второй слушатель у нас появляется через секунду
*/

func (s *AdminServer) Statistics(interval *pb.StatInterval, stream pb.Admin_StatisticsServer) error {
	method := "/main.Admin/Statistics"
	var consumer string
	consumer = GetConsumerFromContext(stream.Context())
	fmt.Println(consumer)

	if !s.Acl.CheckAccess(consumer, method) {
		return status.Errorf(codes.Unauthenticated, "invalid token")
	}

	//s.Counter.UpdateStatistics(consumer, method)
	statCh := s.StatsNotifier.Subscribe()
	defer s.StatsNotifier.Unsubscribe(statCh)

	ticker := time.NewTicker(time.Duration(interval.IntervalSeconds) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case stat := <-statCh:
			if err := stream.Send(stat); err != nil {
				return err
			}
		case <-ticker.C:
			s.Counter.UpdateStatistics(consumer, method)
			methodStats, consumerStats := s.Counter.GetStats()
			stats := &pb.Stat{
				Timestamp:  time.Now().Unix(),
				ByMethod:   methodStats,
				ByConsumer: consumerStats,
			}
			s.StatsNotifier.NotifyAll(stats)
		case <-s.Ctx.Done():
			return stream.Context().Err()
		}
	}
}

func (s *AdminServer) mustEmbedUnimplementedAdminServer() {
	return
}

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
	AddInStatistics(consumer, method string)
	GetStats() *SimpleStatsCounter
	ClearStat()
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

	if !s.Acl.CheckAccess(consumer, method) {
		return status.Errorf(codes.Unauthenticated, "invalid token")
	}

	// Слушатель подписывается на прослушку канала статистки.
	// Его добавляют в мапку слушателей и получает канал.
	statCh := s.StatsNotifier.Subscribe()
	//Отложенный вызов отписки слушателя
	defer s.StatsNotifier.Unsubscribe(statCh)

	// Создаем тиккер на время, которое передал нам слушатель
	ticker := time.NewTicker(time.Duration(interval.IntervalSeconds) * time.Second)
	defer ticker.Stop()

	for {
		select {
		// Должны что-то делать при получении статистики новой
		case <-statCh:
			fmt.Println("Get form statCh")
			// Добавляем в статистику новую запись о слушателе и методе
			s.Counter.AddInStatistics(consumer, method)

		//	С каждым тиком должны ОТПРАВЛЯТЬ статистику
		case <-ticker.C:
			// Получаем всю статистику из счетчиков в виде мапок
			stats := s.Counter.GetStats()

			//Создаем объект статистики, которую будем отправлять
			stat := &pb.Stat{
				Timestamp:  time.Now().Unix(),
				ByMethod:   stats.methodCount,
				ByConsumer: stats.consumerCount,
			}
			//NotifyAll проходится по всем слушателям из мапки и отправляет им в канал статистику.
			s.StatsNotifier.NotifyAll(stat)
			fmt.Println("отправляем статистику слушателям")

			fmt.Println("channel")
			if err := stream.Send(stat); err != nil {
				return err
			}
			fmt.Println("stat before clear", stat)
			s.Counter.ClearStat()
			fmt.Println("stat after clear", stat)
		case <-s.Ctx.Done():
			//Отложенный вызов отписки слушателя
			s.StatsNotifier.Unsubscribe(statCh)
			s.Counter.ClearStat()
			return stream.Context().Err()
		}
	}
}

func (s *AdminServer) mustEmbedUnimplementedAdminServer() {
	return
}

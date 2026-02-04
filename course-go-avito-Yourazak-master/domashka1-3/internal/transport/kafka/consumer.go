package kafka

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	"avito-courier/internal/gateway/order"
	"avito-courier/internal/model"
	"avito-courier/internal/usecase"

	"github.com/IBM/sarama"
)

type Consumer struct {
	ready        chan bool
	factory      *usecase.EventHandlerFactory
	orderGateway order.OrderGateway
}

func NewConsumer(factory *usecase.EventHandlerFactory, gateway order.OrderGateway) *Consumer {
	return &Consumer{
		ready:        make(chan bool),
		factory:      factory,
		orderGateway: gateway,
	}
}

func (c *Consumer) StartConsumerGroup(ctx context.Context, brokers []string, groupID string, topics []string) {
	config := sarama.NewConfig()
	config.Version = sarama.V2_5_0_0
	config.Consumer.Offsets.Initial = sarama.OffsetNewest
	config.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRange()
	config.Consumer.Return.Errors = true

	consumerGroup, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		log.Printf("Failed to create consumer group: %v", err)
		return
	}
	defer consumerGroup.Close()

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for err := range consumerGroup.Errors() {
			log.Printf("Consumer group error: %v", err)
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				if err := consumerGroup.Consume(ctx, topics, c); err != nil {
					log.Printf("Error from consumer: %v", err)
				}
			}
		}
	}()

	<-c.ready
	log.Printf("Kafka consumer group started: topics=%v, group=%s", topics, groupID)

	<-ctx.Done()
	log.Println("Kafka consumer stopped")
	wg.Wait()
}

func (c *Consumer) Setup(sarama.ConsumerGroupSession) error {
	close(c.ready)
	return nil
}

func (c *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (c *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		c.processMessage(session.Context(), message)
		session.MarkMessage(message, "")
	}
	return nil
}

func (c *Consumer) processMessage(ctx context.Context, msg *sarama.ConsumerMessage) {
	var event model.OrderEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		log.Printf("Failed to unmarshal Kafka event: %v", err)
		return
	}

	if !event.Validate() {
		log.Printf("Invalid Kafka event: %+v", event)
		return
	}

	log.Printf("Kafka event received: %s - %s (offset: %d)",
		event.OrderID, event.Status, msg.Offset)

	if c.orderGateway != nil {
		actualOrder, err := c.orderGateway.GetOrderStatus(ctx, event.OrderID)
		if err != nil {
			log.Printf("Failed to verify order status for %s: %v", event.OrderID, err)
		} else if actualOrder.Status != event.Status {
			log.Printf("Status mismatch for %s: event=%s, actual=%s - skipping",
				event.OrderID, event.Status, actualOrder.Status)
			return
		}
	}

	handler := c.factory.GetHandler(event.Status)
	if handler == nil {
		log.Printf("No handler for status: %s", event.Status)
		return
	}

	if err := handler.Handle(ctx, event); err != nil {
		log.Printf("Failed to handle event %s: %v", event.OrderID, err)
	} else {
		log.Printf("Event processed: %s - %s", event.OrderID, event.Status)
	}
}

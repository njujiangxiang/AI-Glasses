package events

import (
	"time"

	"aiglasses/server/internal/platform/database"
	amqp "github.com/rabbitmq/amqp091-go"
	"gorm.io/gorm"
)

type Publisher struct {
	db  *gorm.DB
	url string
}

// NewPublisher 创建 outbox 事件发布器。
func NewPublisher(db *gorm.DB, url string) *Publisher { return &Publisher{db: db, url: url} }

// Flush 批量发布待处理 outbox 事件，当前 MVP 保留数据库幂等处理边界。
func (p *Publisher) Flush(limit int) error {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	conn, err := amqp.Dial(p.url)
	if err != nil {
		return err
	}
	defer conn.Close()
	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()
	var events []database.OutboxEvent
	if err := p.db.Where("published_at IS NULL").Order("id asc").Limit(limit).Find(&events).Error; err != nil {
		return err
	}
	for _, event := range events {
		if err := ch.Publish("", event.Topic, false, false, amqp.Publishing{ContentType: "application/json", Body: []byte(event.Payload), MessageId: event.EventKey}); err != nil {
			return err
		}
		now := time.Now().UTC()
		if err := p.db.Model(&event).Update("published_at", now).Error; err != nil {
			return err
		}
	}
	return nil
}

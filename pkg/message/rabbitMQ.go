package message

import (
	"fmt"
	"github.com/streadway/amqp"
	"log"
)

const MQURL = "amqp://admin:123456@127.0.0.1:5672/"

type RabbitMQ struct {
	conn      *amqp.Connection
	channel   *amqp.Channel
	QueueName string
	Mqurl     string
}

type consumeHandler func(content []byte)

func (r *RabbitMQ) Destroy() {
	err := r.channel.Close()
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}
	err = r.conn.Close()
	if err != nil {
		fmt.Printf("%v\n", err)
	}
}

// error handler.
func (r *RabbitMQ) failOnErr(err error, message string) {
	if err != nil {
		log.Fatalf("%s:%s", message, err)
	}
}

func NewRabbitMQ(queueName string) *RabbitMQ {
	rabbitmq := &RabbitMQ{QueueName: queueName, Mqurl: MQURL}
	var err error
	rabbitmq.conn, err = amqp.Dial(rabbitmq.Mqurl)
	rabbitmq.failOnErr(err, "Failed to connect rabbitmq!")
	rabbitmq.channel, err = rabbitmq.conn.Channel()
	rabbitmq.failOnErr(err, "Failed to open a channel")
	return rabbitmq
}

func (r *RabbitMQ) PublishSimple(message string) {
	_, err := r.channel.QueueDeclare(
		r.QueueName,
		//是否持久化
		false,
		//是否自动删除
		false,
		//是否具有排他性
		false,
		//是否阻塞处理
		false,
		//额外的属性
		nil,
	)
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}
	//调用channel 发送消息到队列中
	err = r.channel.Publish(
		"",
		r.QueueName,
		//如果为true，根据自身exchange类型和routekey规则无法找到符合条件的队列会把消息返还给发送者
		false,
		//如果为true，当exchange发送消息到队列后发现队列上没有消费者，则会把消息返还给发送者
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        []byte(message),
		})
}

func (r *RabbitMQ) ConsumeSimple(handler consumeHandler) {
	q, err := r.channel.QueueDeclare(
		r.QueueName,
		//是否持久化
		false,
		//是否自动删除
		false,
		//是否具有排他性
		false,
		//是否阻塞处理
		false,
		//额外的属性
		nil,
	)
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	//接收消息
	msgs, err := r.channel.Consume(
		q.Name, // queue
		//用来区分多个消费者
		"", // consumer
		//是否自动应答
		true, // auto-ack
		//是否独有
		false, // exclusive
		//设置为true，表示 不能将同一个Conenction中生产者发送的消息传递给这个Connection中 的消费者
		false, // no-local
		//列是否阻塞
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	forever := make(chan bool)
	//启用协程处理消息
	go func() {
		for d := range msgs {
			log.Printf("[INFO]Receive a message:\n%s", d.Body)
			handler(d.Body)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}

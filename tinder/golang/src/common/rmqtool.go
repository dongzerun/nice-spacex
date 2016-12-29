package common

import (
	"fmt"
	"github.com/streadway/amqp"
	"strings"
)

type MQPool struct {
	etype           string
	exchange        string
	routeKeys       []string
	queueName       string
	ConsumeParallel int
	conn            *amqp.Connection
	channel         *amqp.Channel
}

func NewMQPool(conf string, mq string) (mqp *MQPool, err error) {

	config, err := NewConfig(conf)
	if err != nil {
		panic("no config file to read")
	}

	host := config.MustValue(mq, "host", "localhost")
	port := config.MustInt(mq, "port", 5672)
	username := config.MustValue(mq, "username", "guest")
	password := config.MustValue(mq, "password", "guest")
	_type := config.MustValue(mq, "type", "direct")
	_parallel := config.MustInt(mq, "consumeParallel", 1)
	_exchange := config.MustValue(mq, "exchange")
	_routeKeys := strings.Split(config.MustValue(mq, "routekeys"), ",")
	_queueName := config.MustValue(mq, "queueName")

	if host == "" || port == 0 || username == "" || password == "" || _type == "" || _exchange == "" || len(_routeKeys) == 0 || _queueName == "" {
		panic("config file not avaliable")
	}

	url := fmt.Sprintf("amqp://%s:%s@%s:%d/", username, password, host, port)

	_conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	_chan, err := _conn.Channel()
	if err != nil {
		return nil, err
	}

	return &MQPool{
		etype:           _type,
		exchange:        _exchange,
		routeKeys:       _routeKeys,
		queueName:       _queueName,
		conn:            _conn,
		channel:         _chan,
		ConsumeParallel: _parallel,
	}, err
}

func (this *MQPool) SetEx() (<-chan amqp.Delivery, error) {

	err := this.channel.ExchangeDeclare(this.exchange, this.etype, true, false, false, false, nil)
	if err != nil {
		return nil, err
	}

	q, err := this.channel.QueueDeclare(this.queueName, true, false, false, false, nil)
	if err != nil {
		return nil, err
	}

	merrors := make(chan error)
	fmt.Println(this.routeKeys)
	for _, rk := range this.routeKeys {
		err = this.channel.QueueBind(q.Name, rk, this.exchange, false, nil)
		if err != nil {
			merrors <- err
		}
	}
	if len(merrors) > 0 {
		return nil, <-merrors
	}
	msgs, err := this.channel.Consume(q.Name, "", false, false, false, false, nil)
	return msgs, err
}

func (this *MQPool) Close() {
	this.channel.Close()
	this.conn.Close()
}

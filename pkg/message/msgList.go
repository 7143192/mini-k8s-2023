package message

import "fmt"

type MsgList struct {
	list map[string]*RabbitMQ
}

func NewMsgList() *MsgList {
	return &MsgList{
		make(map[string]*RabbitMQ),
	}
}

func (l *MsgList) Put(listName, queueName string) error {
	if _, exist := l.list[listName]; exist {
		return fmt.Errorf("the key %s is in the map", listName)
	}
	q := NewRabbitMQ(queueName)
	l.list[listName] = q
	return nil
}

func (l *MsgList) Get(listName string) *RabbitMQ {
	if _, exist := l.list[listName]; exist {
		return l.list[listName]
	} else {
		return nil
	}
}

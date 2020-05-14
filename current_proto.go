package main

import (
	"os"

	"github.com/emicklei/proto"
)

type ProtoMessageMap map[string]*ProtoMessage

func BuildCurrentProtoMap(filename string) (ProtoMessageMap, error) {
	if filename == "" {
		return ProtoMessageMap{}, nil
	}
	reader, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	parser := proto.NewParser(reader)
	definition, err := parser.Parse()
	if err != nil {
		return nil, err
	}

	p := make(ProtoMessageMap)
	proto.Walk(definition, proto.WithMessage(p.HandleMessage()))

	return p, nil
}

func (p ProtoMessageMap) HandleMessage() func(m *proto.Message) {
	return func(m *proto.Message) {
		p[m.Name] = NewProtoMesssageFromMessage(m)
	}
}

func (p ProtoMessageMap) GetFieldNum(messageName, fieldName string) int {
	_, ok := p[messageName]
	if !ok {
		p[messageName] = &ProtoMessage{}
	}
	fieldNum := p[messageName].GetFieldNum(fieldName)
	return fieldNum
}

func (p ProtoMessageMap) RemoveFieldNum(messageName, fieldName string) {
	_, ok := p[messageName]
	if !ok {
		return
	}
	p[messageName].RemoveFieldNum(fieldName)
	return
}

type ProtoMessage struct {
	currMaxNum  int
	droppedNums []int
	fields      map[string]int
}

func NewProtoMesssageFromMessage(msg *proto.Message) *ProtoMessage {
	fieldsMap := make(map[string]int, len(msg.Elements))
	currMax := 0

	for i, _ := range msg.Elements {
		element := msg.Elements[i]

		switch element.(type) {
		case *proto.NormalField:
			field := element.(*proto.NormalField)
			fieldsMap[field.Name] = field.Sequence
			if field.Sequence > currMax {
				currMax = field.Sequence
			}
		case *proto.MapField:
			field := element.(*proto.MapField)
			fieldsMap[field.Name] = field.Sequence
			if field.Sequence > currMax {
				currMax = field.Sequence
			}
		default:
		}
	}
	return &ProtoMessage{
		currMaxNum: currMax,
		fields:     fieldsMap,
	}
}

func (p *ProtoMessage) GetFieldNum(fieldName string) int {
	if p.fields == nil {
		p.fields = make(map[string]int)
	}
	_, ok := p.fields[fieldName]
	if ok {
		return p.fields[fieldName]
	}
	// assumes fields are dropped in order
	if len(p.droppedNums) == 0 {
		p.currMaxNum++
		p.fields[fieldName] = p.currMaxNum
		return p.fields[fieldName]
	}

	p.fields[fieldName] = p.droppedNums[0]
	p.droppedNums = p.droppedNums[1:]
	return p.fields[fieldName]
}

func (p *ProtoMessage) RemoveFieldNum(fieldName string) {
	if p.fields == nil {
		return
	}
	// assumes fields are dropped in order
	_, ok := p.fields[fieldName]
	if !ok {
		return
	}
	p.droppedNums = append(p.droppedNums, p.fields[fieldName])
	delete(p.fields, fieldName)
	return
}

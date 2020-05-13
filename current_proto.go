package main

import (
	"fmt"
	// "io/ioutil"
	"os"

	eproto "github.com/emicklei/proto"

	// "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	// "google.golang.org/protobuf/types/pluginpb"
)

func BuildCurrentProtoMap(filename string) (ProtoMessageMap, error) {
	if filename == "" {
		return make(ProtoMessageMap), nil
	}
	descriptor, err := ReadCurrentProto(filename)
	if err != nil {
		return nil, err
	}
	return NewProtoMessageMapFromDescriptor(descriptor), nil
}

func ReadCurrentProto(filename string) (*descriptorpb.FileDescriptorProto, error) {

	reader, _ := os.Open(filename)
	defer reader.Close()

	parser := eproto.NewParser(reader)
	definition, _ := parser.Parse()

	p := make(ProtoMessageMap)
	eproto.Walk(definition,
		eproto.WithMessage(p.HandleMessage()))

	return &descriptorpb.FileDescriptorProto{}, nil
	// 	fmt.Printf("protoBytes: %v\n", string(protoBytes))
	// 	req := &pluginpb.CodeGeneratorRequest{}
	// 	err = proto.Unmarshal(protoBytes, req)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	if len(req.ProtoFile) != 1 {
	// 		return nil, fmt.Errorf("Invalid number of proto files. Expected 1. Got: %+v\n", req.ProtoFile)
	// 	}
	// 	return req.ProtoFile[0], nil
}

func handleService(s *eproto.Service) {
	fmt.Println(s.Name)
}

func (p *ProtoMessageMap) HandleMessage() func(m *eproto.Message) {
	return func(m *eproto.Message) {


		fmt.Printf("Message name: %v\n", m.Name)
		fmt.Printf("Elements: %+v\n", m.Elements)
	}
}
func handleMessage(m *eproto.Message) {

}

type ProtoMessageMap map[string]*ProtoMessage

func NewProtoMessageMapFromDescriptor(descriptor *descriptorpb.FileDescriptorProto) ProtoMessageMap {
	protoMessageMap := make(ProtoMessageMap, len(descriptor.MessageType))

	for i, _ := range descriptor.MessageType {
		message := descriptor.MessageType[i]
		protoMessageMap[message.GetName()] = NewProtoMesssageFromDescriptor(message)
	}
	return protoMessageMap
}

func (p *ProtoMessageMap) GetFieldNum(messageName, fieldName string) int32 {
	tempP := *p
	_, ok := tempP[messageName]
	if !ok {
		tempP[messageName] = NewProtoMesssageFromDescriptor(&descriptorpb.DescriptorProto{})
	}
	fieldNum := tempP[messageName].GetFieldNum(fieldName)
	p = &tempP
	return fieldNum
}

type ProtoMessage struct {
	currMaxNum int32
	fields     map[string]int32
}

func NewProtoMesssageFromMessage(msg *eproto.Message) *ProtoMessage {
	fieldsMap := make(map[string]int32, len(msg.))
	var currMax int32

	for i, _ := range descriptor.Field {
		descriptorField := descriptor.Field[i]

		fieldsMap[descriptorField.GetName()] = descriptorField.GetNumber()
		if descriptorField.GetNumber() > currMax {
			currMax = descriptorField.GetNumber()
		}
	}
	return &ProtoMessage{
		currMaxNum: currMax,
		fields:     fieldsMap,
	}
}

func (p *ProtoMessage) GetFieldNum(fieldName string) int32 {
	_, ok := p.fields[fieldName]
	if !ok {
		p.currMaxNum++
		p.fields[fieldName] = p.currMaxNum

	}

	return p.fields[fieldName]
}

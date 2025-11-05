package generator

import (
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"

	natspb "github.com/toyz/protoc-gen-nats-micro/gen/nats/micro"
)

// ServiceOptions contains metadata about a service
type ServiceOptions struct {
	SubjectPrefix string
	Name          string
	Version       string
	Description   string
}

// GetServiceOptions extracts service options from proto service definition
func GetServiceOptions(service *protogen.Service) ServiceOptions {
	// Defaults
	opts := ServiceOptions{
		Name:          ToSnakeCase(service.GoName),
		Version:       "1.0.0",
		Description:   "",
		SubjectPrefix: "",
	}

	// Try to read the nats.micro.service extension
	if service.Desc.Options() != nil && proto.HasExtension(service.Desc.Options(), natspb.E_Service) {
		ext := proto.GetExtension(service.Desc.Options(), natspb.E_Service)
		if svcOpts, ok := ext.(*natspb.ServiceOptions); ok {
			if svcOpts.SubjectPrefix != "" {
				opts.SubjectPrefix = svcOpts.SubjectPrefix
			}
			if svcOpts.Name != "" {
				opts.Name = svcOpts.Name
			}
			if svcOpts.Version != "" {
				opts.Version = svcOpts.Version
			}
			if svcOpts.Description != "" {
				opts.Description = svcOpts.Description
			}
		}
	}

	return opts
}

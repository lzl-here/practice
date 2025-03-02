package service

import (
	"context"
	"rule-engine/internal/model"
)

type ServiceClient interface {
	Process(ctx context.Context, order *model.Order) error
}

func NewServiceMap() map[string]ServiceClient {
	return map[string]ServiceClient{
		"enterprise": &EnterpriseService{},
		"course":     &CourseService{},
	}
}

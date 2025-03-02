package service

import (
	"context"
	"log"
	"rule-engine/internal/model"
)

type CourseService struct{}

func (s *EnterpriseService) Process(ctx context.Context, order *model.Order) error {
	// 调用课程开通API
	log.Printf("Enabling course access for user %d\n", order.UserID)
	// ... 具体API调用逻辑
	return nil
}

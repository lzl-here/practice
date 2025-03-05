package service

import (
	"context"
	"log"
	"rule-mapping/internal/model"
)

type EnterpriseService struct {
}

func (s *CourseService) Process(ctx context.Context, order *model.Order) error {
	// 调用课程开通API
	log.Printf("Enabling course access for user %d\n", order.UserID)
	// ... 具体API调用逻辑
	return nil
}

package services

import "github.com/th1enq/server_management_system/internal/domain/entity"

type ExcelizeService interface {
	Validate(row []string) error
	ParseToServer(row []string) (entity.Server, error)
}

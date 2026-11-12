package equipment_service

import (
	"context"

	"gitlab.com/guv41k/it_equipment_manager/internal/core/domain"
)

type EquipmentRepository interface {
	SaveEquipment(context.Context, domain.Equipment) (domain.Equipment, error)
	GetEquipmentList(context.Context, *int64, *domain.EquipmentStatus, *int64, *int, *int) ([]domain.Equipment, error)
	GetEquipment(context.Context, int64) (domain.Equipment, error)
	UpdateEquipment(context.Context, domain.Equipment) (domain.Equipment, error)
	RetireEquipment(context.Context, int64) (domain.Equipment, error)
}

type EquipmentService struct {
	equipmentRepository EquipmentRepository
}

func NewEquipmentService(repository EquipmentRepository) *EquipmentService {
	return &EquipmentService{
		equipmentRepository: repository,
	}
}

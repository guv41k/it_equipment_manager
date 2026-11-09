package assignment_postgres_repository

import (
	"context"

	"gitlab.com/guv41k/it_equipment_manager/internal/core/domain"
	core_postgres_pool "gitlab.com/guv41k/it_equipment_manager/internal/core/repository/postgres/pool"
	employee_postgres_repository "gitlab.com/guv41k/it_equipment_manager/internal/features/employee/repository/postgres"
	equipment_postgres_repository "gitlab.com/guv41k/it_equipment_manager/internal/features/equipment/repository/postgres"
)

type AssignmentsRepository struct {
	pool                core_postgres_pool.Pool
	equipmentRepository *equipment_postgres_repository.EquipmentRepository
	employeesRepository *employee_postgres_repository.EmployeesRepository
}

func NewAssignmentsRepository(pool core_postgres_pool.Pool) *AssignmentsRepository {
	return &AssignmentsRepository{
		pool:                pool,
		equipmentRepository: equipment_postgres_repository.NewEquipmentRepository(pool),
		employeesRepository: employee_postgres_repository.NewEmployeesRepository(pool),
	}
}

func (r *AssignmentsRepository) GetEquipment(ctx context.Context, id int64) (domain.Equipment, error) {
	return r.equipmentRepository.GetEquipment(ctx, id)
}

func (r *AssignmentsRepository) GetEmployee(ctx context.Context, id int64) (domain.Employee, error) {
	return r.employeesRepository.GetEmployee(ctx, id)
}

package employee_postgres_repository

import core_postgres_pool "gitlab.com/guv41k/it_equipment_manager/internal/core/repository/postgres/pool"

type EmployeesRepository struct {
	pool core_postgres_pool.Pool
}

func NewEmployeesRepository(pool core_postgres_pool.Pool) *EmployeesRepository {
	return &EmployeesRepository{
		pool: pool,
	}
}

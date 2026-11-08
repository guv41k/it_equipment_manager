package equipment_postgres_repository

import core_postgres_pool "gitlab.com/guv41k/it_equipment_manager/internal/core/repository/postgres/pool"

type EquipmentRepository struct {
	pool core_postgres_pool.Pool
}

func NewEquipmentRepository(pool core_postgres_pool.Pool) *EquipmentRepository {
	return &EquipmentRepository{
		pool: pool,
	}
}

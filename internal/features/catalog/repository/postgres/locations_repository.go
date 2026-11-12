package catalog_postgres_repository

import core_postgres_pool "gitlab.com/guv41k/it_equipment_manager/internal/core/repository/postgres/pool"

type LocationsRepository struct {
	pool core_postgres_pool.Pool
}

func NewLocationsRepository(pool core_postgres_pool.Pool) *LocationsRepository {
	return &LocationsRepository{
		pool: pool,
	}
}

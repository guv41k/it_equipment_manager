package catalog_postgres_repository

import core_postgres_pool "gitlab.com/guv41k/it_equipment_manager/internal/core/repository/postgres/pool"

type CategoriesRepository struct {
	pool core_postgres_pool.Pool
}

func NewCategoriesRepository(pool core_postgres_pool.Pool) *CategoriesRepository {
	return &CategoriesRepository{
		pool: pool,
	}
}

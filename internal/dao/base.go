package dao

import (
	"sync/atomic"
	"time"

	"gorm.io/gorm"
)

type DB = gorm.DB

type Dao struct {
	db *DB

	location atomic.Pointer[time.Location]

	// Cache to speed up database query
	//
	// Please access the cache instance by invoking the `CacheAction` func.
	//
	// As the cache could be nil.
	// Therefore, it is necessary to check if the cache is not nil before referencing it.
	cache *DaoCache
}

// Create new dao instance
//
// This function will auto migrate database tables
func NewDao(db *DB) *Dao {
	dao := &Dao{
		db: db,
	}

	dao.MigrateModels()

	return dao
}

func (dao *Dao) DB() *DB {
	return dao.db
}

func (dao *Dao) SetLocation(location *time.Location) {
	if location == nil {
		location = time.Local
	}
	dao.location.Store(location)
}

func (dao *Dao) Location() *time.Location {
	if location := dao.location.Load(); location != nil {
		return location
	}
	return time.Local
}

func (dao *Dao) SetCache(cache *DaoCache) {
	dao.cache = cache
}

func (dao *Dao) CacheAction(fn func(cache *DaoCache)) {
	if dao.cache != nil {
		fn(dao.cache)
	}
}

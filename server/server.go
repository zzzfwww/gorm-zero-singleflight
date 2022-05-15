package server

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"gorm-test/syncx"
	"gorm.io/gorm"
	"log"
)

var (
	cachePrefix = "cache:id:"
)

// QueryCtxFn defines the query method.
type QueryCtxFn func(ctx context.Context, conn *gorm.DB, v interface{}) error

type GetService struct {
	mysqlClient *gorm.DB
	redisClient *redis.Client
	barrier     syncx.SingleFlight
}

func NewGetService(m *gorm.DB, r *redis.Client, b syncx.SingleFlight) *GetService {
	return &GetService{
		mysqlClient: m,
		redisClient: r,
		barrier:     b,
	}
}

//根据主键查询一条数据，走缓存
func (s *GetService) FindOne(ctx context.Context, id int64) (*ThirdPayment, error) {
	key := fmt.Sprintf("%v%v", cachePrefix, id)
	var resp ThirdPayment
	err := s.QueryRowCtx(ctx, &resp, key, func(ctx context.Context, conn *gorm.DB, v interface{}) error {
		return conn.WithContext(ctx).Model(&ThirdPayment{}).Where("id=?", id).Find(&v).Error
	})
	switch err {
	case nil:
		return &resp, nil
	case gorm.ErrRecordNotFound:
		return nil, gorm.ErrRecordNotFound
	default:
		return nil, err
	}
}

// QueryRowCtx unmarshals into v with given key and query func.
func (s GetService) QueryRowCtx(ctx context.Context, v interface{}, key string, query QueryCtxFn) error {
	return s.TakeCtx(ctx, v, key, func(v interface{}) error {
		return query(ctx, s.mysqlClient, v)
	})
}

func (s GetService) TakeCtx(ctx context.Context, val interface{}, key string, query func(val interface{}) error) error {
	return s.doTake(ctx, val, key, query, func(v interface{}) error {
		data, err := json.Marshal(v)
		if err != nil {
			return err
		}
		_, err = s.redisClient.Set(ctx, key, data, 0).Result()
		return err
	})
}

func (s GetService) doTake(ctx context.Context, v interface{}, key string,
	query func(v interface{}) error, cacheVal func(v interface{}) error) error {
	val, fresh, err := s.barrier.DoEx(key, func() (interface{}, error) {
		if err := s.doGetCache(ctx, key, v); err != nil {

			if err = query(v); err == gorm.ErrRecordNotFound {
				return nil, gorm.ErrRecordNotFound
			} else if err != nil {
				return nil, err
			}

			if err = cacheVal(v); err != nil {
				log.Println(err.Error())
			}
		}

		return json.Marshal(v)
	})
	if err != nil {
		return err
	}
	if fresh {
		return nil
	}

	// got the result from previous ongoing query
	// c.stat.IncrementTotal()
	// c.stat.IncrementHit()

	return json.Unmarshal(val.([]byte), v)
}

func (s GetService) doGetCache(ctx context.Context, key string, v interface{}) error {
	data, err := s.redisClient.Get(ctx, key).Result()
	if err != nil {
		return err
	}

	if len(data) == 0 {
		return gorm.ErrRecordNotFound
	}

	return s.processCache(ctx, key, data, v)
}

func (s GetService) processCache(ctx context.Context, key, data string, v interface{}) error {
	err := json.Unmarshal([]byte(data), v)
	if err == nil {
		return nil
	}
	return gorm.ErrRecordNotFound
}

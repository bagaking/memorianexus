package tags

import (
	"context"
	"time"

	"github.com/adjust/redismq"

	"github.com/bagaking/goulp/wlog"
	"github.com/bagaking/memorianexus/internal/utils"
	"github.com/bagaking/memorianexus/internal/utils/cache"
	"github.com/khgame/memstore/cachekey"

	"github.com/khicago/irr"
)

const (
	MaxRetryAttempts      = 3
	TagCacheExpiration    = 24 * time.Hour
	CacheInvalidationFlag = "invalid"
)

type (
	// ParamUserTag represents the parameters for user tag.
	ParamUserTag struct {
		UserID utils.UInt64
		Tag    string
	}

	// ParamUserTagType represents the parameters for user tag with entity type.
	ParamUserTagType[EntityType any] struct {
		UserID utils.UInt64 `cachekey:"uid"`
		Tag    string       `cachekey:"tag"`
		Type   EntityType   `cachekey:"entity_type"`
	}

	// TagCKs represents the cache keys for tags.
	TagCKs[EntityType any] struct {
		User2Tags   *cachekey.KeySchema[utils.UInt64]
		Tag2Users   *cachekey.KeySchema[string]
		Entity2Tags *cachekey.KeySchema[utils.UInt64]
		Entities    *cachekey.KeySchema[ParamUserTagType[EntityType]]
	}

	// Tag represents the tag model.
	Tag[EntityType any] struct {
		UserID     utils.UInt64
		Tag        string
		EntityType EntityType
		EntityID   utils.UInt64
		CreatedAt  time.Time
		UpdatedAt  time.Time
	}

	// TagService represents the tag service.
	TagService[EntityType any] struct {
		repo TagRepository[EntityType]

		Schemas        TagCKs[EntityType]
		supportedTypes []EntityType

		UpdateMgr *TagUpdateManager[EntityType]
	}

	// DirtyEvent represents the tag event type.
	DirtyEvent string

	// TagUpdateMessage represents the message for tag update.
	TagUpdateMessage[EntityType any] struct {
		Action     DirtyEvent   `json:"action"`
		UserID     utils.UInt64 `json:"uid"`
		EntityID   utils.UInt64 `json:"eid"`
		EntityType EntityType   `json:"entity_type"`
		Tag        string       `json:"tags"`
		Propagate  bool         `json:"propagate"`
	}

	TagSvr[EntityType any] interface {
		// -- 查询接口

		GetTagsByUser(ctx context.Context, userID utils.UInt64) ([]string, error)
		GetUsersByTag(ctx context.Context, tag string) ([]utils.UInt64, error)
		GetEntities(ctx context.Context, userID utils.UInt64, tag string, entityType *EntityType) ([]utils.UInt64, error)
		GetTagsOfEntity(ctx context.Context, entityID utils.UInt64) ([]string, error)

		// -- 标脏接口

		InvalidateUserCache(ctx context.Context, userID utils.UInt64, propagate bool) error
		InvalidateTagCache(ctx context.Context, tag string, propagate bool) error
		InvalidateUserTagCache(ctx context.Context, userID utils.UInt64, tag string, propagate bool) error
		InvalidateEntityCache(ctx context.Context, entityID utils.UInt64, propagate bool) error
	}
)

const (
	EventInvalidUser   DirtyEvent = "dirty_user"
	EventInvalidTag    DirtyEvent = "dirty_tag"
	EventInvalidEntity DirtyEvent = "dirty_entity"
)

var (
	_            TagSvr[any] = &TagService[any]{}
	ErrCacheMiss             = irr.Error("cache miss")
)

// NewTagService creates a new TagService.
func NewTagService[EntityType any](
	ctx context.Context,
	repo TagRepository[EntityType],
	supportedTypes []EntityType,
	producer Producer, consumer Consumer[*redismq.Package],
) *TagService[EntityType] {
	svr := &TagService[EntityType]{repo: repo}
	svr.Schemas = TagCKs[EntityType]{
		User2Tags:   cachekey.MustNewSchema[utils.UInt64]("users:{uid}:tags", TagCacheExpiration),
		Tag2Users:   cachekey.MustNewSchema[string]("tags:{tag}:users", TagCacheExpiration),
		Entity2Tags: cachekey.MustNewSchema[utils.UInt64]("entities:{eid}:tags", TagCacheExpiration),
		Entities:    cachekey.MustNewSchema[ParamUserTagType[EntityType]]("users:{uid}:tags:{tag}:{entity_type}", TagCacheExpiration),
	}
	svr.supportedTypes = supportedTypes
	svr.UpdateMgr = NewTagUpdateManager[EntityType](producer, consumer, svr.handleTagUpdateMessage).Start(ctx)
	return svr
}

// GetTagsByUser 查询接口实现
func (s *TagService[EntityType]) GetTagsByUser(ctx context.Context, userID utils.UInt64) ([]string, error) {
	ll, ctx := wlog.ByCtxAndCache(ctx, "GetTagsByUser")
	log := ll.WithField("userID", userID)
	cacheKey := s.Schemas.User2Tags.MustBuild(userID)
	tags, err := cache.SET().GetAll(ctx, cacheKey)
	if err == nil && len(tags) > 0 && tags[0] != CacheInvalidationFlag {
		return tags, nil
	}

	// 获取锁
	locker := cache.Locker(ctx)
	lockerVal, err := locker.Acquire(ctx, cacheKey, time.Second*10)
	if err != nil {
		log.Errorf("failed to acquire lock for cacheKey %s, err= %v", cacheKey, err)
		return nil, err
	}
	defer func() {
		if err = locker.Release(ctx, cacheKey, lockerVal); err != nil {
			log.Warnf("failed to release the lock of cacheKey %s, err= %v", cacheKey, err)
		}
	}()

	// Double check to avoid cache stampede
	tags, err = cache.SET().GetAll(ctx, cacheKey)
	if err == nil && len(tags) > 0 && tags[0] != CacheInvalidationFlag {
		return tags, nil
	}

	tags, err = s.repo.GetTagsByUser(ctx, userID)
	if err != nil {
		log.Errorf("failed to fetch tags from repo for userID %d, err= %v", userID, err)
		return nil, err
	}

	if err = cache.SET().Insert(ctx, cacheKey, TagCacheExpiration, tags...); err != nil {
		log.Warnf("failed to update the string list cache, err= %v", err)
	}
	return tags, nil
}

func (s *TagService[EntityType]) GetUsersByTag(ctx context.Context, tag string) ([]utils.UInt64, error) {
	ll, ctx := wlog.ByCtxAndCache(ctx, "GetUsersByTag")
	log := ll.WithField("tag", tag)
	cacheKey := s.Schemas.Tag2Users.MustBuild(tag)
	users, err := cache.SET().GetAllUInt64s(ctx, cacheKey)
	if err == nil && len(users) > 0 && users[0] != utils.UInt64(0) {
		return users, nil
	}

	// 获取锁
	locker := cache.Locker(ctx)
	lockerVal, err := locker.Acquire(ctx, cacheKey, time.Second*10)
	if err != nil {
		log.Errorf("failed to acquire lock for cacheKey %s, err= %v", cacheKey, err)
		return nil, err
	}
	defer func() {
		if err = locker.Release(ctx, cacheKey, lockerVal); err != nil {
			log.Warnf("failed to release the lock of cacheKey %s, err= %v", cacheKey, err)
		}
	}()

	// Double check to avoid cache stampede
	users, err = cache.SET().GetAllUInt64s(ctx, cacheKey)
	if err == nil && len(users) > 0 && users[0] != utils.UInt64(0) {
		return users, nil
	}

	users, err = s.repo.GetUsersByTag(ctx, tag)
	if err != nil {
		log.Errorf("failed to fetch users from repo for tag %s, err= %v", tag, err)
		return nil, err
	}

	if err = cache.SET().InsertUInt64s(ctx, cacheKey, TagCacheExpiration, users...); err != nil {
		log.Warnf("failed to update the string list cache, err= %v", err)
	}
	return users, nil
}

func (s *TagService[EntityType]) GetEntities(ctx context.Context, userID utils.UInt64, tag string, entityType *EntityType) ([]utils.UInt64, error) {
	if entityType != nil {
		return s.getEntitiesByTagAndType(ctx, userID, tag, *entityType)
	}

	var allEntities []utils.UInt64
	for _, et := range s.supportedTypes {
		entities, err := s.getEntitiesByTagAndType(ctx, userID, tag, et)
		if err != nil {
			return nil, err
		}
		allEntities = append(allEntities, entities...)
	}
	return allEntities, nil
}

func (s *TagService[EntityType]) getEntitiesByTagAndType(ctx context.Context, userID utils.UInt64, tag string, entityType EntityType) ([]utils.UInt64, error) {
	ll, ctx := wlog.ByCtxAndCache(ctx, "getEntitiesByTagAndType")
	log := ll.WithField("userID", userID).WithField("tag", tag).WithField("entityType", entityType)
	cacheKey := s.Schemas.Entities.MustBuild(ParamUserTagType[EntityType]{UserID: userID, Tag: tag, Type: entityType})
	entities, err := cache.SET().GetAllUInt64s(ctx, cacheKey)
	if err == nil && len(entities) > 0 && entities[0] != utils.UInt64(0) {
		return entities, nil
	}

	// 获取锁
	locker := cache.Locker(ctx)
	lockerVal, err := locker.Acquire(ctx, cacheKey, time.Second*10)
	if err != nil {
		log.Errorf("failed to acquire lock for cacheKey %s, err= %v", cacheKey, err)
		return nil, err
	}
	defer func() {
		if err = locker.Release(ctx, cacheKey, lockerVal); err != nil {
			log.Warnf("failed to release the lock of cacheKey %s, err= %v", cacheKey, err)
		}
	}()

	// Double check to avoid cache stampede
	entities, err = cache.SET().GetAllUInt64s(ctx, cacheKey)
	if err == nil && len(entities) > 0 && entities[0] != utils.UInt64(0) {
		return entities, nil
	}

	entities, err = s.repo.GetEntitiesByTag(ctx, userID, tag, entityType)
	if err != nil {
		log.Errorf("failed to fetch entities from repo for userID %d, tag %s, entityType %v, err= %v", userID, tag, entityType, err)
		return nil, err
	}

	if err = cache.SET().InsertUInt64s(ctx, cacheKey, TagCacheExpiration, entities...); err != nil {
		log.Warnf("failed to update the string list cache, err= %v", err)
	}
	return entities, nil
}

func (s *TagService[EntityType]) GetTagsOfEntity(ctx context.Context, entityID utils.UInt64) ([]string, error) {
	ll, ctx := wlog.ByCtxAndCache(ctx, "GetTagsOfEntity")
	log := ll.WithField("entityID", entityID)
	cacheKey := s.Schemas.Entity2Tags.MustBuild(entityID)
	tags, err := cache.SET().GetAll(ctx, cacheKey)
	if err == nil && len(tags) > 0 && tags[0] != CacheInvalidationFlag {
		return tags, nil
	}

	// 获取锁
	locker := cache.Locker(ctx)
	lockerVal, err := locker.Acquire(ctx, cacheKey, time.Second*10)
	if err != nil {
		log.Errorf("failed to acquire lock for cacheKey %s, err= %v", cacheKey, err)
		return nil, err
	}
	defer func() {
		if err = locker.Release(ctx, cacheKey, lockerVal); err != nil {
			log.Warnf("failed to release the lock of cacheKey %s, err= %v", cacheKey, err)
		}
	}()

	// Double check to avoid cache stampede
	tags, err = cache.SET().GetAll(ctx, cacheKey)
	if err == nil && len(tags) > 0 && tags[0] != CacheInvalidationFlag {
		return tags, nil
	}

	tags, err = s.repo.GetTagsByEntity(ctx, entityID)
	if err != nil {
		log.Errorf("failed to fetch tags from repo for entityID %d, err= %v", entityID, err)
		return nil, err
	}

	if err = cache.SET().Insert(ctx, cacheKey, TagCacheExpiration, tags...); err != nil {
		log.Warnf("failed to update the string list cache, err= %v", err)
	}

	return tags, nil
}

// --- 标脏接口实现 ---

// InvalidateUserCache 清除和用户有关的 cache
func (s *TagService[EntityType]) InvalidateUserCache(ctx context.Context, userID utils.UInt64, propagate bool) error {
	log := wlog.ByCtx(ctx, "InvalidateUserCache").WithField("userID", userID)

	// 1. 查找用户关联的所有标签
	cacheKey := s.Schemas.User2Tags.MustBuild(userID)
	tags, err := cache.SET().GetAll(ctx, cacheKey)
	if err != nil {
		log.Errorf("Failed to get tags from cache: %v", err)
		return err
	}
	if len(tags) == 0 || tags[0] == CacheInvalidationFlag {
		log.Infof("No tags found for user %v", userID)
		return nil
	}

	if propagate {
		// 2. 发送标签标脏事件
		for _, tag := range tags {
			if err = s.UpdateMgr.Put(ctx, TagUpdateMessage[EntityType]{UserID: userID, Tag: tag, Action: EventInvalidTag, Propagate: false}); err != nil {
				log.Errorf("Failed to enqueue tag invalidation message: %v", err)
				return err
			}
		}

		// 3. 发送实体标脏事件
		for _, tag := range tags {
			for _, entityType := range s.supportedTypes {
				if err = s.UpdateMgr.Put(ctx, TagUpdateMessage[EntityType]{UserID: userID, Tag: tag, EntityType: entityType, Action: EventInvalidEntity, Propagate: false}); err != nil {
					log.Errorf("Failed to enqueue entity invalidation message: %v", err)
					return err
				}
			}
		}
	}

	// 4. 清除用户自己的缓存
	if err = cache.SET().Clear(ctx, cacheKey, MaxRetryAttempts); err != nil {
		log.Errorf("Failed to clear user cache: %v", err)
		return err
	}

	return nil
}

// InvalidateTagCache 清除和 Tag 有关的 cache
func (s *TagService[EntityType]) InvalidateTagCache(ctx context.Context, tag string, propagate bool) error {
	log := wlog.ByCtx(ctx, "InvalidateTagCache").WithField("tag", tag)

	// 1. 查找标签关联的所有用户
	cacheKey := s.Schemas.Tag2Users.MustBuild(tag)
	userIDs, err := cache.SET().GetAllUInt64s(ctx, cacheKey)
	if err != nil {
		log.Errorf("Failed to get users from cache: %v", err)
		return err
	}
	if len(userIDs) == 0 || userIDs[0] == utils.UInt64(0) {
		log.Infof("No users found for tag %v", tag)
		return nil
	}

	if propagate {
		// 2. 发送用户标脏事件
		for _, userID := range userIDs {
			if err = s.UpdateMgr.Put(ctx, TagUpdateMessage[EntityType]{UserID: userID, Tag: tag, Action: EventInvalidUser, Propagate: false}); err != nil {
				log.Errorf("Failed to enqueue user invalidation message: %v", err)
				return err
			}
		}

		// 3. 发送实体标脏事件
		for _, userID := range userIDs {
			for _, entityType := range s.supportedTypes {
				if err = s.UpdateMgr.Put(ctx, TagUpdateMessage[EntityType]{UserID: userID, Tag: tag, EntityType: entityType, Action: EventInvalidEntity, Propagate: false}); err != nil {
					log.Errorf("Failed to enqueue entity invalidation message: %v", err)
					return err
				}
			}
		}
	}

	// 4. 清除标签自己的缓存
	if err = cache.SET().Clear(ctx, cacheKey, MaxRetryAttempts); err != nil {
		log.Errorf("Failed to clear tag cache: %v", err)
		return err
	}

	return nil
}

// InvalidateUserTagCache 清除和 User、Tag 有关的 cache, 只针对精确的 user tag 对
func (s *TagService[EntityType]) InvalidateUserTagCache(ctx context.Context, userID utils.UInt64, tag string, propagate bool) error {
	log := wlog.ByCtx(ctx, "InvalidateUserTagCache").WithField("userID", userID).WithField("tag", tag)

	if propagate {
		// 1. 发送实体标脏事件
		for _, entityType := range s.supportedTypes {
			if err := s.UpdateMgr.Put(ctx, TagUpdateMessage[EntityType]{UserID: userID, Tag: tag, EntityType: entityType, Action: EventInvalidEntity, Propagate: false}); err != nil {
				log.Errorf("Failed to enqueue entity invalidation message: %v", err)
				return err
			}
		}
	}

	// 2. 清除用户标签自己的缓存
	for _, entityType := range s.supportedTypes {
		cacheKey := s.Schemas.Entities.MustBuild(ParamUserTagType[EntityType]{UserID: userID, Tag: tag, Type: entityType})
		if err := cache.SET().Clear(ctx, cacheKey, MaxRetryAttempts); err != nil {
			log.Errorf("Failed to clear user tag cache: %v", err)
			return err
		}
	}

	return nil
}

// InvalidateEntityCache 清除 entity 相关的 cache
func (s *TagService[EntityType]) InvalidateEntityCache(ctx context.Context, entityID utils.UInt64, propagate bool) error {
	log := wlog.ByCtx(ctx, "InvalidateEntityCache").WithField("entityID", entityID)

	// 1. 查找实体关联的所有标签
	cacheKey := s.Schemas.Entity2Tags.MustBuild(entityID)
	tags, err := cache.SET().GetAll(ctx, cacheKey)
	if err != nil {
		log.Errorf("Failed to get tags from cache: %v", err)
		return err
	}
	if len(tags) == 0 || tags[0] == CacheInvalidationFlag {
		log.Infof("No tags found for entity %v", entityID)
		return nil
	}

	if propagate {
		// 2. 发送标签标脏事件
		for _, tag := range tags {
			if err = s.UpdateMgr.Put(ctx, TagUpdateMessage[EntityType]{EntityID: entityID, Tag: tag, Action: EventInvalidTag, Propagate: false}); err != nil {
				log.Errorf("Failed to enqueue tag invalidation message: %v", err)
				return err
			}
		}
	}

	// 3. 清除实体自己的缓存
	if err = cache.SET().Clear(ctx, cacheKey, MaxRetryAttempts); err != nil {
		log.Errorf("Failed to clear entity cache: %v", err)
		return err
	}

	return nil
}

// handleTagUpdateMessage handles a tag update message.
func (s *TagService[EntityType]) handleTagUpdateMessage(ctx context.Context, message TagUpdateMessage[EntityType]) error {
	switch message.Action {
	case EventInvalidUser:
		return s.InvalidateUserCache(ctx, message.UserID, false)
	case EventInvalidTag:
		return s.InvalidateTagCache(ctx, message.Tag, false)
	case EventInvalidEntity:
		return s.InvalidateEntityCache(ctx, message.EntityID, false)
	default:
		return irr.Trace("unknown action: %v", message.Action)
	}
}

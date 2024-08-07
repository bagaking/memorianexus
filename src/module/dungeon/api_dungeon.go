package dungeon

import (
	"errors"
	"net/http"
	"time"

	"github.com/bagaking/memorianexus/internal/utils"

	"github.com/bagaking/memorianexus/src/def"
	"github.com/khicago/irr"
	"gorm.io/gorm"

	"github.com/bagaking/memorianexus/src/module/dto"

	"github.com/gin-gonic/gin"

	"github.com/bagaking/goulp/wlog"
	"github.com/bagaking/memorianexus/src/model"
)

type ReqCreateDungeon struct {
	dto.DungeonData

	Books []utils.UInt64 `json:"books,omitempty"`
	Items []utils.UInt64 `json:"items,omitempty"`
	Tags  []string       `json:"tags,omitempty"`
}

type ReqUpdateDungeon struct {
	dto.DungeonData
}

// CreateDungeon handles the creation of a new dungeon campaign
// @Summary Create a new dungeon campaign
// @Description 创建新的复习计划
// @Tags dungeon
// @Accept json
// @Produce json
// @Param campaign body ReqCreateDungeon true "Dungeon campaign data"
// @Success 201 {object} dto.RespDungeon "Successfully created dungeon"
// @Failure 400 {object} utils.ErrorResponse "Invalid request body"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Router /dungeon/dungeons [post]
func (svr *Service) CreateDungeon(c *gin.Context) {
	userID := utils.GinMustGetUserID(c)
	log := wlog.ByCtx(c, "CreateDungeon").WithField("user_id", userID)

	profile, err := model.FindProfile(c, svr.db, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.GinHandleError(c, log, http.StatusNotFound, err, "profile not found")
			return
		}
		utils.GinHandleError(c, log, http.StatusInternalServerError, err, "Internal server error, find profile failed")
		return
	}

	var req ReqCreateDungeon
	if err = c.ShouldBindJSON(&req); err != nil {
		utils.GinHandleError(c, log, http.StatusBadRequest, irr.Wrap(err, "parse request body failed"), "Invalid request body")
		return
	}

	if req.Type != def.DungeonTypeCampaign && req.Type != def.DungeonTypeEndless {
		utils.GinHandleError(c, log, http.StatusBadRequest,
			irr.Error("invalid dungeon type %v", req.Type), "Invalid request body", utils.GinErrWithReqBody(req))
		return
	}

	dungeonID, err := utils.GenIDU64(c)
	if err != nil {
		utils.GinHandleError(c, log, http.StatusInternalServerError, err, "Failed to generate ID", utils.GinErrWithReqBody(req))
		return
	}

	memorizationSetting := model.DefaultMemorizationSetting // copy
	if req.SettingsMemorization != nil {
		req.SettingsMemorization.ToModel(&memorizationSetting)
	} else {
		s, err := profile.GetSettingsMemorizationOrDefault(c, svr.db)
		if err != nil {
			utils.GinHandleError(c, log, http.StatusInternalServerError, err, "Internal server error, get user settings failed")
			return
		}
		memorizationSetting = s.MemorizationSetting
	}

	dungeon, err := model.CreateDungeon(c, svr.db, &model.Dungeon{
		ID:                  dungeonID,
		UserID:              userID,
		Type:                req.Type,
		Title:               req.Title,
		Description:         req.Description,
		MemorizationSetting: memorizationSetting, // fork setting form profile
	})
	// Create dungeon entry in the database
	if err != nil {
		utils.GinHandleError(c, log, http.StatusInternalServerError, err, "Internal server error, create dungeon failed", utils.GinErrWithReqBody(req))
		return
	}

	// Add books to dungeon
	if err = dungeon.AddMonsterFromBook(c, svr.db, req.Books); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.GinHandleError(c, log, http.StatusNotFound, err, "Internal server error, books not found", utils.GinErrWithReqBody(req))
		} else {
			utils.GinHandleError(c, log, http.StatusInternalServerError, err, "Internal server error", utils.GinErrWithReqBody(req))
		}
		return
	}

	// Add items to dungeon
	if err = dungeon.AddMonsters(c, svr.db, req.Items); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.GinHandleError(c, log, http.StatusNotFound, err, "Internal server error, items not found", utils.GinErrWithReqBody(req))
		} else {
			utils.GinHandleError(c, log, http.StatusInternalServerError, err, "Internal server error", utils.GinErrWithReqBody(req))
		}
		return
	}

	resp := new(dto.RespDungeon).With(new(dto.Dungeon).FromModel(dungeon))
	resp.Data.Books = req.Books
	resp.Data.Items = req.Items
	resp.Data.Tags = req.Tags
	resp.Response(c, "dungeon created")
}

// GetDungeons handles fetching the list of dungeon campaigns
// @Summary Get the list of dungeon campaigns
// @Description 获取复习计划列表
// @Tags dungeon
// @Produce json
// @Success 200 {array} dto.RespDungeonList "Successfully retrieved dungeons"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Router /dungeon/dungeons [get]
func (svr *Service) GetDungeons(c *gin.Context) {
	l, ctx := wlog.ByCtxAndCache(c, "GetDungeons")
	userID := utils.GinMustGetUserID(c)

	log := l.WithField("user_id", userID)

	var req ReqGetDungeon
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.GinHandleError(c, log, http.StatusBadRequest, err, "Invalid query parameters")
		return
	}

	pager := utils.GinGetPagerFromQuery(c)

	var dungeons []model.Dungeon

	tx := svr.db.Where("user_id = ?", userID).Offset(pager.Offset).Limit(pager.Limit).Find(&dungeons)
	if req.Type.Valid() {
		tx.Where("type = ?", req.Type)
	}
	if err := tx.Error; err != nil {
		utils.GinHandleError(c, log, http.StatusInternalServerError, err, "Failed to fetch dungeon campaigns")
		return
	}

	resp := new(dto.RespDungeonList).WithPager(pager)
	for i := range dungeons {
		dungeon := dungeons[i]
		books, items, tags, err := dungeon.GetAssociations(ctx, svr.db)
		if err != nil {
			utils.GinHandleError(c, log, http.StatusInternalServerError, err, "Failed to fetch dungeon associations")
			return
		}
		d := new(dto.Dungeon).FromModel(&dungeon)
		d.Books = books
		d.Items = items
		d.Tags = tags
		resp.Append(d)
	}
	resp.Response(c)
}

// GetDungeon handles fetching the details of a specific dungeon campaign
// @Summary Get the details of a specific dungeon campaign
// @Description 获取复习计划详情
// @Tags dungeon
// @Produce json
// @Param id path uint64 true "Dungeon ID"
// @Success 200 {object} dto.RespDungeon "Successfully retrieved dungeon"
// @Failure 404 {object} utils.ErrorResponse "Dungeon not found"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Router /dungeon/dungeons/{id} [get]
func (svr *Service) GetDungeon(c *gin.Context) {
	l, ctx := wlog.ByCtxAndCache(c, "GetDungeon")
	userID := utils.GinMustGetUserID(c)

	id := utils.GinMustGetID(c)
	log := l.WithField("user_id", userID).WithField("dungeon_id", id)

	var dungeon model.Dungeon
	if err := svr.db.Where("user_id = ? and id = ?", userID, id).First(&dungeon).Error; err != nil {
		utils.GinHandleError(c, log, http.StatusNotFound, err, "Dungeon not found")
		return
	}

	books, items, tags, err := dungeon.GetAssociations(ctx, svr.db)
	if err != nil {
		utils.GinHandleError(c, log, http.StatusInternalServerError, err, "Failed to fetch dungeon associations")
		return
	}

	resp := new(dto.RespDungeon).With(new(dto.Dungeon).FromModel(&dungeon))
	resp.Data.Books = books
	resp.Data.Items = items
	resp.Data.Tags = tags
	resp.Response(c)
}

// UpdateDungeon handles updating a specific dungeon campaign
// @Summary Update a specific dungeon campaign
// @Description 更新复习计划
// @Tags dungeon
// @Accept json
// @Produce json
// @Param id path uint64 true "Dungeon ID"
// @Param campaign body ReqUpdateDungeon true "Dungeon campaign data"
// @Success 200 {object} dto.RespDungeon "Successfully updated dungeon"
// @Failure 400 {object} utils.ErrorResponse "Invalid request body"
// @Failure 404 {object} utils.ErrorResponse "Dungeon not found"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Router /dungeon/dungeons/{id} [put]
func (svr *Service) UpdateDungeon(c *gin.Context) {
	userID := utils.GinMustGetUserID(c)
	id := utils.GinMustGetID(c)
	log := wlog.ByCtx(c, "UpdateDungeon").WithField("user_id", userID).WithField("dungeon_id", id)

	var req ReqUpdateDungeon

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.GinHandleError(c, log, http.StatusBadRequest, err, "Invalid request body")
		return
	}

	updater := &model.Dungeon{
		Type:        req.Type,
		Title:       req.Title,
		Description: req.Description,
		UpdatedAt:   time.Now(),
	}

	if req.SettingsMemorization != nil {
		req.SettingsMemorization.ToModel(&updater.MemorizationSetting)
	}

	if err := svr.db.Where("user_id = ? AND id = ?", userID, id).Updates(updater).Error; err != nil {
		utils.GinHandleError(c, log, http.StatusNotFound, err, "Failed to update dungeon")
		return
	}

	resp := new(dto.RespDungeon).With(new(dto.Dungeon).FromModel(updater))
	resp.Response(c, "dungeon updated")
}

// DeleteDungeon handles deleting a specific dungeon campaign
// @Summary Delete a specific dungeon campaign
// @Description 删除复习计划
// @Tags dungeon
// @Param id path uint64 true "Dungeon ID"
// @Success 204 {object} dto.RespDungeon "Successfully deleted dungeon"
// @Failure 404 {object} utils.ErrorResponse "Dungeon not found"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Router /dungeon/dungeons/{id} [delete]
func (svr *Service) DeleteDungeon(c *gin.Context) {
	userID := utils.GinMustGetUserID(c)
	id := utils.GinMustGetID(c)
	log := wlog.ByCtx(c, "DeleteDungeon").WithField("user_id", userID).WithField("dungeon_id", id)

	var dungeon model.Dungeon

	tx := svr.db.Begin()
	if err := tx.Where("user_id = ? AND id = ?", userID, id).First(&dungeon).Error; err != nil {
		utils.GinHandleError(c, log, http.StatusNotFound, err, "Dungeon not found")
		return
	}

	// Delete dungeon entry in the database
	if err := tx.Delete(&dungeon).Error; err != nil {
		utils.GinHandleError(c, log, http.StatusInternalServerError, err, "Failed to delete dungeon")
		return
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		utils.GinHandleError(c, log, http.StatusInternalServerError, err, "commit failed")
		return
	}

	resp := new(dto.RespDungeon).With(new(dto.Dungeon).FromModel(&dungeon))
	resp.Response(c, "dungeon deleted")
}

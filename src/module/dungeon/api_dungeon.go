package dungeon

import (
	"errors"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/bagaking/memorianexus/src/def"
	"github.com/khicago/irr"
	"gorm.io/gorm"

	"github.com/bagaking/memorianexus/src/module/dto"

	"github.com/gin-gonic/gin"

	"github.com/bagaking/goulp/wlog"
	"github.com/bagaking/memorianexus/internal/utils"
	"github.com/bagaking/memorianexus/src/model"
)

type ReqCreateDungeon struct {
	dto.DungeonFullData
}

type ReqUpdateDungeon struct {
	dto.DungeonData
}

// CreateDungeon handles the creation of a new dungeon campaign
// @Summary Create a new dungeon campaign
// @Description 创建新的复习计划
// @TagNames dungeon
// @Accept json
// @Produce json
// @Param campaign body ReqCreateDungeon true "Dungeon campaign data"
// @Success 201 {object} dto.RespDungeon
// @Failure 400 {object} utils.ErrorResponse "Invalid request body"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Router /dungeon/dungeons [post]
func (svr *Service) CreateDungeon(c *gin.Context) {
	log := wlog.ByCtx(c, "CreateDungeon")
	// 从请求上下文中获取当前用户ID
	userID, exists := utils.GetUIDFromGinCtx(c)
	if !exists {
		utils.GinHandleError(c, log, http.StatusUnauthorized, errors.New("user not authenticated"), "User not authenticated")
		return
	}

	var req ReqCreateDungeon
	if err := c.ShouldBindJSON(&req); err != nil {
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

	dungeon := model.Dungeon{
		ID:          dungeonID,
		UserID:      userID,
		Type:        req.Type,
		Title:       req.Title,
		Description: req.Description,
		Rule:        req.Rule,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Create dungeon entry in the database
	if err = svr.db.Create(&dungeon).Error; err != nil {
		utils.GinHandleError(c, log, http.StatusInternalServerError, err, "Internal server error, create dungeon failed", utils.GinErrWithReqBody(req))
		return
	}

	// Add books to dungeon
	if err = dungeon.AddMonster(svr.db, model.MonsterSourceBook, req.Books); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.GinHandleError(c, log, http.StatusNotFound, err, "Internal server error, books not found", utils.GinErrWithReqBody(req))
		} else {
			utils.GinHandleError(c, log, http.StatusInternalServerError, err, "Internal server error", utils.GinErrWithReqBody(req))
		}
		return
	}

	tagIDs, err := model.FindTagsIDByName(svr.db, req.TagNames) // todo: 未创建的 tag 会被忽略
	if err != nil {
		utils.GinHandleError(c, log, http.StatusNotFound, err,
			"Internal server error, get tagID failed", utils.GinErrWithReqBody(req))
		return
	}

	// Add tags to dungeon
	if err = dungeon.AddMonster(svr.db, model.MonsterSourceTag, tagIDs); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.GinHandleError(c, log, http.StatusNotFound, err,
				"Internal server error, tags not found", utils.GinErrWithReqBody(req))
		} else {
			utils.GinHandleError(c, log, http.StatusInternalServerError, err,
				"Internal server error", utils.GinErrWithReqBody(req))
		}
		return
	}
	// Add items to dungeon
	if err = dungeon.AddMonster(svr.db, model.MonsterSourceItem, req.Items); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.GinHandleError(c, log, http.StatusNotFound, err,
				"Internal server error, items not found", utils.GinErrWithReqBody(req))
		} else {
			utils.GinHandleError(c, log, http.StatusInternalServerError, err,
				"Internal server error", utils.GinErrWithReqBody(req))
		}
		return
	}

	if err = dungeon.AddMonster(svr.db, model.MonsterSourceTag, tagIDs); err != nil {
		utils.GinHandleError(c, log, http.StatusInternalServerError, err, "Internal server error", utils.GinErrWithReqBody(req))
		return
	}

	resp := new(dto.RespDungeon).With(new(dto.Dungeon).FromModel(&dungeon))
	resp.Data.Books = req.Books
	resp.Data.Items = req.Items
	resp.Data.TagNames = req.TagNames
	resp.Data.TagIDs = tagIDs
	resp.Response(c, "dungeon created")
}

// GetDungeons handles fetching the list of dungeon campaigns
// @Summary Get the list of dungeon campaigns
// @Description 获取复习计划列表
// @TagNames dungeon
// @Produce json
// @Success 200 {array} dto.RespDungeonList
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Router /dungeon/dungeons [get]
func (svr *Service) GetDungeons(c *gin.Context) {
	log := wlog.ByCtx(c, "GetDungeons")
	// 从请求上下文中获取当前用户ID
	userID, exists := utils.GetUIDFromGinCtx(c)
	if !exists {
		utils.GinHandleError(c, log, http.StatusUnauthorized, errors.New("user not authenticated"), "User not authenticated")
		return
	}

	var dungeons []model.Dungeon
	if err := svr.db.Where("user_id = ?", userID).Find(&dungeons).Error; err != nil {
		utils.GinHandleError(c, log, http.StatusInternalServerError, err, "Failed to fetch dungeon campaigns")
		return
	}

	resp := new(dto.RespDungeonList)
	for i := range dungeons {
		dungeon := dungeons[i]
		books, items, tags, err := model.GetDungeonAssociations(svr.db, dungeon.ID)
		if err != nil {
			utils.GinHandleError(c, log, http.StatusInternalServerError, err, "Failed to fetch dungeon associations")
			return
		}
		d := new(dto.Dungeon).FromModel(&dungeon)
		d.Books = books
		d.Items = items
		d.TagIDs = tags
		resp.Append(d)
	}
	resp.Response(c)
}

// GetDungeon handles fetching the details of a specific dungeon campaign
// @Summary Get the details of a specific dungeon campaign
// @Description 获取复习计划详情
// @TagNames dungeon
// @Produce json
// @Param id path uint64 true "Dungeon ID"
// @Success 200 {object} dto.RespDungeon "the dungeon with its associations"
// @Failure 404 {object} utils.ErrorResponse "Dungeon not found"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Router /dungeon/dungeons/{id} [get]
func (svr *Service) GetDungeon(c *gin.Context) {
	log := wlog.ByCtx(c, "GetDungeon")
	userID, exists := utils.GetUIDFromGinCtx(c)
	if !exists {
		utils.GinHandleError(c, log, http.StatusUnauthorized, errors.New("user not authenticated"), "User not authenticated")
		return
	}

	var dungeon model.Dungeon
	id := c.Param("id")

	if err := svr.db.Where("id = ? and user_id = ?", id, userID).First(&dungeon).Error; err != nil {
		utils.GinHandleError(c, log, http.StatusNotFound, err, "Dungeon not found")
		return
	}

	books, items, tags, err := model.GetDungeonAssociations(svr.db, dungeon.ID)
	if err != nil {
		utils.GinHandleError(c, log, http.StatusInternalServerError, err, "Failed to fetch dungeon associations")
		return
	}

	resp := new(dto.RespDungeon).With(new(dto.Dungeon).FromModel(&dungeon))
	resp.Data.Books = books
	resp.Data.Items = items
	resp.Data.TagIDs = tags
	resp.Response(c)
}

// UpdateDungeon handles updating a specific dungeon campaign
// @Summary Update a specific dungeon campaign
// @Description 更新复习计划
// @TagNames dungeon
// @Accept json
// @Produce json
// @Param id path uint64 true "Dungeon ID"
// @Param campaign body ReqUpdateDungeon true "Dungeon campaign data"
// @Success 200 {object} dto.RespDungeon "updater"
// @Failure 400 {object} utils.ErrorResponse "Invalid request body"
// @Failure 404 {object} utils.ErrorResponse "Dungeon not found"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Router /dungeon/dungeons/{id} [put]
func (svr *Service) UpdateDungeon(c *gin.Context) {
	var log logrus.FieldLogger = wlog.ByCtx(c, "UpdateDungeon")
	userID, exists := utils.GetUIDFromGinCtx(c)
	if !exists {
		utils.GinHandleError(c, log, http.StatusUnauthorized, errors.New("user not authenticated"), "User not authenticated")
		return
	}
	log = log.WithField("user_id", userID)

	var req ReqUpdateDungeon

	id := c.Param("id")
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.GinHandleError(c, wlog.ByCtx(c), http.StatusBadRequest, err, "Invalid request body")
		return
	}

	dungeon := &model.Dungeon{
		Type:        req.Type,
		Title:       req.Title,
		Description: req.Description,
		Rule:        req.Rule,
		UpdatedAt:   time.Now(),
	}
	if err := svr.db.Where("user_id = ? AND id = ?", userID, id).Updates(dungeon).Error; err != nil {
		utils.GinHandleError(c, wlog.ByCtx(c), http.StatusNotFound, err, "Failed to update dungeon")
		return
	}

	// Fetch the updated dungeon
	var updatedDungeon model.Dungeon
	if err := svr.db.Where("user_id = ? AND id = ?", userID, id).First(&updatedDungeon).Error; err != nil {
		utils.GinHandleError(c, wlog.ByCtx(c), http.StatusNotFound, err, "Failed to fetch updated dungeon")
		return
	}

	resp := new(dto.RespDungeon).With(new(dto.Dungeon).FromModel(&updatedDungeon))
	resp.Response(c, "dungeon updated")
}

// DeleteDungeon handles deleting a specific dungeon campaign
// @Summary Delete a specific dungeon campaign
// @Description 删除复习计划
// @TagNames dungeon
// @Param id path uint64 true "Dungeon ID"
// @Success 204 {object} dto.RespDungeon "updater"
// @Failure 404 {object} utils.ErrorResponse "Dungeon not found"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Router /dungeon/dungeons/{id} [delete]
func (svr *Service) DeleteDungeon(c *gin.Context) {
	log := wlog.ByCtx(c, "DeleteDungeon")
	userID, exists := utils.GetUIDFromGinCtx(c)
	if !exists {
		utils.GinHandleError(c, log, http.StatusUnauthorized, errors.New("user not authenticated"), "User not authenticated")
		return
	}

	var dungeon model.Dungeon
	id := c.Param("id")

	tx := svr.db.Begin()
	if err := tx.Where("user_id = ? AND id = ?", userID, id).First(&dungeon).Error; err != nil {
		utils.GinHandleError(c, wlog.ByCtx(c), http.StatusNotFound, err, "Dungeon not found")
		return
	}

	// Delete dungeon entry in the database
	if err := tx.Delete(&dungeon).Error; err != nil {
		utils.GinHandleError(c, wlog.ByCtx(c), http.StatusInternalServerError, err, "Failed to delete dungeon")
		return
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		utils.GinHandleError(c, wlog.ByCtx(c), http.StatusInternalServerError, err, "commit failed")
		return
	}

	resp := new(dto.RespDungeon).With(new(dto.Dungeon).FromModel(&dungeon))
	resp.Response(c, "dungeon deleted")
}

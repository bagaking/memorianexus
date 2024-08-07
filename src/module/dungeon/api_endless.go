package dungeon

import (
	"net/http"
	"strconv"

	"github.com/bagaking/goulp/wlog"

	"github.com/gin-gonic/gin"

	"github.com/bagaking/memorianexus/internal/utils"
	"github.com/bagaking/memorianexus/src/model"
	"github.com/bagaking/memorianexus/src/module/dto"
)

// GetMonstersOfEndlessDungeon handles fetching all the monsters of a specific endless dungeon with associations
// @Summary Get all the monsters of a specific endless dungeon with associations
// @Description 获取复习计划的所有Monsters及其关联的 Items, Books, Tags
// @Tags dungeon
// @Produce json
// @Param id path uint64 true "Dungeon ID"
// @Param offset query int true "Offset for pagination"
// @Param limit query int true "Limit for pagination"
// @Success 200 {object} dto.RespMonsterList
// @Failure 404 {object} utils.ErrorResponse "Dungeon not found"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Router /dungeon/endless/{id}/monsters [get]
func (svr *Service) GetMonstersOfEndlessDungeon(c *gin.Context) {
	id := c.Param("id")

	// sortBy := c.DefaultQuery("sort_by", "familiarity") // (familiarity, difficulty, importance)
	offsetStr := c.DefaultQuery("offset", "0")
	limitStr := c.DefaultQuery("limit", "10")
	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		offset = 0
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 10
	}

	var dungeon model.Dungeon
	if err = svr.db.Where("id = ?", id).First(&dungeon).Error; err != nil {
		utils.GinHandleError(c, wlog.ByCtx(c), http.StatusNotFound, err, "Dungeon not found")
		return
	}

	monsters, err := dungeon.GetMonstersWithExpandedAssociations(c, svr.db, offset, limit)
	if err != nil {
		utils.GinHandleError(c, wlog.ByCtx(c), http.StatusInternalServerError, err, "Failed to fetch dungeon monsters with associations")
		return
	}

	resp := new(dto.RespMonsterList)
	for _, monster := range monsters {
		resp.Append(new(dto.DungeonMonster).FromModel(monster))
	}
	resp.Response(c, "monsters found")
}

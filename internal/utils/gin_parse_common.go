package utils

import (
	"net/http"
	"strconv"

	"github.com/khgame/ranger_iam/pkg/authcli"

	"github.com/bagaking/goulp/wlog"
	"github.com/gin-gonic/gin"
	"github.com/khicago/irr"
)

func GinMWParseID() gin.HandlerFunc {
	return func(c *gin.Context) {
		log := wlog.ByCtx(c, "gin_parse_id")
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			GinHandleError(c, log, http.StatusBadRequest, irr.Wrap(err, "id= %s", idStr), "Invalid ID")
			c.Abort()
			return
		}
		c.Set("__parsed_id", UInt64(id))
		c.Next()
	}
}

func GinMWParseTAG() gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("tag")
		c.Set("__parsed_tag", idStr)
		c.Next()
	}
}

// GinMustGetID should be used with GinMWParseID
func GinMustGetID(c *gin.Context) (id UInt64) {
	return c.MustGet("__parsed_id").(UInt64)
}

// GinMustGetTAG should be used with GinMWParseID
func GinMustGetTAG(c *gin.Context) (tag string) {
	return c.MustGet("__parsed_tag").(string)
}

func GinMustGetUserID(c *gin.Context) (id UInt64) {
	idu, exist := authcli.GetUIDFromGinCtx(c)
	if !exist {
		panic("UserID does not exist")
	}
	return UInt64(idu)
}

func GinGetPagerFromQuery(c *gin.Context) (pager *Pager) {
	pageStr := c.DefaultQuery("page", "1")
	offsetStr := c.DefaultQuery("offset", "0")
	limitStr := c.DefaultQuery("limit", "10")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 10
	}
	if limit <= 0 {
		limit = 10
	}

	page, err := strconv.Atoi(pageStr)
	if err != nil {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil {
			offset = 0
		}
		if offset < 0 {
			offset = 0
		}
		return new(Pager).SetOffsetAndLimit(offset, limit)
	}
	if page <= 0 {
		page = 1
	}

	return new(Pager).SetPageAndLimit(page, limit)
}

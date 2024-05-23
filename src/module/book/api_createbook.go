package book

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/bagaking/goulp/wlog"
	"github.com/bagaking/memorianexus/internal/utils"
	"github.com/bagaking/memorianexus/src/model"
	"github.com/bagaking/memorianexus/src/module"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CreateBook handles the request to create a new book.
// @Summary Create a book
// @Description Creates a new book and optionally associates tags with it
// @Tags book
// @Accept json
// @Produce json
// @Param book body ReqCreateBook true "Book creation data"
// @Success 201 {object} RespBook "Successfully created book"
// @Failure 400 {object} module.ErrorResponse "Invalid parameters"
// @Router /books [post]
func (svr *Service) CreateBook(c *gin.Context) {
	log := wlog.ByCtx(c, "CreateBook")
	userID, exists := utils.GetUIDFromGinCtx(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req ReqCreateBook
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.GinHandleError(c, log, http.StatusBadRequest, err, "Invalid request body")
		return
	}

	bookID, err := utils.GenIDU64(c)
	if err != nil {
		utils.GinHandleError(c, log, http.StatusInternalServerError, err, "Failed to generate ID")
		return
	}

	book := &model.Book{
		ID:          bookID,
		UserID:      userID,
		Title:       req.Title,
		Description: req.Description,
	}

	// Begin a transaction
	tx := svr.db.Begin()
	defer func() {
		// Ensure transaction rollback on error or panic
		if r := recover(); r != nil || tx.Error != nil {
			tx.Rollback()
			if r != nil {
				utils.GinHandleError(c, log, http.StatusInternalServerError, fmt.Errorf("%v", r), "Transaction failed")
			} else {
				utils.GinHandleError(c, log, http.StatusInternalServerError, tx.Error, "Transaction failed")
			}
		}
	}()

	// Create the book record in the database
	if err = tx.Create(book).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			c.JSON(http.StatusConflict, module.ErrorResponse{Message: "Book already exists"})
		} else {
			utils.GinHandleError(c, log, http.StatusInternalServerError, err, "Failed to create book")
		}
		return
	}

	// Update tags associated with the book
	if err = model.UpdateBookTagsRef(c, tx, book.ID, req.Tags); err != nil {
		utils.GinHandleError(c, log, http.StatusInternalServerError, err, "Failed to update book tags")
		return
	}

	// Commit the transaction
	if err = tx.Commit().Error; err != nil {
		utils.GinHandleError(c, log, http.StatusInternalServerError, err, "Failed to commit transaction")
		return
	}

	// Construct the response
	resp := RespBook{
		ID:          book.ID,
		UserID:      book.UserID,
		Title:       book.Title,
		Description: book.Description,
		CreatedAt:   book.CreatedAt,
		UpdatedAt:   book.UpdatedAt,
	}

	// Send the response
	c.JSON(http.StatusCreated, resp)
}
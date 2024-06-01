package profile

import (
	"net/http"

	"github.com/bagaking/goulp/wlog"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"

	"github.com/bagaking/memorianexus/internal/utils"
	"github.com/bagaking/memorianexus/src/model"
	"github.com/bagaking/memorianexus/src/module/dto"
)

// ReqUpdateUserSettingsMemorization defines the request format for updating user settings.
type ReqUpdateUserSettingsMemorization struct {
	ReviewInterval       *uint   `json:"review_interval"`
	DifficultyPreference *uint8  `json:"difficulty_preference"`
	QuizMode             *string `json:"quiz_mode"`
}

// ReqUpdateUserSettingsAdvance defines the request to update advanced settings.
type ReqUpdateUserSettingsAdvance struct {
	Theme              *string `json:"theme"`
	Language           *string `json:"language"`
	EmailNotifications *bool   `json:"email_notifications"`
	PushNotifications  *bool   `json:"push_notifications"`
}

// GetUserSettingsMemorization handles a request to get the current user's settings.
// @Summary Get user settings
// @Description Retrieves settings information for the user who made the request.
// @TagNames profile
// @Produce  json
// @Security ApiKeyAuth
// @Success 200 {object} dto.RespSettingsMemorization "Successfully retrieved user settings"
// @Failure 400 {object} utils.ErrorResponse "Bad Request"
// @Failure 404 {object} utils.ErrorResponse "Not Found"
// @Failure 500 {object} utils.ErrorResponse "Internal Server Error"
// @Router /profile/settings/memorization [get]
func (svr *Service) GetUserSettingsMemorization(c *gin.Context) {
	log := wlog.ByCtx(c)
	userID, exists := utils.GetUIDFromGinCtx(c)
	if !exists {
		utils.GinHandleError(c, log, http.StatusUnauthorized, nil, "User not authenticated")
		return
	}

	profile, err := model.EnsureLoadProfile(svr.db, userID)
	if err != nil {
		utils.GinHandleError(c, log, http.StatusNotFound, err, "Profile not found")
		return
	}

	settings, err := profile.EnsureLoadProfileSettingsMemorization(svr.db)
	if err != nil {
		utils.GinHandleError(c, log, http.StatusInternalServerError, err, "Failed to retrieve profile settings")
		return
	}

	new(dto.RespSettingsMemorization).With(new(dto.SettingsMemorization).FromModel(settings)).Response(c)
}

// GetUserSettingsAdvance retrieves advanced settings for the authenticated user.
// @Summary Get user advanced settings
// @Description Retrieves advanced settings information for the current user.
// @TagNames profile
// @Produce  json
// @Security ApiKeyAuth
// @Success 200 {object} dto.RespSettingsAdvance "Successfully retrieved user advanced settings"
// @Failure 400 {object} utils.ErrorResponse "Bad Request"
// @Failure 404 {object} utils.ErrorResponse "Not Found"
// @Failure 500 {object} utils.ErrorResponse "Internal Server Error"
// @Router /profile/settings/advance [get]
func (svr *Service) GetUserSettingsAdvance(c *gin.Context) {
	log := wlog.ByCtx(c)
	userID, exists := utils.GetUIDFromGinCtx(c)
	if !exists {
		utils.GinHandleError(c, log, http.StatusUnauthorized, nil, "User not authenticated")
		return
	}

	profile, err := model.EnsureLoadProfile(svr.db, userID)
	if err != nil || profile == nil {
		utils.GinHandleError(c, log, http.StatusNotFound, err, "Profile not found")
		return
	}

	// Assuming EnsureLoadProfileSettingsAdvance will either load or create if not exists.
	advanceSettings, err := profile.EnsureLoadProfileSettingsAdvance(svr.db)
	if err != nil {
		utils.GinHandleError(c, log, http.StatusInternalServerError, err, "Failed to retrieve advanced settings")
		return
	}

	new(dto.RespSettingsAdvance).With(new(dto.SettingsAdvance).FromModel(advanceSettings)).Response(c)
}

// UpdateUserSettingsMemorization handles a request to update the current user's settings.
// @Summary Update user settings
// @Description Updates the settings for the user who made the request.
// @TagNames profile
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param settings body ReqUpdateUserSettingsMemorization true "User settings update info"
// @Success 200 {object} dto.RespSettingsMemorization "Successfully updated user settings"
// @Failure 400 {object} utils.ErrorResponse "Bad Request"
// @Failure 404 {object} utils.ErrorResponse "Not Found"
// @Failure 500 {object} utils.ErrorResponse "Internal Server Error"
// @Router /profile/settings/memorization [put]
func (svr *Service) UpdateUserSettingsMemorization(c *gin.Context) {
	log := wlog.ByCtx(c)
	userID, exists := utils.GetUIDFromGinCtx(c)
	if !exists {
		utils.GinHandleError(c, log, http.StatusUnauthorized, nil, "User not authenticated")
		return
	}

	var updateReq ReqUpdateUserSettingsMemorization
	if err := c.ShouldBindWith(&updateReq, binding.JSON); err != nil {
		utils.GinHandleError(c, log, http.StatusBadRequest, err, "Invalid request body")
		return
	}

	profile, err := model.EnsureLoadProfile(svr.db, userID)
	if err != nil {
		utils.GinHandleError(c, log, http.StatusNotFound, err, "Profile not found")
		return
	}

	settingsToUpdate := &model.ProfileMemorizationSetting{
		ID: userID,
	}

	// Update the fields that were provided in the request.
	if updateReq.ReviewInterval != nil {
		settingsToUpdate.ReviewInterval = *updateReq.ReviewInterval
	}
	if updateReq.DifficultyPreference != nil {
		settingsToUpdate.DifficultyPreference = *updateReq.DifficultyPreference
	}
	if updateReq.QuizMode != nil {
		settingsToUpdate.QuizMode = *updateReq.QuizMode
	}

	if err = profile.UpdateSettingsMemorization(svr.db, settingsToUpdate); err != nil {
		utils.GinHandleError(c, log, http.StatusInternalServerError, err, "Failed to update settings")
		return
	}

	new(dto.RespSettingsMemorization).With(
		new(dto.SettingsMemorization).FromModel(settingsToUpdate),
	).Response(c, "settings updated successfully")
}

// UpdateUserSettingsAdvance updates the advanced settings for the current user.
// @Summary Update user advanced settings
// @Description Updates advanced settings for the authenticated user.
// @TagNames profile
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param settings body ReqUpdateUserSettingsAdvance true "User advanced settings update info"
// @Success 200 {object} dto.RespSettingsAdvance "Successfully updated user advanced settings"
// @Failure 400 {object} utils.ErrorResponse "Bad Request"
// @Failure 404 {object} utils.ErrorResponse "Not Found"
// @Failure 500 {object} utils.ErrorResponse "Internal Server Error"
// @Router /profile/settings/advance [put]
func (svr *Service) UpdateUserSettingsAdvance(c *gin.Context) {
	log := wlog.ByCtx(c)
	userID, exists := utils.GetUIDFromGinCtx(c)
	if !exists {
		utils.GinHandleError(c, log, http.StatusUnauthorized, nil, "User not authenticated")
		return
	}

	var updateReq ReqUpdateUserSettingsAdvance
	if err := c.ShouldBindWith(&updateReq, binding.JSON); err != nil {
		utils.GinHandleError(c, log, http.StatusBadRequest, err, "Invalid request body")
		return
	}

	profile, err := model.EnsureLoadProfile(svr.db, userID)
	if err != nil || profile == nil {
		utils.GinHandleError(c, log, http.StatusNotFound, err, "Profile not found")
		return
	}

	advanceSettings := &model.ProfileAdvanceSetting{
		ID: userID,
	}

	// Update the fields that were provided in the request.
	if updateReq.Theme != nil {
		advanceSettings.Theme = *updateReq.Theme
	}
	if updateReq.Language != nil {
		advanceSettings.Language = *updateReq.Language
	}
	if updateReq.EmailNotifications != nil {
		advanceSettings.EmailNotifications = *updateReq.EmailNotifications
	}
	if updateReq.PushNotifications != nil {
		advanceSettings.PushNotifications = *updateReq.PushNotifications
	}

	if err = profile.UpdateSettingsAdvance(svr.db, advanceSettings); err != nil {
		utils.GinHandleError(c, log, http.StatusInternalServerError, err, "Failed to update advanced settings")
		return
	}

	new(dto.RespSettingsAdvance).With(
		new(dto.SettingsAdvance).FromModel(advanceSettings),
	).Response(c, "advanced settings updated successfully")
}
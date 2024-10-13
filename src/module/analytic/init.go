package analytic

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/bagaking/memorianexus/src/module/dto"
)

type Service struct {
	// db *model.db
}

type StudyPattern struct {
	Label       string  `json:"label"`
	ReviewCount int     `json:"review_count"`
	Accuracy    float64 `json:"accuracy"`
}

type StudyPatternsResponse struct {
	Patterns []StudyPattern `json:"patterns"`
}

type TimeSpentBucket struct {
	Period  string `json:"period"`
	Minutes int    `json:"minutes"`
}

type TimeSpentResponse struct {
	Buckets      []TimeSpentBucket `json:"buckets"`
	TotalMinutes int               `json:"total_minutes"`
}

var svr *Service

func Init(db *gorm.DB) (*Service, error) {
	svr = &Service{
		// db: model.NewRepo(db),
	}
	return svr, nil
}

func (svr *Service) ApplyMux(group gin.IRouter) {
	group.GET("/studyPatterns", svr.GetStudyPatterns)
	group.GET("/timeSpent", svr.GetTimeSpent)
}

func (svr *Service) GetStudyPatterns(c *gin.Context) {
	new(dto.RespSuccess[StudyPatternsResponse]).
		With(StudyPatternsResponse{Patterns: []StudyPattern{}}).
		Response(c, "study patterns scaffold")
}

func (svr *Service) GetTimeSpent(c *gin.Context) {
	new(dto.RespSuccess[TimeSpentResponse]).
		With(TimeSpentResponse{Buckets: []TimeSpentBucket{}}).
		Response(c, "time spent scaffold")
}

package router

import (
	"net/http"
	"server/database"
	"server/funcs"
	"server/models"
	"server/ranking"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type LoginAdminRequest struct {
	Password string `json:"password"`
}

// POST /admin/auth - Checks if an admin is authenticated
func AdminAuthenticated(ctx *gin.Context) {
	// This route will run the middleware first, and if the middleware
	// passes, then that means the admin is authenticated

	ctx.JSON(http.StatusOK, gin.H{"yes_no": 1})
}

// GET /admin/stats - GetAdminStats returns stats about the system
func GetAdminStats(ctx *gin.Context) {
	// Get the database from the context
	db := ctx.MustGet("db").(*mongo.Database)

	// Aggregate the stats
	stats, err := database.AggregateStats(db)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error aggregating stats: " + err.Error()})
		return
	}

	// Send OK
	ctx.JSON(http.StatusOK, stats)
}

// GET /admin/clock - GetClock returns the current clock state
func GetClock(ctx *gin.Context) {
	// Get the database from the context
	db := ctx.MustGet("db").(*mongo.Database)

	// Get the clock from the context
	clock := ctx.MustGet("clock").(*models.ClockState)

	// Save the options in the database
	err := database.UpdateClock(db, clock)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error saving options: " + err.Error()})
		return
	}

	// Send OK
	ctx.JSON(http.StatusOK, gin.H{"running": clock.Running, "time": clock.GetDuration()})
}

func PauseClock(ctx *gin.Context) *models.ClockState {
	// Get the database from the context
	db := ctx.MustGet("db").(*mongo.Database)

	// Get the clock from the context
	clock := ctx.MustGet("clock").(*models.ClockState)

	// Pause the clock
	clock.Pause()

	// Save the clock in the database
	err := database.UpdateClock(db, clock)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error saving clock: " + err.Error()})
		return nil
	}
	return clock
}

// POST /admin/clock/pause - PauseClockHandler pauses the clock
func PauseClockHandler(ctx *gin.Context) {
	clock := PauseClock(ctx)

	if clock != nil {
		// Send OK
		ctx.JSON(http.StatusOK, gin.H{"clock": clock})
	}
}

// POST /admin/clock/unpause - UnpauseClock unpauses the clock
func UnpauseClock(ctx *gin.Context) {
	// Get the database from the context
	db := ctx.MustGet("db").(*mongo.Database)

	// Check if judging has ended
	judgingEnded, err := database.GetJudgingEnded(db)
	if judgingEnded {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Judging has been ended. Clock cannot be unpaused."})
		return
	}

	// Get the clock from the context
	clock := ctx.MustGet("clock").(*models.ClockState)

	// Unpause the clock
	clock.Resume()

	// Save the clock in the database
	err = database.UpdateClock(db, clock)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error saving clock: " + err.Error()})
		return
	}

	// Send OK
	ctx.JSON(http.StatusOK, gin.H{"clock": clock})
}

// POST /admin/clock/reset - ResetClock resets the clock
func ResetClock(ctx *gin.Context) {
	// Get the database from the context
	db := ctx.MustGet("db").(*mongo.Database)

	// Get the clock from the context
	clock := ctx.MustGet("clock").(*models.ClockState)

	// Reset the clock
	clock.Reset()

	// Save the clock in the database
	err := database.UpdateClock(db, clock)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error saving clock: " + err.Error()})
		return
	}

	// Send OK
	ctx.JSON(http.StatusOK, gin.H{"clock": clock, "yes_no": 1})
}

func IsClockPaused(ctx *gin.Context) {
	// Get the clock from the context
	clock := ctx.MustGet("clock").(*models.ClockState)

	// Send OK
	if clock.Running {
		ctx.JSON(http.StatusOK, gin.H{"yes_no": 1})
	} else {
		ctx.JSON(http.StatusOK, gin.H{"yes_no": 0})
	}
}

// POST /admin/reset - ResetDatabase resets the database
func ResetDatabase(ctx *gin.Context) {
	// Get the database from the context
	db := ctx.MustGet("db").(*mongo.Database)

	// Reset the database
	err := database.DropAll(db)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error resetting database: " + err.Error()})
		return
	}

	// Send OK
	ctx.JSON(http.StatusOK, gin.H{"yes_no": 1})
}

// POST /admin/flags - GetFlags returns all flags
func GetFlags(ctx *gin.Context) {
	// Get the database from the context
	db := ctx.MustGet("db").(*mongo.Database)

	// Get all the flags
	flags, err := database.FindAllFlags(db)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error getting flags: " + err.Error()})
		return
	}

	// Send OK
	ctx.JSON(http.StatusOK, flags)
}

func GetOptions(ctx *gin.Context) {
	// Get the database from the context
	db := ctx.MustGet("db").(*mongo.Database)

	// Get the options
	options, err := database.GetOptions(db)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error getting options: " + err.Error()})
		return
	}

	// Send OK
	ctx.JSON(http.StatusOK, options)
}

// POST /admin/export/projects - ExportProjects exports all projects to a CSV
func ExportProjects(ctx *gin.Context) {
	// Get the database from the context
	db := ctx.MustGet("db").(*mongo.Database)

	// Get all the projects
	projects, err := database.FindAllProjects(db)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error getting projects: " + err.Error()})
		return
	}
	err, errStr, scores := ranking.GetScoresFromDB(db)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": errStr + err.Error()})
		return
	}

	// Create the CSV
	csvData := funcs.CreateProjectCSV(projects, scores)

	// Send CSV
	funcs.AddCsvData("projects", csvData, ctx)
}

// POST /admin/export/challenges - ExportProjectsByChallenge exports all projects to a zip file, with CSVs each
// containing projects that only belong to a single challenge
func ExportProjectsByChallenge(ctx *gin.Context) {
	// Get the database from the context
	db := ctx.MustGet("db").(*mongo.Database)

	// Get all the projects
	projects, err := database.FindAllProjects(db)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error getting projects: " + err.Error()})
		return
	}
	err, errStr, scores := ranking.GetScoresFromDB(db)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": errStr + err.Error()})
		return
	}

	// Create the zip file
	zipData, err := funcs.CreateProjectChallengeZip(projects, scores)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error creating zip file: " + err.Error()})
		return
	}

	// Send zip file
	funcs.AddZipFile("projects", zipData, ctx)
}

// POST /admin/export/rankings - ExportRankings exports the rankings of each judge as a CSV
func ExportRankings(ctx *gin.Context) {
	// Get the database from the context
	db := ctx.MustGet("db").(*mongo.Database)

	// Get all the judges
	judges, err := database.FindAllJudges(db)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error getting judges: " + err.Error()})
		return
	}

	// Create the CSV
	csvData := funcs.CreateJudgeRankingCSV(judges)

	// Send CSV
	funcs.AddCsvData("rankings", csvData, ctx)
}

// GET /admin/timer - GetJudgingTimer returns the judging timer
func GetJudgingTimer(ctx *gin.Context) {
	// Get the database from the context
	db := ctx.MustGet("db").(*mongo.Database)

	// Get the options
	options, err := database.GetOptions(db)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error getting options: " + err.Error()})
		return
	}

	// Send OK
	ctx.JSON(http.StatusOK, gin.H{"judging_timer": options.JudgingTimer})
}

type SetJudgingTimerRequest struct {
	JudgingTimer int64 `json:"judging_timer"`
}

// POST /admin/timer - SetJudgingTimer sets the judging timer
func SetJudgingTimer(ctx *gin.Context) {
	// Get the database from the context
	db := ctx.MustGet("db").(*mongo.Database)

	// Get request
	var req SetJudgingTimerRequest
	err := ctx.BindJSON(&req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "error parsing request: " + err.Error()})
		return
	}

	// Get the options
	options, err := database.GetOptions(db)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error getting options: " + err.Error()})
		return
	}

	options.JudgingTimer = req.JudgingTimer

	// Save the options in the database
	err = database.UpdateOptions(db, options)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error saving options: " + err.Error()})
		return
	}

	// Send OK
	ctx.JSON(http.StatusOK, gin.H{"yes_no": 1})
}

type SetCategoriesRequest struct {
	Categories []string `json:"categories"`
}

// POST /admin/categories - sets the categories
func SetCategories(ctx *gin.Context) {
	// Get the database from the context
	db := ctx.MustGet("db").(*mongo.Database)

	// Get the categories
	var categoriesReq SetCategoriesRequest
	err := ctx.BindJSON(&categoriesReq)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "error parsing request: " + err.Error()})
		return
	}

	// Save the categories in the database
	err = database.UpdateCategories(db, categoriesReq.Categories)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error saving categories: " + err.Error()})
		return
	}

	// Send OK
	ctx.JSON(http.StatusOK, gin.H{"yes_no": 1})
}

type MinViewsRequest struct {
	MinViews int `json:"min_views"`
}

type BatchRankingSizeRequest struct {
	BRS int `json:"batch_ranking_size"`
}

// POST /admin/min-views - sets the min views
func SetMinViews(ctx *gin.Context) {
	// Get the database from the context
	db := ctx.MustGet("db").(*mongo.Database)

	// Get the views
	var minViewsReq MinViewsRequest
	err := ctx.BindJSON(&minViewsReq)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "error parsing request: " + err.Error()})
	}

	// Save the min views in the db
	err = database.UpdateMinViews(db, minViewsReq.MinViews)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error saving min views: " + err.Error()})
		return
	}

	// Send OK
	ctx.JSON(http.StatusOK, gin.H{"yes_no": 1})
}

// POST /admin/batch-ranking-size - sets the ranking batch size
func SetRankingBatchSize(ctx *gin.Context) {
	// Get the database from the context
	db := ctx.MustGet("db").(*mongo.Database)

	// Get the views
	var brsReq BatchRankingSizeRequest
	err := ctx.BindJSON(&brsReq)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "error parsing request: " + err.Error()})
	}

	// Save the ranking batch size in the db
	err = database.UpdateBatchRankingSize(db, brsReq.BRS)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error saving batch ranking size: " + err.Error()})
		return
	}

	// Send OK
	ctx.JSON(http.StatusOK, gin.H{"yes_no": 1})
}

// GET /admin/score - GetScores returns the calculated scores of all projects
func GetScores(ctx *gin.Context) {
	// Get the database from the context
	db := ctx.MustGet("db").(*mongo.Database)

	err, errStr, scores := ranking.GetScoresFromDB(db)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": errStr + err.Error()})
		return
	}

	// Send OK
	ctx.JSON(http.StatusOK, scores)
}

// GET /check-judging-over - isJudgingEnded returns a yes_no indicating the value of the judging_ended boolean flag
func isJudgingEnded(ctx *gin.Context) {
	db := ctx.MustGet("db").(*mongo.Database)

	// Get judging_ended flag from database
	judgingEnded, err := database.GetJudgingEnded(db)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error getting judging_ended flag: " + err.Error()})
		return
	}

	// no type conversion from bool to int directly :'( : https://stackoverflow.com/a/38627381/7253717
	var judgingEndedVar int8
	if judgingEnded {
		judgingEndedVar = 1
	}

	ctx.JSON(http.StatusOK, gin.H{"yes_no": judgingEndedVar})
}

// POST /admin/end-judging - endJudging ends the judging process by setting the judging_ended flag to true
func endJudging(ctx *gin.Context) {
	// Get the database from the context
	db := ctx.MustGet("db").(*mongo.Database)

	// Pause the clock
	clock := PauseClock(ctx)

	if clock == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error pausing clock"})
		return
	}

	// Save the judging_ended flag in the db
	err := database.SetEndJudging(db)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error setting judging_ended: " + err.Error()})
		return
	}

	// Send OK
	ctx.JSON(http.StatusOK, gin.H{"yes_no": 1})
}

// contains checks if a string is in a list of strings
func contains(list []primitive.ObjectID, str primitive.ObjectID) bool {
	for _, s := range list {
		if s == str {
			return true
		}
	}
	return false
}

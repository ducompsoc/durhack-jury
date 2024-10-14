package router

import (
	"log"
	"net/url"
	"server/config"
	"server/database"
	"server/judging"
	"server/models"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	sessionsMongodriver "github.com/gin-contrib/sessions/mongo/mongodriver"
	"github.com/gin-contrib/static"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

// Creates a new Gin router with all routes defined
func NewRouter(db *mongo.Database) *gin.Engine {
	// Create the router
	router := gin.Default()

	// Get the clock state from the database
	clock := getClockFromDb(db)

	// Create the comparisons object
	comps, err := judging.LoadComparisons(db)
	if err != nil {
		log.Fatalf("error loading projects from the database: %s\n", err.Error())
	}

	// Add shared variables to router
	router.Use(useVar("db", db))
	router.Use(useVar("clock", &clock))
	router.Use(useVar("comps", comps))

	// CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{config.Origin},
		AllowMethods:     []string{"GET", "POST", "DELETE", "OPTIONS", "PUT"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token", "Authorization", "Accept", "Cache-Control", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * 3600,
	}))

	// Sessions
	sessionCollection := db.Collection("sessions")
	store := sessionsMongodriver.NewStore(sessionCollection, 60*60, true, []byte("secret"))
	parsedUrl, err := url.Parse(config.ApiOrigin)
	if err != nil {
		log.Fatalf("error parsing host url: %s\n", err.Error())
	}
	store.Options(sessions.Options{
		HttpOnly: true,
		Secure:   false,
		Domain:   parsedUrl.Host,
		Path:     "/",
	})
	router.Use(sessions.Sessions("durhack-jury-session", store))

	// Create router groups for judge and admins
	// lucatodo: document routing behaviour r.e. login and auth
	authenticatedRouter := router.Group("", Authenticate())
	judgeRouter := authenticatedRouter.Group("/api", AuthoriseJudge())
	adminRouter := authenticatedRouter.Group("/api", AuthoriseAdmin())
	defaultRouter := router.Group("/api")

	// lucatodo: improved error handling middleware: https://stackoverflow.com/questions/69948784/how-to-handle-errors-in-gin-middleware/69948929#69948929
	// Authenticated login routes
	defaultRouter.GET("/auth/keycloak/login", BeginKeycloakOAuth2Flow())
	defaultRouter.GET("/auth/keycloak/callback", KeycloakOAuth2FlowCallback(), HandleLoginSuccess())
	authenticatedRouter.GET("/api/auth/keycloak/logout", Logout())

	// Add routes
	judgeRouter.GET("/judge", GetJudge)
	judgeRouter.POST("/judge/auth", JudgeAuthenticated)
	judgeRouter.GET("/judge/welcome", CheckJudgeReadWelcome)
	judgeRouter.POST("/judge/welcome", SetJudgeReadWelcome)
	adminRouter.GET("/judge/list", ListJudges)
	adminRouter.GET("/judge/stats", JudgeStats)
	adminRouter.DELETE("/judge/:id", DeleteJudge)
	judgeRouter.GET("/judge/projects", GetJudgeProjects)
	judgeRouter.POST("/judge/next", GetNextJudgeProject)
	judgeRouter.POST("/judge/skip", JudgeSkip)
	judgeRouter.POST("/judge/score", JudgeScore)
	judgeRouter.POST("/judge/rank", JudgeRank)
	judgeRouter.POST("/judge/submit_batch_ranking", JudgeSubmitBatchRanking)
	judgeRouter.PUT("/judge/score", JudgeUpdateScore)
	judgeRouter.POST("/judge/break", JudgeBreak)

	adminRouter.POST("/project/devpost", AddDevpostCsv)
	adminRouter.POST("/project/new", AddProject)
	adminRouter.GET("/project/list", ListProjects)
	defaultRouter.GET("/project/list/public", ListPublicProjects)
	adminRouter.POST("/project/csv", AddProjectsCsv)
	judgeRouter.GET("/project/:id", GetProject)
	judgeRouter.GET("/project/count", GetProjectCount)
	judgeRouter.GET("/judge/project/:id", GetJudgedProject)
	adminRouter.DELETE("/project/:id", DeleteProject)
	adminRouter.GET("/project/stats", ProjectStats)

	adminRouter.GET("/admin/stats", GetAdminStats)
	adminRouter.GET("/admin/score", GetScores)
	adminRouter.GET("/admin/clock", GetClock)
	adminRouter.POST("/admin/clock/pause", PauseClock)
	adminRouter.POST("/admin/clock/unpause", UnpauseClock)
	adminRouter.POST("/admin/clock/reset", ResetClock)
	adminRouter.POST("/admin/auth", AdminAuthenticated)
	adminRouter.POST("/admin/reset", ResetDatabase)
	adminRouter.POST("/judge/hide", HideJudge)
	adminRouter.POST("/judge/unhide", UnhideJudge)
	adminRouter.POST("/project/hide", HideProject)
	adminRouter.POST("/project/unhide", UnhideProject)
	adminRouter.POST("/project/prioritize", PrioritizeProject)
	adminRouter.POST("/project/unprioritize", UnprioritizeProject)
	adminRouter.PUT("/judge/:id", EditJudge)
	defaultRouter.GET("/admin/started", IsClockPaused)
	adminRouter.GET("/admin/flags", GetFlags)
	adminRouter.POST("/project/reassign", ReassignProjectNums)
	adminRouter.GET("/admin/options", GetOptions)
	adminRouter.GET("/admin/export/projects", ExportProjects)
	adminRouter.GET("/admin/export/challenges", ExportProjectsByChallenge)
	adminRouter.GET("/admin/export/rankings", ExportRankings)
	judgeRouter.GET("/admin/timer", GetJudgingTimer)
	adminRouter.POST("/admin/timer", SetJudgingTimer)
	adminRouter.POST("/admin/min-views", SetMinViews)
	adminRouter.POST("/admin/ranking-batch-size", SetRankingBatchSize)
	adminRouter.POST("/admin/categories", SetCategories)
	judgeRouter.GET("/categories", GetCategories)
	judgeRouter.POST("/judge/notes", JudgeUpdateNotes)
	judgeRouter.GET("/rbs", GetRankingBatchSize)

	// Serve frontend static files
	router.Use(static.Serve("/assets", static.LocalFile("./public/assets", true)))
	router.StaticFile("/favicon.ico", "./public/favicon.ico")
	router.LoadHTMLFiles("./public/index.html")

	// Add no route handler
	router.NoRoute(func(ctx *gin.Context) {
		ctx.HTML(200, "index.html", nil)
	})

	return router
}

// useVar is a middleware that adds a variable to the context
func useVar(key string, v any) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Set(key, v)
		ctx.Next()
	}
}

// getClockFromDb gets the clock state from the database
// and on init will pause the clock
func getClockFromDb(db *mongo.Database) models.ClockState {
	// Get the clock state from the database
	options, err := database.GetOptions(db)
	if err != nil {
		log.Fatalln("error getting options: " + err.Error())
	}
	clock := options.Clock

	// Pause the clock
	clock.Pause()

	// Update the clock in the database
	err = database.UpdateClock(db, &clock)
	if err != nil {
		log.Fatalln("error updating clock: " + err.Error())
	}

	return clock
}

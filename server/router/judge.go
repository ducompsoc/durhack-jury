package router

import (
	"context"
	"encoding/json"
	"github.com/Nerzal/gocloak/v13"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/sync/errgroup"
	"net/http"
	"server/auth"
	"server/config"
	"server/database"
	"server/judging"
	"server/models"
	"strconv"
)

type GetJudgeResponse struct {
	models.Judge
	Email string `json:"email"`
	Name  string `json:"name"`
}

// GET /judge - Endpoint to get the judge from the token
func GetJudge(ctx *gin.Context) {
	// Get the judge from the context (See middleware.go)
	judge := ctx.MustGet("judge").(*models.Judge)
	// Get user info from separate authenticated user info created in earlier middleware: Authenticate()
	userInfo := ctx.MustGet("user").(*auth.DurHackKeycloakUserInfo)

	// Send Judge
	ctx.JSON(http.StatusOK, GetJudgeResponse{Judge: *judge, Email: userInfo.Email, Name: userInfo.GetNames()})
}

// POST /judge/auth - Check to make sure a judge is authenticated
func JudgeAuthenticated(ctx *gin.Context) {
	// This route will run the middleware first, and if the middleware
	// passes, then that means the judge is authenticated
	ctx.JSON(http.StatusOK, gin.H{"yes_no": 1})
}

// GET /judge/welcome - Endpoint to check if a judge has read the welcome message
func CheckJudgeReadWelcome(ctx *gin.Context) {
	// Get the judge from the context
	judge := ctx.MustGet("judge").(*models.Judge)

	// Send OK
	if judge.ReadWelcome {
		ctx.JSON(http.StatusOK, gin.H{"yes_no": 1})
	} else {
		ctx.JSON(http.StatusOK, gin.H{"yes_no": 0})
	}
}

// POST /judge/welcome - Endpoint to set a judge's readWelcome field to true
func SetJudgeReadWelcome(ctx *gin.Context) {
	// Get the database from the context
	db := ctx.MustGet("db").(*mongo.Database)

	// Get the judge from the context
	judge := ctx.MustGet("judge").(*models.Judge)

	// Set judge's readWelcome field to true
	judge.ReadWelcome = true

	// Update judge in database
	err := database.UpdateJudge(db, judge)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error updating judge in database: " + err.Error()})
		return
	}

	// Send OK
	ctx.JSON(http.StatusOK, gin.H{"yes_no": 1})
}

type judgeWithKeycloak struct {
	Judge          models.Judge `json:"judge"`
	PreferredNames *string      `json:"preferred_names"`
	FirstNames     string       `json:"first_names"`
	LastNames      string       `json:"last_names"`
}

func (j *judgeWithKeycloak) MarshalJSON() ([]byte, error) {
	type Alias judgeWithKeycloak
	return json.Marshal(&struct {
		*Alias
		LastActivity int64 `json:"last_activity"`
	}{
		Alias:        (*Alias)(j),
		LastActivity: int64(j.Judge.LastActivity),
	})
}

// GET /judge/list - Endpoint to get a list of all judges
func ListJudges(ctx *gin.Context) {
	// Get the database from the context
	db := ctx.MustGet("db").(*mongo.Database)

	// Get judges from database
	judges, err := database.FindAllJudges(db)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error finding judges in database: " + err.Error()})
		return
	}

	keycloakAdminClient := auth.KeycloakAdminClient

	judgesWithKeycloak := make([]*judgeWithKeycloak, len(judges))
	errGroup, _ := errgroup.WithContext(context.Background())
	for i, judge := range judges {
		errGroup.Go(func() error {
			judgeKeycloakInfo, err := getJudgeKeycloakInfo(keycloakAdminClient, judge.KeycloakUserId)
			if err != nil {
				return err
			}
			judgesWithKeycloak[i] = &judgeWithKeycloak{
				*judge,
				judgeKeycloakInfo.PreferredNames,
				judgeKeycloakInfo.FirstNames,
				judgeKeycloakInfo.LastNames,
			}
			return nil
		})
	}
	err = errGroup.Wait()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error getting judge info from keycloak: " + err.Error()})
		return
	}

	// Send OK
	ctx.JSON(http.StatusOK, judgesWithKeycloak)
}

type judgeKeycloak struct {
	PreferredNames *string
	FirstNames     string
	LastNames      string
}

func getJudgeKeycloakInfo(adminClient *gocloak.GoCloak, userID string) (*judgeKeycloak, error) {
	accessToken, err := auth.GetKeycloakAdminClientAccessToken(context.Background())
	if err != nil {
		return nil, err
	}

	user, err := adminClient.GetUserByID(context.Background(), *accessToken, config.KeycloakRealm, userID)
	if err != nil {
		return nil, err
	}
	userAttributes := *user.Attributes
	preferredNamesAttribute, preferredNamesAttributeExist := userAttributes["preferredNames"]
	var preferredNames *string
	if preferredNamesAttributeExist {
		preferredNames = &preferredNamesAttribute[0]
	} else {
		preferredNames = nil
	}
	attributes := judgeKeycloak{
		preferredNames,
		userAttributes["firstNames"][0],
		userAttributes["lastNames"][0],
	}
	return &attributes, nil
}

// GET /judge/stats - Endpoint to get stats about the judges
func JudgeStats(ctx *gin.Context) {
	// Get the database from the context
	db := ctx.MustGet("db").(*mongo.Database)

	// Aggregate judge stats
	stats, err := database.AggregateJudgeStats(db)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error aggregating judge stats: " + err.Error()})
		return
	}

	// Send OK
	ctx.JSON(http.StatusOK, stats)
}

// DELETE /judge/:id - Endpoint to delete a judge
func DeleteJudge(ctx *gin.Context) {
	// Get the database from the context
	db := ctx.MustGet("db").(*mongo.Database)

	// Get the judge ID from the URL
	judgeId := ctx.Param("id")

	// Convert judge ID string to ObjectID
	judgeObjectId, err := primitive.ObjectIDFromHex(judgeId)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid judge ID"})
		return
	}

	// Delete judge from database
	err = database.DeleteJudgeById(db, judgeObjectId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error deleting judge from database: " + err.Error()})
		return
	}

	// Send OK
	ctx.JSON(http.StatusOK, gin.H{"yes_no": 1})
}

// POST /judge/next - Endpoint to get the next project for a judge
func GetNextJudgeProject(ctx *gin.Context) {
	// Get the database from the context
	db := ctx.MustGet("db").(*mongo.Database)

	// If judging is over, return an empty object (prevents any picking of new projects)
	judgingEnded, err := database.GetJudgingEnded(db)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error getting judging_ended flag: " + err.Error()})
		return
	}
	if judgingEnded {
		ctx.JSON(http.StatusOK, gin.H{})
		return
	}

	// Get the judge from the context
	judge := ctx.MustGet("judge").(*models.Judge)

	// Get the comparisons from the context
	comps := ctx.MustGet("comps").(*judging.Comparisons)

	// If the judge already has a next project, return that project
	if judge.Current != nil {
		ctx.JSON(http.StatusOK, gin.H{"project_id": judge.Current.Hex()})
		return
	}

	// Otherwise, get the next project for the judge
	// TODO: This wrapping is a little ridiculous...
	var project *models.Project
	err = database.WithTransaction(db, func(ctx mongo.SessionContext) (interface{}, error) {
		var err error
		project, err = judging.PickNextProject(db, judge, ctx, comps)
		return nil, err
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error picking next project: " + err.Error()})
		return
	}

	// If there is no next project, return an empty object
	if project == nil {
		ctx.JSON(http.StatusOK, gin.H{})
		return
	}

	// Update judge and project
	err = database.UpdateAfterPicked(db, project, judge)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error updating next project in database: " + err.Error()})
		return
	}

	// Send OK and project ID
	ctx.JSON(http.StatusOK, gin.H{"project_id": project.Id.Hex()})
}

// GET /judge/projects - Endpoint to get a list of projects that a judge has seen
func GetJudgeProjects(ctx *gin.Context) {
	// Get the judge from the context
	judge := ctx.MustGet("judge").(*models.Judge)

	// Return the judge's seen projects list
	ctx.JSON(http.StatusOK, judge.SeenProjects)
}

type JudgedProjectWithUrl struct {
	models.JudgedProject
	Url string `bson:"url" json:"url"`
}

func addUrlToJudgedProject(project *models.JudgedProject, url string) *JudgedProjectWithUrl {
	return &JudgedProjectWithUrl{
		JudgedProject: models.JudgedProject{
			ProjectId:   project.ProjectId,
			Categories:  project.Categories,
			Name:        project.Name,
			Guild:       project.Guild,
			Location:    project.Location,
			Description: project.Description,
			Notes:       project.Notes,
		},
		Url: url,
	}
}

// GET /judge/project/:id - Gets a project that's been judged by ID
func GetJudgedProject(ctx *gin.Context) {
	// Get the judge from the context
	judge := ctx.MustGet("judge").(*models.Judge)

	// Get the database from the context
	db := ctx.MustGet("db").(*mongo.Database)

	// Get the project ID from the URL
	projectId := ctx.Param("id")

	// Search through the judge seen projects for the project ID
	for _, p := range judge.SeenProjects {
		if p.ProjectId.Hex() == projectId {
			// Add URL to judged project
			proj, err := database.FindProjectById(db, &p.ProjectId)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error getting project url: " + err.Error()})
				return
			}
			jpWithUrl := addUrlToJudgedProject(&p, proj.Url)

			// Parse and send JSON
			ctx.JSON(http.StatusOK, jpWithUrl)
			return
		}
	}

	// Send bad request bc project ID invalid
	ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid project ID"})
}

type SkipRequest struct {
	Reason string `json:"reason"`
}

// POST /judge/skip - Endpoint to skip a project
func JudgeSkip(ctx *gin.Context) {
	// Get the database from the context
	db := ctx.MustGet("db").(*mongo.Database)

	// Get the judge from the context
	judge := ctx.MustGet("judge").(*models.Judge)

	// Get the user info from the context to save the judge name
	judgeName := ctx.MustGet("user").(*auth.DurHackKeycloakUserInfo).GetNames()

	// Get the comparisons object
	comps := ctx.MustGet("comps").(*judging.Comparisons)

	// Get the skip reason from the request
	var skipReq SkipRequest
	err := ctx.BindJSON(&skipReq)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "error reading request body: " + err.Error()})
		return
	}

	// Skip the project
	err = judging.SkipCurrentProject(db, judge, judgeName, comps, skipReq.Reason, true)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Send OK
	ctx.JSON(http.StatusOK, gin.H{"yes_no": 1})
}

// POST /judge/hide - Endpoint to hide a judge
func HideJudge(ctx *gin.Context) {
	// Get the database from the context
	db := ctx.MustGet("db").(*mongo.Database)

	// Get ID from body
	var idReq models.IdRequest
	err := ctx.BindJSON(&idReq)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "error reading request body: " + err.Error()})
		return
	}

	// Convert ID string to ObjectID
	judgeObjectId, err := primitive.ObjectIDFromHex(idReq.Id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid judge ID"})
		return
	}

	// Hide judge in database
	err = database.SetJudgeHidden(db, &judgeObjectId, true)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error hiding judge in database: " + err.Error()})
		return
	}

	// Send OK
	ctx.JSON(http.StatusOK, gin.H{"yes_no": 1})
}

// POST /judge/unhide - Endpoint to unhide a judge
func UnhideJudge(ctx *gin.Context) {
	// Get the database from the context
	db := ctx.MustGet("db").(*mongo.Database)

	// Get ID from body
	var idReq models.IdRequest
	err := ctx.BindJSON(&idReq)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "error reading request body: " + err.Error()})
		return
	}
	id := idReq.Id

	// Convert ID string to ObjectID
	judgeObjectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid judge ID"})
		return
	}

	// Unhide judge in database
	err = database.SetJudgeHidden(db, &judgeObjectId, false)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error unhiding judge in database: " + err.Error()})
		return
	}

	// Send OK
	ctx.JSON(http.StatusOK, gin.H{"yes_no": 1})
}

// PUT /judge/:id - Endpoint to edit a judge
func EditJudge(ctx *gin.Context) {
	// Get the database from the context
	db := ctx.MustGet("db").(*mongo.Database)

	// Get the body content
	var judgeReq models.EditJudgeRequest
	err := ctx.BindJSON(&judgeReq)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "error reading request body: " + err.Error()})
		return
	}

	// Get the judge ID from the path
	judgeId := ctx.Param("id")

	// Convert ID string to ObjectID
	judgeObjectId, err := primitive.ObjectIDFromHex(judgeId)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid judge ID"})
		return
	}

	// Edit judge in database
	err = database.UpdateJudgeBasicInfo(db, &judgeObjectId, &judgeReq)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error editing judge in database: " + err.Error()})
		return
	}

	// Send OK
	ctx.JSON(http.StatusOK, gin.H{"yes_no": 1})
}

type JudgeScoreRequest struct {
	Categories map[string]int `json:"categories"`
}

// POST /judge/score - Endpoint to finish judging a project (give it a score in categories)
func JudgeScore(ctx *gin.Context) {
	// Get the database from the context
	db := ctx.MustGet("db").(*mongo.Database)

	// Get the judge from the context
	judge := ctx.MustGet("judge").(*models.Judge)

	// Get the request object
	var scoreReq JudgeScoreRequest
	err := ctx.BindJSON(&scoreReq)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "error reading request body: " + err.Error()})
		return
	}

	// Get the project from the database
	project, err := database.FindProjectById(db, judge.Current)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error finding project in database: " + err.Error()})
		return
	}

	// Create the judged project object
	judgedProject := models.JudgeProjectFromProject(project, scoreReq.Categories)

	// Update the project with the score
	err = database.UpdateAfterSeen(db, judge, judgedProject)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error storing scores in database: " + err.Error()})
		return
	}

	// Send OK
	ctx.JSON(http.StatusOK, gin.H{"yes_no": 1})
}

type RankRequest struct {
	Ranking []primitive.ObjectID `json:"ranking"`
}

// POST /judge/rank - Update the judge's ranking of projects
func JudgeRank(ctx *gin.Context) {
	// Get the database from the context
	db := ctx.MustGet("db").(*mongo.Database)

	// Get the judge from the context
	judge := ctx.MustGet("judge").(*models.Judge)

	// Get the request object
	var rankReq RankRequest
	err := ctx.BindJSON(&rankReq)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "error reading request body: " + err.Error()})
		return
	}

	// Update the judge's ranking
	err = database.UpdateJudgeRanking(db, judge, rankReq.Ranking)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error updating judge ranking in database: " + err.Error()})
		return
	}

	// Send OK
	ctx.JSON(http.StatusOK, gin.H{"yes_no": 1})
}

type BatchRankingRequest struct {
	BatchRanking []primitive.ObjectID `json:"batch_ranking"`
}

// POST /judge/submit-batch-ranking -
func JudgeSubmitBatchRanking(ctx *gin.Context) {
	// Get the database from the context
	db := ctx.MustGet("db").(*mongo.Database)

	// Get the judge from the context
	judge := ctx.MustGet("judge").(*models.Judge)

	// Get the request object
	var batchRankingReq BatchRankingRequest
	err := ctx.BindJSON(&batchRankingReq)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "error reading request body: " + err.Error()})
		return
	}

	// Validate batch ranking size
	brs, err := database.GetBatchRankingSize(db)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error getting batch ranking size: " + err.Error()})
		return
	}
	judgingOver, err := database.GetJudgingEnded(db)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error getting judging_ended flag: " + err.Error()})
		return
	}
	if !judgingOver && len(batchRankingReq.BatchRanking) != int(brs) { // If judging hasn't ended, the batch should be the size of the batch ranking size
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "batch should be at least " + strconv.FormatInt(brs, 10) + " (current BRS value) projects large."})
		return
	}
	if len(batchRankingReq.BatchRanking) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "batch ranking should be at least 1 project large."}) // note: can't enforce at least 2 since judging might be ended early
		return
	}

	// Update the judge's variables based on this new batch ranking
	err = database.UpdateJudgePostBatchRank(db, judge, batchRankingReq.BatchRanking)
	if err != nil {
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{"error": "error updating judge ranking in database post batch submission: " + err.Error()},
		)
		return
	}

	// Send OK
	ctx.JSON(http.StatusOK, gin.H{"yes_no": 1})
}

// POST /judge/break - Allows a judge to take a break and free up their current project
func JudgeBreak(ctx *gin.Context) {
	// Get the database from the context
	db := ctx.MustGet("db").(*mongo.Database)

	// Get the judge from the context
	judge := ctx.MustGet("judge").(*models.Judge)

	// Get the user info from the context to save the judge name
	judgeName := ctx.MustGet("user").(*auth.DurHackKeycloakUserInfo).GetNames()

	// Get the comparisons from the context
	comps := ctx.MustGet("comps").(*judging.Comparisons)

	// Error if the judge doesn't have a current project
	if judge.Current == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "judge doesn't have a current project"})
		return
	}

	// Basically skip the project for the judge
	err := judging.SkipCurrentProject(db, judge, judgeName, comps, "break", false)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error skipping project: " + err.Error()})
		return
	}

	// Send OK
	ctx.JSON(http.StatusOK, gin.H{"yes_no": 1})
}

// GET /categories - Endpoint to get the categories
func GetCategories(ctx *gin.Context) {
	// Get the database from the context
	db := ctx.MustGet("db").(*mongo.Database)

	// Get categories from database
	categories, err := database.GetCategories(db)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error getting categories: " + err.Error()})
		return
	}

	// Send OK
	ctx.JSON(http.StatusOK, categories)
}

// GET /brs - Endpoint to return ranking batch size
func GetRankingBatchSize(ctx *gin.Context) {
	// Get the database from the context
	db := ctx.MustGet("db").(*mongo.Database)

	// Get ranking batch size from database
	brs, err := database.GetBatchRankingSize(db)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error getting ranking batch size: " + err.Error()})
		return
	}

	// Send OK
	ctx.JSON(http.StatusOK, gin.H{"brs": brs})
}

type UpdateScoreRequest struct {
	Categories map[string]int     `json:"categories"`
	Project    primitive.ObjectID `json:"project"`
	Initial    bool               `json:"initial"`
}

// PUT /judge/score - Endpoint to update a judge's score for a certain project
func JudgeUpdateScore(ctx *gin.Context) {
	// Get the database from the context
	db := ctx.MustGet("db").(*mongo.Database)

	// Get the judge from the context
	judge := ctx.MustGet("judge").(*models.Judge)

	// Get the comparisons object from the context
	comps := ctx.MustGet("comps").(*judging.Comparisons)

	// Get the request object
	var scoreReq UpdateScoreRequest
	err := ctx.BindJSON(&scoreReq)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "error reading request body: " + err.Error()})
		return
	}

	// Find the index of the project in the judge's seen projects
	index := -1
	for i, p := range judge.SeenProjects {
		if p.ProjectId == scoreReq.Project {
			index = i
			break
		}
	}

	// If the project isn't in the judge's seen projects, return an error
	if index == -1 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "judge hasn't seen project or project is invalid"})
		return
	}

	// Update that specific index of the seen projects array
	judge.SeenProjects[index].Categories = scoreReq.Categories

	// Update the judge's score for the project
	err = database.UpdateJudgeSeenProjects(db, judge)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error updating judge score in database: " + err.Error()})
		return
	}

	// If this is the initial scoring, update the comparisons array
	comps.UpdateProjectComparisonCount(judge.SeenProjects, scoreReq.Project)

	// Send OK
	ctx.JSON(http.StatusOK, gin.H{"yes_no": 1})
}

type UpdateNotesRequest struct {
	Notes   string             `json:"notes"`
	Project primitive.ObjectID `json:"project"`
}

// POST /judge/notes - Update the notes of a judge
func JudgeUpdateNotes(ctx *gin.Context) {
	// Get the database from the context
	db := ctx.MustGet("db").(*mongo.Database)

	// Get the judge from the context
	judge := ctx.MustGet("judge").(*models.Judge)

	// Get the request object
	var scoreReq UpdateNotesRequest
	err := ctx.BindJSON(&scoreReq)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "error reading request body: " + err.Error()})
		return
	}

	// Find the index of the project in the judge's seen projects
	// TODO: Extract to diff function to get rid of repeated code from JudgeUpdateScore
	index := -1
	for i, p := range judge.SeenProjects {
		if p.ProjectId == scoreReq.Project {
			index = i
			break
		}
	}

	// If the project isn't in the judge's seen projects, return an error
	if index == -1 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "judge hasn't seen project or project is invalid"})
		return
	}

	// Update that specific index of the seen projects array
	judge.SeenProjects[index].Notes = scoreReq.Notes

	// Update the judge's object for the project
	err = database.UpdateJudgeSeenProjects(db, judge)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error updating judge score in database: " + err.Error()})
		return
	}

	// Send OK
	ctx.JSON(http.StatusOK, gin.H{"yes_no": 1})
}

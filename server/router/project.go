package router

import (
	"net/http"
	"server/database"
	"server/funcs"
	"server/models"
	"strings"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// POST /project/devpost - AddDevpostCsv adds a csv export from devpost to the database
func AddDevpostCsv(ctx *gin.Context) {
	// Get the database from the context
	db := ctx.MustGet("db").(*mongo.Database)

	// Get the CSV file from the request
	file, err := ctx.FormFile("csv")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "error reading CSV file from request: " + err.Error()})
		return
	}

	// Open the file
	f, err := file.Open()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error opening CSV file: " + err.Error()})
		return
	}

	// Read the file
	content := make([]byte, file.Size)
	_, err = f.Read(content)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error reading CSV file: " + err.Error()})
		return
	}

	// Parse the CSV file
	projects, err := funcs.ParseDevpostCSV(string(content))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "error parsing CSV file: " + err.Error()})
		return
	}

	// Insert projects into the database
	err = database.InsertProjects(db, projects)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error inserting judges into database: " + err.Error()})
		return
	}

	// Send OK
	ctx.JSON(http.StatusOK, gin.H{"yes_no": 1})
}

type AddProjectRequest struct {
	Name          string `json:"name"`
	Guild         string `json:"guild"`
	Location      string `json:"location"`
	Description   string `json:"description"`
	Url           string `json:"url"`
	TryLink       string `json:"tryLink"`
	VideoLink     string `json:"videoLink"`
	ChallengeList string `json:"challengeList"`
}

// POST /project/new - AddProject adds a project to the database
func AddProject(ctx *gin.Context) {
	// Get the database from the context
	db := ctx.MustGet("db").(*mongo.Database)

	// Get the projectReq from the request
	var projectReq AddProjectRequest
	err := ctx.BindJSON(&projectReq)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "error binding project from request: " + err.Error()})
		return
	}

	// Make sure name, description, and url are defined
	if projectReq.Name == "" || projectReq.Description == "" || projectReq.Url == "" || projectReq.Location == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "name, description, url and a location are required"})
		return
	}

	// Get the challenge list
	challengeList := strings.Split(projectReq.ChallengeList, ",")
	if projectReq.ChallengeList == "" {
		challengeList = []string{}
	}
	for i := range challengeList {
		challengeList[i] = strings.TrimSpace(challengeList[i])
	}

	// Create the project
	project := models.NewProject(projectReq.Name, projectReq.Guild, projectReq.Location, projectReq.Description, projectReq.Url, projectReq.TryLink, projectReq.VideoLink, challengeList)

	// Insert project and update the next table num field in options
	err = database.WithTransaction(db, func(ctx mongo.SessionContext) (interface{}, error) {
		// Insert project
		err := database.InsertProject(db, project)
		return nil, err
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error inserting project into database: " + err.Error()})
		return
	}

	// Send OK
	ctx.JSON(http.StatusOK, gin.H{"yes_no": 1})
}

// GET /project/list - ListProjects lists all projects in the database
func ListProjects(ctx *gin.Context) {
	// Get the database from the context
	db := ctx.MustGet("db").(*mongo.Database)

	// Get the projects from the database
	projects, err := database.FindAllProjects(db)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error getting projects from database: " + err.Error()})
		return
	}

	// Send OK
	ctx.JSON(http.StatusOK, projects)
}

type PublicProject struct {
	Name          string `json:"name"`
	Location      string `json:"location"`
	Description   string `json:"description"`
	Url           string `json:"url"`
	TryLink       string `json:"tryLink"`
	VideoLink     string `json:"videoLink"`
	ChallengeList string `json:"challengeList"`
}

func ListPublicProjects(ctx *gin.Context) {
	// Get the database from the context
	db := ctx.MustGet("db").(*mongo.Database)

	// Get the projects from the database
	projects, err := database.FindAllProjects(db)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error getting projects from database: " + err.Error()})
		return
	}

	// Convert projects to public projects
	publicProjects := make([]PublicProject, len(projects))
	for i, project := range projects {
		publicProjects[i] = PublicProject{
			Name:          project.Name,
			Location:      project.Location,
			Description:   project.Description,
			Url:           project.Url,
			TryLink:       project.TryLink,
			VideoLink:     project.VideoLink,
			ChallengeList: strings.Join(project.ChallengeList, ", "),
		}
	}

	// Send OK
	ctx.JSON(http.StatusOK, publicProjects)
}

// POST /project/csv - Endpoint to add projects from a CSV file
func AddProjectsCsv(ctx *gin.Context) {
	// Get the database from the context
	db := ctx.MustGet("db").(*mongo.Database)

	// Get the CSV file from the request
	file, err := ctx.FormFile("csv")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "error reading CSV file from request: " + err.Error()})
		return
	}

	// Get the hasHeader parameter from the request
	hasHeader := ctx.PostForm("hasHeader") == "true"

	// Open the file
	f, err := file.Open()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error opening CSV file: " + err.Error()})
		return
	}

	// Read the file
	content := make([]byte, file.Size)
	_, err = f.Read(content)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error reading CSV file: " + err.Error()})
		return
	}

	// Parse the CSV file
	projects, err := funcs.ParseProjectCsv(string(content), hasHeader)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "error parsing CSV file: " + err.Error()})
		return
	}

	// Insert projects into the database
	err = database.InsertProjects(db, projects)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error inserting projects into database: " + err.Error()})
		return
	}

	// Send OK
	ctx.JSON(http.StatusOK, gin.H{"yes_no": 1})
}

// DELETE /project/:id - DeleteProject deletes a project from the database
func DeleteProject(ctx *gin.Context) {
	// Get the database from the context
	db := ctx.MustGet("db").(*mongo.Database)

	// Get the id from the request
	id := ctx.Param("id")

	// Convert judge ID string to ObjectID
	projectObjectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid project ID"})
		return
	}

	// Delete the project from the database
	err = database.DeleteProjectById(db, projectObjectId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error deleting project from database: " + err.Error()})
		return
	}

	// Send OK
	ctx.JSON(http.StatusOK, gin.H{"yes_no": 1})
}

// POST /project/stats - ProjectStats returns stats about projects
func ProjectStats(ctx *gin.Context) {
	// Get the database from the context
	db := ctx.MustGet("db").(*mongo.Database)

	// Aggregate project stats
	stats, err := database.AggregateProjectStats(db)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error aggregating project stats: " + err.Error()})
		return
	}

	// Send OK
	ctx.JSON(http.StatusOK, stats)
}

// GET /project/:id - GetProject returns a project by ID
func GetProject(ctx *gin.Context) {
	// Get the database from the context
	db := ctx.MustGet("db").(*mongo.Database)

	// Get the id from the request
	id := ctx.Param("id")

	// Convert project ID string to ObjectID
	projectObjectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid project ID"})
		return
	}

	// Get the project from the database
	project, err := database.FindProjectById(db, &projectObjectId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error getting project from database: " + err.Error()})
		return
	}

	// Send OK
	ctx.JSON(http.StatusOK, project)
}

// GET /project/count - GetProjectCount returns the number of projects in the database
func GetProjectCount(ctx *gin.Context) {
	// Get the database from the context
	db := ctx.MustGet("db").(*mongo.Database)

	// Get the project from the database
	count, err := database.CountProjectDocuments(db)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error getting project count from database: " + err.Error()})
		return
	}

	// Send OK
	ctx.JSON(http.StatusOK, gin.H{"count": count})
}

// POST /project/hide - HideProject hides a project
func HideProject(ctx *gin.Context) {
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

	// Convert project ID string to ObjectID
	projectObjectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid project ID"})
		return
	}

	// Update the project in the database
	err = database.SetProjectHidden(db, &projectObjectId, true)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error updating project in database: " + err.Error()})
		return
	}

	// Send OK
	ctx.JSON(http.StatusOK, gin.H{"yes_no": 1})
}

// POST /project/hide-unhide-many - HideUnhideManyProjects hides projects in bulk (used for guilds being away)
func HideUnhideManyProjects(ctx *gin.Context) {
	// Get the database from the context
	db := ctx.MustGet("db").(*mongo.Database)

	// Get ID from body
	var multiHideReq models.MultiIdHideRequest
	err := ctx.BindJSON(&multiHideReq)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "error reading request body: " + err.Error()})
		return
	}
	ids := multiHideReq.Ids

	// Convert project ID strings to ObjectIDs
	projectObjectIds := make([]primitive.ObjectID, len(ids))
	for _, id := range ids {
		projectObjectId, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid project ID (in array): " + id})
			return
		}
		projectObjectIds = append(projectObjectIds, projectObjectId)
	}

	// Update the project in the database
	err = database.SetProjectsHidden(db, &projectObjectIds, multiHideReq.Hide)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error updating projects in database: " + err.Error()})
		return
	}

	// Send OK
	ctx.JSON(http.StatusOK, gin.H{"yes_no": 1})
}

// POST /project/unhide - UnhideProject unhides a project
func UnhideProject(ctx *gin.Context) {
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

	// Convert project ID string to ObjectID
	projectObjectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid project ID"})
		return
	}

	// Update the project in the database
	err = database.SetProjectHidden(db, &projectObjectId, false)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error updating project in database: " + err.Error()})
		return
	}

	// Send OK
	ctx.JSON(http.StatusOK, gin.H{"yes_no": 1})
}

// POST /project/prioritize - PrioritizeProject prioritizes a project
func PrioritizeProject(ctx *gin.Context) {
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

	// Convert project ID string to ObjectID
	projectObjectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid project ID"})
		return
	}

	// Update the project in the database
	err = database.SetProjectPrioritized(db, &projectObjectId, true)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error updating project in database: " + err.Error()})
		return
	}

	// Send OK
	ctx.JSON(http.StatusOK, gin.H{"yes_no": 1})
}

// POST /project/unprioritize - UnprioritizeProject unprioritizes a project
func UnprioritizeProject(ctx *gin.Context) {
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

	// Convert project ID string to ObjectID
	projectObjectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid project ID"})
		return
	}

	// Update the project in the database
	err = database.SetProjectPrioritized(db, &projectObjectId, false)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error updating project in database: " + err.Error()})
		return
	}

	// Send OK
	ctx.JSON(http.StatusOK, gin.H{"yes_no": 1})
}

func UpdateProjectLocation(ctx *gin.Context) {
	// Get the database from the context
	db := ctx.MustGet("db").(*mongo.Database)

	// Get ID from body
	var projLocReq models.ProjectLocationRequest

	err := ctx.BindJSON(&projLocReq)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "error reading request body: " + err.Error()})
		return
	}
	id := projLocReq.Id

	// Convert project ID string to ObjectID
	projectObjectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid project ID"})
		return
	}

	// Update the project in the database
	err = database.UpdateProjectLocationValue(db, &projectObjectId, projLocReq.Location)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error updating project location in database: " + err.Error()})
		return
	}

	// Send OK
	ctx.JSON(http.StatusOK, gin.H{"yes_no": 1})
}

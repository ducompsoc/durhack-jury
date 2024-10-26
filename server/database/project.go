package database

import (
	"context"
	"errors"
	"server/models"
	"server/util"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// UpdateProjectLastActivity to the current time
func UpdateProjectLastActivity(db *mongo.Database, ctx context.Context, id *primitive.ObjectID) error {
	// Get current time
	lastActivity := util.Now()
	_, err := db.Collection("projects").UpdateOne(ctx, gin.H{"_id": id}, gin.H{"$set": gin.H{"last_activity": lastActivity}})
	return err
}

// InsertProjects inserts a list of projects into the database
func InsertProjects(db *mongo.Database, projects []*models.Project) error {
	var docs []interface{}
	for _, project := range projects {
		docs = append(docs, project)
	}
	_, err := db.Collection("projects").InsertMany(context.Background(), docs)
	return err
}

// InsertProject inserts a project into the database
func InsertProject(db *mongo.Database, project *models.Project) error {
	_, err := db.Collection("projects").InsertOne(context.Background(), project)
	return err
}

// FindAllProjects returns a list of all projects in the database
func FindAllProjects(db *mongo.Database) ([]*models.Project, error) {
	projects := make([]*models.Project, 0)
	cursor, err := db.Collection("projects").Find(context.Background(), gin.H{})
	if err != nil {
		return nil, err
	}
	err = cursor.All(context.Background(), &projects)
	if err != nil {
		return nil, err
	}
	return projects, nil
}

// DeleteProjectById deletes a project from the database by id
func DeleteProjectById(db *mongo.Database, id primitive.ObjectID) error {
	_, err := db.Collection("projects").DeleteOne(context.Background(), gin.H{"_id": id})
	return err
}

// AggregateProjectStats aggregates all stats from the database for a project
func AggregateProjectStats(db *mongo.Database) (*models.ProjectStats, error) {
	// Get the totoal number of projects
	totalProjects, err := db.Collection("projects").EstimatedDocumentCount(context.Background())
	if err != nil {
		return nil, err
	}

	// Get the average votes and average seen using an aggregation pipeline
	cursor, err := db.Collection("projects").Aggregate(context.Background(), []gin.H{
		{"$match": gin.H{"active": true}},
		{"$group": gin.H{
			"_id": nil,
			"avgSeen": gin.H{
				"$avg": "$seen",
			},
			"numActive": gin.H{
				"$sum": 1,
			},
		}},
	})
	if err != nil {
		return nil, err
	}

	// Get the first document from the cursor
	var stats models.ProjectStats
	cursor.Next(context.Background())
	err = cursor.Decode(&stats)
	if err != nil {
		if err.Error() == "EOF" {
			stats = models.ProjectStats{Num: 0, AvgSeen: 0, NumActive: 0}
		} else {
			return nil, err
		}
	}

	// Set the total number of projects
	stats.Num = totalProjects

	return &stats, nil
}

// FindActiveProjects returns a list of all active projects in the database
func FindActiveProjects(db *mongo.Database, ctx mongo.SessionContext) ([]*models.Project, error) {
	var projects []*models.Project
	cursor, err := db.Collection("projects").Find(ctx, gin.H{"active": true})
	if err != nil {
		return nil, err
	}
	err = cursor.All(context.Background(), &projects)
	if err != nil {
		return nil, err
	}
	return projects, nil
}

// FindBusyProjects returns a list of all projects that are currently being judged.
// To do this, we collect all projects in the judge's "current" field
func FindBusyProjects(db *mongo.Database, ctx mongo.SessionContext) ([]*primitive.ObjectID, error) {
	// Get all judges that are currently judging a project
	// TODO: This query can be optimized by projecting on the "current" field
	var judges []*models.Judge
	cursor, err := db.Collection("judges").Find(ctx, gin.H{
		"current": gin.H{
			"$ne": nil,
		},
		"active": true,
	})
	if err != nil {
		return nil, err
	}
	err = cursor.All(ctx, &judges)
	if err != nil {
		return nil, err
	}

	// Extract the project IDs from the judges
	var projects []*primitive.ObjectID
	for _, judge := range judges {
		projects = append(projects, judge.Current)
	}
	return projects, nil
}

// FindProjectById returns a project from the database by id
func FindProjectById(db *mongo.Database, id *primitive.ObjectID) (*models.Project, error) {
	var project models.Project
	err := db.Collection("projects").FindOne(context.Background(), gin.H{"_id": id}).Decode(&project)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &project, nil
}

// UpdateAfterPicked updates the seen value of the new project picked and the judge's current project
func UpdateAfterPicked(db *mongo.Database, project *models.Project, judge *models.Judge) error {
	err := WithTransaction(db, func(ctx mongo.SessionContext) (interface{}, error) {
		return UpdateAfterPickedWithTx(db, project, judge, ctx)
	})
	return err
}

// UpdateAfterPickedWithTx updates the seen value of the new project picked and the judge's current project
func UpdateAfterPickedWithTx(db *mongo.Database, project *models.Project, judge *models.Judge, ctx mongo.SessionContext) (interface{}, error) {
	// Update the project's seen value and de-prioritize it
	_, err := db.Collection("projects").UpdateOne(
		ctx,
		gin.H{"_id": project.Id},
		gin.H{"$inc": gin.H{"seen": 1}, "$set": gin.H{"prioritized": false, "last_activity": util.Now()}},
	)
	if err != nil {
		return nil, err
	}
	// Set the judge's current project
	_, err = db.Collection("judges").UpdateOne(
		ctx,
		gin.H{"_id": judge.Id},
		gin.H{"$set": gin.H{"current": project.Id, "last_activity": util.Now()}},
	)
	return nil, err
}

// CountProjectDocuments returns the number of documents in the projects collection
func CountProjectDocuments(db *mongo.Database) (int64, error) {
	return db.Collection("projects").EstimatedDocumentCount(context.Background())
}

// SetProjectHidden sets the active field of a project
func SetProjectHidden(db *mongo.Database, id *primitive.ObjectID, hidden bool) error {
	_, err := db.Collection("projects").UpdateOne(context.Background(), gin.H{"_id": id}, gin.H{"$set": gin.H{"active": !hidden}})
	return err
}

// SetProjectPrioritized sets the prioritized field of a project
func SetProjectPrioritized(db *mongo.Database, id *primitive.ObjectID, prioritized bool) error {
	_, err := db.Collection("projects").UpdateOne(context.Background(), gin.H{"_id": id}, gin.H{"$set": gin.H{"prioritized": prioritized}})
	return err
}

func UpdateProjectLocationValue(db *mongo.Database, id *primitive.ObjectID, location string) error {
	_, err := db.Collection("projects").UpdateOne(context.Background(), gin.H{"_id": id}, gin.H{"$set": gin.H{"location": location}})
	return err
}

// UpdateProjects will update ALL projects in the database
func UpdateProjects(db *mongo.Database, projects []*models.Project) error {
	mongoModels := make([]mongo.WriteModel, 0, len(projects))
	for _, project := range projects {
		mongoModels = append(mongoModels, mongo.NewUpdateOneModel().SetFilter(gin.H{"_id": project.Id}).SetUpdate(gin.H{"$set": project}))
	}

	opts := options.BulkWrite().SetOrdered(false)
	_, err := db.Collection("projects").BulkWrite(context.Background(), mongoModels, opts)
	return err
}

// DecrementProjectSeenCount decrements the seen count of a project (after being skipped)
func DecrementProjectSeenCount(db *mongo.Database, ctx context.Context, project *models.Project) error {
	_, err := db.Collection("projects").UpdateOne(ctx, gin.H{"_id": project.Id}, gin.H{"$inc": gin.H{"seen": -1}})
	return err
}

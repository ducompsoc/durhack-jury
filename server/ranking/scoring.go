package ranking

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"server/database"
)

func GetScoresFromDB(db *mongo.Database) (error, string, []RankedObject) {
	// Get all the projects
	projects, err := database.FindAllProjects(db)
	if err != nil {
		return err, "error getting projects: ", nil
	}

	// Get all the judges
	judges, err := database.FindAllJudges(db)
	if err != nil {
		return err, "error getting judges: ", nil
	}

	// Create judge ranking objects
	// Create an array of {CurrentRankings: [], Unranked: []}
	judgeRankings := make([]JudgeRankings, 0)
	for _, judge := range judges {
		//for _, proj := range judge.SeenProjects {
		//	if !contains(judge.CurrentRankings, proj.ProjectId) {
		//		unranked = append(unranked, proj.ProjectId)
		//	}
		//}

		judgeRankings = append(judgeRankings, JudgeRankings{
			Rankings: judge.PastRankings,
			//Unranked: unranked,
		})
	}

	// Map all projects to their object IDs
	projectIds := make([]primitive.ObjectID, 0)
	for _, proj := range projects {
		projectIds = append(projectIds, proj.Id)
	}

	// Calculate the scores
	scores := CalcCopelandRanking(judgeRankings, projectIds)
	return nil, "", scores
}

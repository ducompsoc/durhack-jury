package models

import (
	"encoding/json"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Judge struct {
	Id             primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	KeycloakUserId string               `bson:"keycloak_user_id" json:"keycloak_user_id"`
	Active         bool                 `bson:"active" json:"active"`
	ReadWelcome    bool                 `bson:"read_welcome" json:"read_welcome"`
	Notes          string               `bson:"notes" json:"notes"`
	Current        *primitive.ObjectID  `bson:"current" json:"current"`
	Seen           int64                `bson:"seen" json:"seen"`
	SeenProjects   []JudgedProject      `bson:"seen_projects" json:"seen_projects"`
	Rankings       []primitive.ObjectID `bson:"rankings" json:"rankings"`
	LastActivity   primitive.DateTime   `bson:"last_activity" json:"last_activity"`
}

type JudgedProject struct {
	ProjectId   primitive.ObjectID `bson:"project_id" json:"project_id"`
	Categories  map[string]int     `bson:"categories" json:"categories"`
	Notes       string             `bson:"notes" json:"notes"`
	Name        string             `bson:"name" json:"name"`
	Location    int64              `bson:"location" json:"location"`
	Description string             `bson:"description" json:"description"`
}

func NewJudge(keycloakUserId string) *Judge {
	return &Judge{
		KeycloakUserId: keycloakUserId,
		Active:         true,
		ReadWelcome:    false,
		Notes:          "",
		Current:        nil,
		Seen:           0,
		SeenProjects:   []JudgedProject{},
		Rankings:       []primitive.ObjectID{},
		LastActivity:   primitive.DateTime(0),
	}
}

func JudgeProjectFromProject(project *Project, categories map[string]int) *JudgedProject {
	return &JudgedProject{
		ProjectId:   project.Id,
		Categories:  categories,
		Name:        project.Name,
		Location:    project.Location,
		Description: project.Description,
		Notes:       "",
	}
}

// Create custom marshal function to change the format of the primitive.DateTime to a unix timestamp
func (j *Judge) MarshalJSON() ([]byte, error) {
	type Alias Judge
	return json.Marshal(&struct {
		*Alias
		LastActivity int64 `json:"last_activity"`
	}{
		Alias:        (*Alias)(j),
		LastActivity: int64(j.LastActivity),
	})
}

// Create custom unmarshal function to change the format of the primitive.DateTime from a unix timestamp
func (j *Judge) UnmarshalJSON(data []byte) error {
	type Alias Judge
	aux := &struct {
		LastActivity int64 `json:"last_activity"`
		*Alias
	}{
		Alias: (*Alias)(j),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	j.LastActivity = primitive.DateTime(aux.LastActivity)
	return nil
}

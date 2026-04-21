package models

import (
	"encoding/json"
	"server/util"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type HiddenReason struct {
	Reason        string             `bson:"reason" json:"reason"`
	When          primitive.DateTime `bson:"when" json:"when"`
}

type Project struct {
	Id            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name          string             `bson:"name" json:"name"`
	Guild         string             `bson:"guild" json:"guild"`
	Location      string             `bson:"location" json:"location"`
	Description   string             `bson:"description" json:"description"`
	Url           string             `bson:"url" json:"url"`
	TryLink       string             `bson:"try_link" json:"try_link"`
	VideoLink     string             `bson:"video_link" json:"video_link"`
	ChallengeList []string           `bson:"challenge_list" json:"challenge_list"`
	Seen          int64              `bson:"seen" json:"seen"`
	HiddenReasons []HiddenReason     `bson:"hidden_reasons" json:"hidden_reasons"`
	LastActivity  primitive.DateTime `bson:"last_activity" json:"last_activity"`
}

func (p *Project) GetLocationString() string {
	if p.Guild == "" {
		return p.Location
	} else {
		return p.Guild + "|" + p.Location
	}
}

func NewProject(name string, guild string, location string, description string, url string, tryLink string, videoLink string, challengeList []string) *Project {
	return &Project{
		Name:          name,
		Guild:         guild,
		Location:      location,
		Description:   description,
		Url:           url,
		TryLink:       tryLink,
		VideoLink:     videoLink,
		ChallengeList: challengeList,
		Seen:          0,
		HiddenReasons: []HiddenReason{},
		LastActivity:  primitive.DateTime(0),
	}
}

func NewHiddenReason(reason string) *HiddenReason {
	return &HiddenReason{
		Reason: reason,
		When: util.Now(),
	}
}

// Create custom marshal function to change the format of the primitive.DateTime to a unix timestamp
func (p *Project) MarshalJSON() ([]byte, error) {
	type Alias Project
	return json.Marshal(&struct {
		*Alias
		LastActivity int64 `json:"last_activity"`
	}{
		Alias:        (*Alias)(p),
		LastActivity: int64(p.LastActivity),
	})
}

// Create custom unmarshal function to change the format of the primitive.DateTime from a unix timestamp
func (p *Project) UnmarshalJSON(data []byte) error {
	type Alias Project
	aux := &struct {
		LastActivity int64 `json:"last_activity"`
		*Alias
	}{
		Alias: (*Alias)(p),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	p.LastActivity = primitive.DateTime(aux.LastActivity)
	return nil
}

// Custom marhal function to change the format of primative.DateTime to a unix timestamp
func (h *HiddenReason) MarshalJSON() ([]byte, error) {
	type Alias HiddenReason
	return json.Marshal(&struct {
		When         int64 `json:"when"`
		*Alias
	}{
		Alias:       (*Alias)(h),
		When:        int64(h.When),
	})
}

func (h *HiddenReason) UnmarshalJSON(data []byte) error {
	type Alias HiddenReason
	aux := &struct {
		When         int64 `json:"when"`
		*Alias
	}{
		Alias:       (*Alias)(h),
	}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}
	h.When = primitive.DateTime(aux.When)
	return nil
}

// Custom unmarhal function to change the format of primative.DateTime from a unix timestamp

// Type to sort by table number
type ByTableNumber []*Project

func (a ByTableNumber) Len() int           { return len(a) }
func (a ByTableNumber) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByTableNumber) Less(i, j int) bool { return a[i].Location < a[j].Location }

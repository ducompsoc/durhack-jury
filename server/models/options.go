package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Options struct {
	Id               primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Ref              int64              `bson:"ref" json:"ref"`
	CurrTableNum     int64              `bson:"curr_table_num" json:"curr_table_num"`
	Clock            ClockState         `bson:"clock" json:"clock"`
	JudgingTimer     int64              `bson:"judging_timer" json:"judging_timer"`
	MinViews         int64              `bson:"min_views" json:"min_views"`
	Categories       []string           `bson:"categories" json:"categories"`
	BatchRankingSize int64              `bson:"batch_ranking_size" json:"batch_ranking_size"`
	JudgingEnded     bool               `bson:"judging_ended" json:"judging_ended"`
}

func NewOptions() *Options {
	return &Options{
		Ref:              0,
		CurrTableNum:     0,
		JudgingTimer:     300,
		MinViews:         3,
		Clock:            *NewClockState(),
		Categories:       []string{"Creativity/Innovation", "Technical Competence/Execution", "Research/Design", "Presentation"},
		BatchRankingSize: 8,
		JudgingEnded:     false,
	}
}

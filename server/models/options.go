package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Options struct {
	Id               primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Ref              int64              `bson:"ref" json:"ref"`
	CurrTableNum     int64              `bson:"curr_table_num" json:"curr_table_num"`
	Clock            ClockState         `bson:"clock" json:"clock"`
	JudgingTimer     int64              `bson:"judging_timer" json:"judging_timer"`
	Categories       []string           `bson:"categories" json:"categories"`
	RankingBatchSize int64              `bson:"ranking_batch_size" json:"ranking_batch_size"`
}

func NewOptions() *Options {
	return &Options{
		Ref:              0,
		CurrTableNum:     0,
		JudgingTimer:     300,
		Clock:            *NewClockState(),
		Categories:       []string{"Creativity/Innovation", "Technical Competence/Execution", "Research/Design", "Presentation"},
		RankingBatchSize: 8,
	}
}

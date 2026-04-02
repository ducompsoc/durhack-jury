package models

type JudgeStats struct {
	Num       int64   `json:"num"`
	AvgSeen   float64 `json:"avg_seen"`
	NumActive int64   `json:"num_active"`
}

type ProjectStats struct {
	Num       int64   `json:"num"`
	AvgSeen   float64 `json:"avg_seen"`
	NumActive int64   `json:"num_active"`
}

type Stats struct {
	Projects       int64   `json:"num_projects"`
	HiddenProjects int64   `json:"hidden_projects"`
	Judges         int64   `json:"num_judges"`
	AvgProjectSeen float64 `json:"avg_project_seen"`
	AvgJudgeSeen   float64 `json:"avg_judge_seen"`
}

type JudgeVote struct {
	CurrWinner bool `json:"curr_winner"`
}

type IdRequest struct {
	Id string `json:"id"`
}

type MultiIdReq struct {
	Ids  []string `json:ids`
}

type IdHideReq struct {
	Id     string `json:"id"`
	Reason string `json:"reason"`
}

type MultiIdHideReq struct {
	Ids  []string `json:"ids"`
	Reason string `json:"reason"`
}

type ProjectLocationRequest struct {
	IdRequest
	Location string `json:"location"`
}

type EditJudgeRequest struct {
	Notes string `json:"notes"`
}

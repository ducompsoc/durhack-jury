package models

import "time"

type ClockState struct {
	// The duration (in milliseconds) elapsed while the clock was running, not including
	// time elapsed since the most recent resume.
	ElapsedDuration int64 `json:"elapsed_duration" bson:"elapsed_duration"`
	// The timestamp (milliseconds since UNIX epoch) of the most recent clock resume.
	// If `Running` is `false`, this value has no meaning and should be zero.
	ResumedTime int64 `json:"resumed_time" bson:"resumed_time"`
	// Whether the clock is presently running (a.k.a. resumed).
	Running bool `json:"running" bson:"running"`
}

func NewClockState() *ClockState {
	return &ClockState{
		ElapsedDuration: 0,
		ResumedTime:     0,
		Running:         false,
	}
}

// Gets the current clock time in milliseconds
func GetCurrTime() int64 {
	return time.Now().UnixNano() / 1000000
}

func (c *ClockState) Pause() {
	if !c.Running {
		return
	}
	c.ElapsedDuration += GetCurrTime() - c.ResumedTime
	c.Running = false
	c.ResumedTime = 0
}

func (c *ClockState) Resume() {
	if c.Running {
		return
	}
	c.Running = true
	c.ResumedTime = GetCurrTime()
}

func (c *ClockState) Reset() {
	c.ElapsedDuration = 0
	c.ResumedTime = 0
	c.Running = false
}

func (c *ClockState) GetDuration() int64 {
	if !c.Running {
		return c.ElapsedDuration
	}
	return c.ElapsedDuration + (GetCurrTime() - c.ResumedTime)
}

package funcs

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"strings"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"server/models"
	"server/ranking"
	"server/util"

	"github.com/gin-gonic/gin"
)

// Read CSV file and return a slice of project structs
func ParseProjectCsv(content string, hasHeader bool) ([]*models.Project, error) {
	r := csv.NewReader(strings.NewReader(content))

	// Empty CSV file
	if content == "" {
		return []*models.Project{}, nil
	}

	// If the CSV file has a header, skip the first line
	if hasHeader {
		r.Read()
	}

	// Read the CSV file, looping through each record
	var projects []*models.Project
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		// Make sure the record has at least 4 elements (name, location, description, URL)
		if len(record) < 4 {
			return nil, fmt.Errorf("record contains less than 4 elements: '%s'", strings.Join(record, ","))
		}

		// Get the challenge list
		var challengeList []string
		if len(record) > 5 && record[6] != "" {
			challengeList = strings.Split(record[6], ",")
		}
		for i := range challengeList {
			challengeList[i] = strings.TrimSpace(challengeList[i])
		}

		// Optional fields
		var tryLink string
		if len(record) > 3 && record[4] != "" {
			tryLink = record[3]
		}
		var videoLink string
		if len(record) > 4 && record[5] != "" {
			videoLink = record[4]
		}

		// Add project to slice
		projects = append(projects, models.NewProject(record[0], "", record[1], record[2], record[3], tryLink, videoLink, challengeList))
	}

	return projects, nil
}

// TODO: After event, devpost will add a column between 0 and 1, the "auto assigned table numbers" TT - idk what to do abt this
// Generate a workable CSV for Jury based on the output CSV from Devpost
// Columns:
//  0. Project Title - title
//  1. Submission Url - url
//  2. Project Status - Draft or Submitted (ignore drafts)
//  3. Judging Status - ignore
//  4. Highest Step Completed - ignore
//  5. Project Created At - ignore
//  6. About The Project - description
//  7. "Try it out" Links" - try_link
//  8. Video Demo Link - video_link
//  9. Opt-In Prizes - challenge_list
//  10. Built With - ignore
//  11. Notes - ignore
//  12. Team Colleges/Universities - ignore
//  13. Additional Team Member Count - ignore
//  14. !Megateam/Guild - guild  [column O in excel]
//  15. !Table number - location [column P in excel]
//
// (16+.and remiaining columns) Custom questions - custom_questions (ignore for now)
func ParseDevpostCSV(content string) ([]*models.Project, error) {
	r := csv.NewReader(strings.NewReader(content))

	// Empty CSV file
	if content == "" {
		return []*models.Project{}, nil
	}

	// Skip the first line
	_, err := r.Read()
	if err != nil {
		return nil, err
	}

	// Read the CSV file, looping through each record
	var projects []*models.Project
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		// Make sure the record has 14 or more elements (see above)
		if len(record) < 14 {
			return nil, fmt.Errorf("record does not contain 15 or more elements (invalid devpost csv): '%s'", strings.Join(record, ","))
		}

		// If the project is a Draft, skip it
		if record[2] == "Draft" {
			continue
		}

		// Split challenge list into a slice and trim them
		challengeList := strings.Split(record[9], ",")
		if record[9] == "" {
			challengeList = []string{}
		}
		for i := range challengeList {
			challengeList[i] = strings.TrimSpace(challengeList[i])
		}

		// Add project to slice
		projects = append(projects, models.NewProject(
			record[0],
			record[14],
			record[15],
			record[6],
			record[1],
			record[7],
			record[8],
			challengeList,
		))
	}

	return projects, nil
}

// Adapted from https://stackoverflow.com/a/74700627/7253717
func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[0:maxLen])
}

// AddCSVData adds a CSV file to the response
func AddCsvData(name string, content []byte, ctx *gin.Context) {
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s.csv", name))
	ctx.Header("Content-Type", "text/csv")
	ctx.Data(http.StatusOK, "text/csv", content)
}

// AddZipFile adds a zip file to the response
func AddZipFile(name string, content []byte, ctx *gin.Context) {
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s.zip", name))
	ctx.Header("Content-Type", "application/octet-stream")
	ctx.Data(http.StatusOK, "application/octet-stream", content)
}

// Create a CSV file from the judges but only the rankings  lucatodo: update for new ranking object structures (2D array)
func CreateJudgeRankingCSV(judges []*models.Judge) []byte {
	csvBuffer := &bytes.Buffer{}

	// Create a new CSV writer
	w := csv.NewWriter(csvBuffer)

	// Write the header
	// lucatodo: remove unranked concept
	w.Write([]string{"Name", "Code", "Ranked", "Unranked"})

	// Write each judge
	for _, judge := range judges {
		// Don't include if their rankings are empty :/
		if len(judge.CurrentRankings) == 0 {
			continue
		}

		// Create a list of all ranked projects (just their location)
		ranked := make([]string, 0, len(judge.CurrentRankings))
		for _, projId := range judge.CurrentRankings {
			idx := util.IndexFunc(judge.SeenProjects, func(p models.JudgedProject) bool {
				return p.ProjectId == projId
			})
			if idx == -1 {
				continue
			}
			proj := judge.SeenProjects[idx]
			ranked = append(ranked, proj.GetLocationString())
		}

		// Create a list of all unranked projects (filter using ranked projects)
		unranked := make([]string, 0, len(judge.SeenProjects)-len(judge.CurrentRankings))
		for _, proj := range judge.SeenProjects {
			if util.ContainsFunc(ranked, func(location string) bool { return location == proj.Location }) {
				unranked = append(unranked, proj.Location)
			}
		}

		// Write line to CSV
		// lucatodo: get judge info (from gocloak using admin api) to write to csv
		w.Write([]string{judge.KeycloakUserId, strings.Join(ranked, ","), strings.Join(unranked, ",")})
	}

	// Flush the writer
	w.Flush()

	return csvBuffer.Bytes()
}

// Create a CSV file from a list of projects
func CreateProjectCSV(projects []*models.Project, scores []ranking.RankedObject) []byte {
	csvBuffer := &bytes.Buffer{}

	// Create a new CSV writer
	w := csv.NewWriter(csvBuffer)

	// Write the header
	w.Write([]string{"Name", "Location", "Description", "URL", "TryLink", "VideoLink", "ChallengeList", "TimesSeen", "Active", "LastActivity", "Score"})

	// Create a map to store scores by project ID
	scoreMap := make(map[primitive.ObjectID]float64)
	for _, ps := range scores {
		scoreMap[ps.Id] = ps.Score
	}

	// Write each project
	for _, project := range projects {
		w.Write([]string{project.Name, project.Location, project.Description, project.Url, project.TryLink,
			project.VideoLink, strings.Join(project.ChallengeList, ","), fmt.Sprintf("%d", project.Seen),
			fmt.Sprintf("%t", project.Active), fmt.Sprintf("%d", project.LastActivity),
			fmt.Sprintf("%.1f", scoreMap[project.Id])})
	}

	// Flush the writer
	w.Flush()

	return csvBuffer.Bytes()
}

// CreateProjectChallengeZip creates a zip file with a CSV for each challenge
func CreateProjectChallengeZip(projects []*models.Project, scores []ranking.RankedObject) ([]byte, error) {
	var csvList [][]byte

	// Get list of challenges
	var challengeList []string
	for _, project := range projects {
		for _, challenge := range project.ChallengeList {
			if !contains(challengeList, challenge) {
				challengeList = append(challengeList, challenge)
			}
		}
	}

	// Create a CSV for each challenge
	for _, challenge := range challengeList {
		var currChallengeProjects []*models.Project
		for _, project := range projects {
			if contains(project.ChallengeList, challenge) {
				currChallengeProjects = append(currChallengeProjects, project)
			}
		}

		// Create CSV for the challenge
		challengeCSV := CreateProjectCSV(currChallengeProjects, scores)
		csvList = append(csvList, challengeCSV)
	}

	// Create buffer for zip file
	zipBuffer := &bytes.Buffer{}

	// Create a new zip writer
	w := zip.NewWriter(zipBuffer)

	// Write each CSV to the zip file
	for i, challengeCSV := range csvList {
		f, err := w.Create(fmt.Sprintf("%s.csv", challengeList[i]))
		if err != nil {
			return nil, err
		}

		_, err = f.Write(challengeCSV)
		if err != nil {
			return nil, err
		}
	}

	// Close the zip writer
	err := w.Close()
	if err != nil {
		return nil, err
	}

	return zipBuffer.Bytes(), nil
}

// contains checks if a string is in a list of strings
func contains(list []string, str string) bool {
	for _, s := range list {
		if s == str {
			return true
		}
	}
	return false
}

package main

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/Songmu/smartcache"
)

// Candidate Model
type Candidate struct {
	ID             int
	Name           string
	PoliticalParty string
	Sex            string
}

// CandidateElectionResult type
type CandidateElectionResult struct {
	ID             int
	Name           string
	PoliticalParty string
	Sex            string
	VoteCount      int
}

// PartyElectionResult type
type PartyElectionResult struct {
	PoliticalParty string
	VoteCount      int
}

var _candidates = []Candidate{}
var _candidateByID = map[int]Candidate{}
var _candidateByName = map[string]Candidate{}
var _candidatesByParty = map[string][]Candidate{}

func initAllCandidate() {
	_candidates = []Candidate{}
	_candidateByID = make(map[int]Candidate)
	_candidateByName = make(map[string]Candidate)
	_candidatesByParty = make(map[string][]Candidate)
	rows, err := db.Query("SELECT * FROM candidates")
	if err != nil {
		panic(err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		c := Candidate{}
		err = rows.Scan(&c.ID, &c.Name, &c.PoliticalParty, &c.Sex)
		if err != nil {
			panic(err.Error())
		}
		_candidateByID[c.ID] = c
		_candidateByName[c.Name] = c
		if _, ok := _candidatesByParty[c.PoliticalParty]; !ok {
			_candidatesByParty[c.PoliticalParty] = []Candidate{c}
		} else {
			_candidatesByParty[c.PoliticalParty] = append(_candidatesByParty[c.PoliticalParty], c)
		}
		_candidates = append(_candidates, c)
	}
}

func getAllCandidate() []Candidate {
	return _candidates
}

func getCandidate(candidateID int) (Candidate, error) {
	c, ok := _candidateByID[candidateID]
	if !ok {
		return Candidate{}, errors.New("not found")
	}
	return c, nil
}

func getCandidateByName(name string) (c Candidate, err error) {
	c, ok := _candidateByName[name]
	if !ok {
		return Candidate{}, errors.New("not found")
	}
	return c, nil
}

func getAllPartyName() (partyNames []string) {
	rows, err := db.Query("SELECT political_party FROM candidates GROUP BY political_party")
	if err != nil {
		panic(err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		err = rows.Scan(&name)
		if err != nil {
			panic(err.Error())
		}
		partyNames = append(partyNames, name)
	}
	return
}

func getCandidatesByPoliticalParty(party string) []Candidate {
	return _candidatesByParty[party]
}

var getElectionResultCache = smartcache.New(1*time.Minute, 30*time.Second, func(ctx context.Context) (interface{}, error) {
	val := _getElectionResult()
	return val, nil
})

func getElectionResult() (result []CandidateElectionResult) {
	val, _ := getElectionResultCache.Get(context.Background())

	return val.([]CandidateElectionResult)
}

func _getElectionResult() (result []CandidateElectionResult) {

	ctx := context.TODO()
	res := rdb.ZRevRangeWithScores(ctx, candidateVoteRedisKey, 0, -1).Val()

	// candidates := getAllCandidate()
	// まず1票以上投票がある人を追加していく
	scoreByCandidate := make(map[int]int)
	for _, m := range res {
		candidateID, _ := strconv.Atoi(m.Member.(string))
		scoreByCandidate[candidateID] = int(m.Score)

		candidate, _ := getCandidate(candidateID)
		r := CandidateElectionResult{
			ID:             candidate.ID,
			Name:           candidate.Name,
			PoliticalParty: candidate.PoliticalParty,
			Sex:            candidate.Sex,
			VoteCount:      int(m.Score),
		}
		result = append(result, r)
	}

	// 0票の人がいたらそれを追加する
	//	if len(result) != len(res) {
	candidates := getAllCandidate()
	for _, candidate := range candidates {
		if scoreByCandidate[candidate.ID] == 0 {
			r := CandidateElectionResult{
				ID:             candidate.ID,
				Name:           candidate.Name,
				PoliticalParty: candidate.PoliticalParty,
				Sex:            candidate.Sex,
				VoteCount:      0,
			}
			result = append(result, r)

		}
	}
	//}

	return
}

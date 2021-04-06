package main

import (
	"context"
	"errors"
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

var (
	expire     = 1 * time.Minute
	softExpire = 30 * time.Second
)
var ca = smartcache.New(expire, softExpire, func(ctx context.Context) (interface{}, error) {
	val := _getElectionResult()
	return val, nil
})

func getElectionResult() (result []CandidateElectionResult) {
	val, _ := ca.Get(context.Background())

	return val.([]CandidateElectionResult)
}

func _getElectionResult() (result []CandidateElectionResult) {

	rows, err := db.Query(`
		SELECT c.id, c.name, c.political_party, c.sex, IFNULL(v.count, 0)
		FROM candidates AS c
		LEFT OUTER JOIN
	  	(SELECT candidate_id, COUNT(*) AS count
	  	FROM votes
	  	GROUP BY candidate_id) AS v
		ON c.id = v.candidate_id
		ORDER BY v.count DESC`)
	if err != nil {
		panic(err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		r := CandidateElectionResult{}
		err = rows.Scan(&r.ID, &r.Name, &r.PoliticalParty, &r.Sex, &r.VoteCount)
		if err != nil {
			panic(err.Error())
		}
		result = append(result, r)
	}
	return
}

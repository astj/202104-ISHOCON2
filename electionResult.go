package main

import "sort"

type topRenderArgs struct {
	candidates []CandidateElectionResult
	parties    []PartyElectionResult
	sexRatio   map[string]int
}

func getTopRenderArgs() topRenderArgs {
	electionResults := getElectionResult()

	// 上位10人と最下位のみ表示
	tmp := make([]CandidateElectionResult, len(electionResults))
	copy(tmp, electionResults)
	candidates := tmp[:10]
	candidates = append(candidates, tmp[len(tmp)-1])

	partyNames := getAllPartyName()
	partyResultMap := map[string]int{}
	for _, name := range partyNames {
		partyResultMap[name] = 0
	}
	for _, r := range electionResults {
		partyResultMap[r.PoliticalParty] += r.VoteCount
	}
	partyResults := []PartyElectionResult{}
	for name, count := range partyResultMap {
		r := PartyElectionResult{}
		r.PoliticalParty = name
		r.VoteCount = count
		partyResults = append(partyResults, r)
	}
	// 投票数でソート
	sort.Slice(partyResults, func(i, j int) bool { return partyResults[i].VoteCount > partyResults[j].VoteCount })

	sexRatio := map[string]int{
		"men":   0,
		"women": 0,
	}
	for _, r := range electionResults {
		if r.Sex == "男" {
			sexRatio["men"] += r.VoteCount
		} else if r.Sex == "女" {
			sexRatio["women"] += r.VoteCount
		}
	}

	return topRenderArgs{
		candidates: candidates,
		parties:    partyResults,
		sexRatio:   sexRatio,
	}
}

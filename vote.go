package main

import (
	"context"
	"fmt"
)

// Vote Model
type Vote struct {
	ID          int
	UserID      int
	CandidateID int
	Keyword     string
}

var candidateVoteRedisKey = "vote-c"

func voteRedisMemberForCandidate(candidateID int) string {
	return fmt.Sprintf("%d", candidateID)
}

func voteRedisKeyForUser(userID int) string {
	return fmt.Sprintf("vote-u-%d", userID)
}

func getVoteCountByCandidateID(candidateID int) (count int) {
	ctx := context.TODO()
	c := rdb.ZScore(ctx, candidateVoteRedisKey, voteRedisMemberForCandidate(candidateID)).Val()
	return int(c)
}

func getUserVotedCount(userID int) (count int) {
	count, _ = rdb.Get(context.TODO(), voteRedisKeyForUser(userID)).Int()
	return
}

func createVote(userID int, candidateID int, keyword string) {
	ctx := context.TODO()
	rdb.ZIncrBy(ctx, candidateVoteRedisKey, 1, voteRedisMemberForCandidate(candidateID))
	rdb.IncrBy(ctx, voteRedisKeyForUser(userID), 1)
}

func createVoteBy(userID int, candidateID int, count int64) {
	ctx := context.TODO()
	rdb.ZIncrBy(ctx, candidateVoteRedisKey, float64(count), voteRedisMemberForCandidate(candidateID))
	rdb.IncrBy(ctx, voteRedisKeyForUser(userID), count)
}

func getVoiceOfSupporter(candidateIDs []int) (voices []string) {
	return []string{}
}

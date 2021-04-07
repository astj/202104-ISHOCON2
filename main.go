package main

import (
	"context"
	"database/sql"
	"html/template"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB
var rdb *redis.Client

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func main() {
	// database setting
	user := getEnv("ISHOCON2_DB_USER", "ishocon")
	pass := getEnv("ISHOCON2_DB_PASSWORD", "ishocon")
	dbname := getEnv("ISHOCON2_DB_NAME", "ishocon2")
	// mysql は static なデータしか乗ってないのでそれぞれのホストを見る
	db, _ = sql.Open("mysql", user+":"+pass+"@/"+dbname)
	db.SetMaxIdleConns(5)
	redisAddr := getEnv("ISHOCON2_REDIS_ADDR", "localhost:6379")
	rdb = redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	// initialize data
	// /initialize で呼んでもいいかも
	initAllCandidate()
	initUsers()
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		panic(err)
	}

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(static.Serve("/css", static.LocalFile("public/css", true)))
	layout := "templates/layout.tmpl"

	// session store
	store := sessions.NewCookieStore([]byte("mysession"))
	store.Options(sessions.Options{HttpOnly: true})
	r.Use(sessions.Sessions("showwin_happy", store))

	// GET /
	r.GET("/", func(c *gin.Context) {
		args := getTopRenderArgs()

		funcs := template.FuncMap{"indexPlus1": func(i int) int { return i + 1 }}
		r.SetHTMLTemplate(template.Must(template.New("main").Funcs(funcs).ParseFiles(layout, "templates/index.tmpl")))
		c.Header("Cache-Control", "public; max-age=15")
		c.HTML(http.StatusOK, "base", gin.H{
			"candidateResults": args.candidates,
			"parties":          args.parties,
			"sexRatio":         args.sexRatio,
		})
	})

	// GET /candidates/:candidateID(int)
	r.GET("/candidates/:candidateID", func(c *gin.Context) {
		candidateID, _ := strconv.Atoi(c.Param("candidateID"))
		candidate, err := getCandidate(candidateID)
		if err != nil {
			c.Redirect(http.StatusFound, "/")
		}
		votes := getVoteCountByCandidateID(candidateID)
		candidateIDs := []int{candidateID}
		keywords := getVoiceOfSupporter(candidateIDs)

		r.SetHTMLTemplate(template.Must(template.ParseFiles(layout, "templates/candidate.tmpl")))
		c.Header("Cache-Control", "public; max-age=15")
		c.HTML(http.StatusOK, "base", gin.H{
			"candidate": candidate,
			"votes":     votes,
			"keywords":  keywords,
		})
	})

	// GET /political_parties/:name(string)
	r.GET("/political_parties/:name", func(c *gin.Context) {
		partyName := c.Param("name")
		var votes int
		electionResults := getElectionResult()
		for _, r := range electionResults {
			if r.PoliticalParty == partyName {
				votes += r.VoteCount
			}
		}

		candidates := getCandidatesByPoliticalParty(partyName)
		candidateIDs := []int{}
		for _, c := range candidates {
			candidateIDs = append(candidateIDs, c.ID)
		}
		keywords := getVoiceOfSupporter(candidateIDs)

		r.SetHTMLTemplate(template.Must(template.ParseFiles(layout, "templates/political_party.tmpl")))
		c.Header("Cache-Control", "public; max-age=15")
		c.HTML(http.StatusOK, "base", gin.H{
			"politicalParty": partyName,
			"votes":          votes,
			"candidates":     candidates,
			"keywords":       keywords,
		})
	})

	// GET /vote
	r.GET("/vote", func(c *gin.Context) {
		candidates := getAllCandidate()
		r.SetHTMLTemplate(template.Must(template.ParseFiles(layout, "templates/vote.tmpl")))
		c.Header("Cache-Control", "public; max-age=64800")
		c.HTML(http.StatusOK, "base", gin.H{
			"candidates": candidates,
			"message":    "",
		})
	})

	// POST /vote
	r.POST("/vote", func(c *gin.Context) {
		candidates := getAllCandidate()
		user, userErr := getUser(c.PostForm("name"), c.PostForm("address"), c.PostForm("mynumber"))
		candidate, cndErr := getCandidateByName(c.PostForm("candidate"))
		votedCount := getUserVotedCount(user.ID)
		voteCount, _ := strconv.Atoi(c.PostForm("vote_count"))

		var message string
		r.SetHTMLTemplate(template.Must(template.ParseFiles(layout, "templates/vote.tmpl")))
		if userErr != nil {
			message = "個人情報に誤りがあります"
		} else if user.Votes < voteCount+votedCount {
			message = "投票数が上限を超えています"
		} else if c.PostForm("candidate") == "" {
			message = "候補者を記入してください"
		} else if cndErr != nil {
			message = "候補者を正しく記入してください"
		} else if c.PostForm("keyword") == "" {
			message = "投票理由を記入してください"
		} else {
			for i := 1; i <= voteCount; i++ {
				createVote(user.ID, candidate.ID, c.PostForm("keyword"))
			}
			message = "投票に成功しました"
		}
		c.HTML(http.StatusOK, "base", gin.H{
			"candidates": candidates,
			"message":    message,
		})
	})

	r.GET("/initialize", func(c *gin.Context) {
		rdb.FlushAll(c)

		c.String(http.StatusOK, "Finish")
	})

	r.Run(":8080")
}

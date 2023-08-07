package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

var dinnerSchema = `CREATE TABLE IF NOT EXISTS dinner (
	id   VARCHAR(36) PRIMARY KEY,
  name VARCHAR(80),
	date DATETIME,
	team_size   INTEGER,
	teams_per_course INTEGER
);`

var courseSchema = `CREATE TABLE IF NOT EXISTS course (
	id   VARCHAR(36) PRIMARY KEY,
  name VARCHAR(80),
	position INTEGER,
	team_size   INTEGER
);`

var teamSchema = `CREATE TABLE IF NOT EXISTS team (
	id   VARCHAR(36) PRIMARY KEY,
	dinner_id VARCHAR(36)
);`

var teamMemberSchema = `CREATE TABLE IF NOT EXISTS team (
	id   VARCHAR(36) PRIMARY KEY,
	name VARCHAR(50),
	team_id VARCHAR(36)
);`

type Dinner struct {
	Id             string    `db:"id" json:"id"`
	Name           string    `db:"name" json:"name"`
	Date           time.Time `db:"date" json:"date"`
	TeamSize       int       `db:"team_size" json:"team_size"`
	TeamsPerCourse int       `db:"teams_per_course" json:"teams_per_course"`
}

type Course struct {
	Id       string `db:"id" json:"id"`
	Name     string `db:"name" json:"name"`
	Position int    `db:"position" json:"position"`
}

type Team struct {
	Id       string `db:"id" json:"id"`
	DinnerId string `db:"dinner_id" json:"dinner_id"`
}

type TeamMember struct {
	Id     string `db:"id" json:"id"`
	Name   string `db:"name" json:"name"`
	TeamId string `db:"team_id" json:"team_id"`
}

func main() {
	initDb()
	g := gin.Default()
	g.POST("/dinner", createDinner)
	g.GET("/dinner/:uuid", getDinner)

	g.Run()
	log.Println("Started Gin server")
}

func initDb() {
	db := getDb()
	defer db.Close()

	db.MustExec(dinnerSchema)
	db.MustExec(courseSchema)
	db.MustExec(teamSchema)
	db.MustExec(teamMemberSchema)
}

func getDb() *sqlx.DB {
	db, err := sqlx.Connect("sqlite3", "dinner.db")
	if err != nil {
		log.Fatalln(err)
	}
	return db
}

func createDinner(c *gin.Context) {
	var dinner Dinner
	dinner.Id = uuid.New().String()
	dinner.Date = time.Now()

	if err := c.BindJSON(&dinner); err != nil {
		log.Println(err)
		c.Status(http.StatusBadRequest)
		return
	}

	db := getDb()
	defer db.Close()

	log.Println(dinner)
	_, err := db.NamedExec("INSERT INTO dinner (id, name, date, team_size, teams_per_course) VALUES (:id, :name, :date, :team_size, :teams_per_course)", dinner)
	if err != nil {
		log.Println(err)
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, dinner.Id)
}

func getDinner(c *gin.Context) {
	uuid := c.Param("uuid")
	db := getDb()
	defer db.Close()

	dinners := []Dinner{}
	err := db.Select(&dinners, `SELECT * FROM dinner WHERE id = $1`, uuid)
	if err != nil {
		log.Println(err)
		c.Status(http.StatusInternalServerError)
		return
	}
	if len(dinners) == 0 {
		c.Status(http.StatusNotFound)
		return
	}
	c.JSON(http.StatusOK, dinners[0])
}

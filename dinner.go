package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Dinner struct {
	DinnerId       string    `json:"dinnerId" gorm:"primaryKey"`
	SignupId       string    `json:"signupId"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	Date           time.Time `json:"date"`
	TeamSize       int       `json:"teamSize"`
	TeamsPerCourse int       `json:"teamsPerCourse"`
	Courses        []Course  `json:"courses"`
	Teams          []Team    `json:"teams"`
}

type Course struct {
	CourseId      string        `json:"courseId" uri:"courseId" gorm:"primaryKey"`
	DinnerId      string        `json:"-" uri:"dinnerId"`
	Name          string        `json:"name"`
	Time          time.Time     `json:"time"`
	Duration      int           `json:"duration"`
	CourseMatches []CourseMatch `json:"courseMatches"`
}

type Team struct {
	TeamId        string        `json:"teamId" uri:"teamId" gorm:"primaryKey"`
	SignupId      string        `json:"signupId"`
	DinnerId      string        `json:"-"`
	TeamMembers   []string      `json:"teamMembers" gorm:"serializer:json"`
	CourseMatches []CourseMatch `json:"courseMatches" gorm:"many2many:team_course_matches"`
	Address       string        `json:"address"`
}

type CourseMatch struct {
	CourseMatchId string `json:"-" gorm:"primaryKey"`
	CourseId      string `json:"course"`
	Teams         []Team `json:"teams" gorm:"many2many:team_course_matches"`
	Host          Team   `json:"host" gorm:"foreignKey:TeamId"`
}

var db *gorm.DB

func main() {
	var err error
	db, err = gorm.Open(sqlite.Open("dinner.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	db.AutoMigrate(&Dinner{})
	db.AutoMigrate(&Course{})
	db.AutoMigrate(&Team{})
	db.AutoMigrate(&CourseMatch{})

	r := gin.Default()

	r.POST("/dinners", createDinner)
	r.GET("/dinners/:dinnerId", getDinner)
	r.PUT("/dinners", updateDinner)
	r.DELETE("/dinners/:dinnerId", deleteDinner)

	r.POST("/teams", createTeam)
	r.GET("/teams/:teamId", getTeam)
	r.PUT("/teams", updateTeam)
	r.DELETE("/teams/:teamId", deleteTeam)

	r.Run()
}

func createDinner(c *gin.Context) {
	var dinner Dinner
	dinner.DinnerId = uuid.New().String()
	dinner.SignupId = uuid.New().String()
	dinner.Date = time.Now()

	if err := c.BindJSON(&dinner); err != nil {
		log.Println(err)
		c.Status(http.StatusBadRequest)
		return
	}

	if err := db.Create(dinner).Error; err != nil {
		log.Println(err)
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusCreated, dinner)
}

func getDinner(c *gin.Context) {
	dinnerId := c.Param("dinnerId")

	var dinner Dinner
	if err := db.Preload("Teams").Preload("Courses").First(&dinner, &Dinner{DinnerId: dinnerId}).Error; err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	c.JSON(http.StatusOK, dinner)
}

func updateDinner(c *gin.Context) {
	var dinner Dinner

	if err := c.BindJSON(&dinner); err != nil {
		log.Println(err)
		c.Status(http.StatusBadRequest)
		return
	}

	if err := db.Updates(&dinner).Error; err != nil {
		log.Println(err)
		c.Status(http.StatusNotFound)
		return
	}

	c.JSON(http.StatusOK, dinner)
}

func deleteDinner(c *gin.Context) {
	dinnerId := c.Param("dinnerId")

	if err := db.Delete(&Dinner{}, &Dinner{DinnerId: dinnerId}).Error; err != nil {
		log.Println(err)
		c.Status(http.StatusNotFound)
		return
	}

	c.Status(http.StatusOK)
}

func createTeam(c *gin.Context) {
	var team Team
	team.TeamId = uuid.New().String()
	team.CourseMatches = []CourseMatch{}

	if err := c.BindJSON(&team); err != nil {
		log.Println(err)
		c.Status(http.StatusBadRequest)
		return
	}

	var dinner Dinner
	if err := db.Preload("Teams").First(&dinner, "signup_id = ?", team.SignupId).Error; err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	dinner.Teams = append(dinner.Teams, team)

	if err := db.Save(&dinner).Error; err != nil {
		log.Println(err)
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusCreated, team)
}

func getTeam(c *gin.Context) {
	teamId := c.Param("teamId")

	var team Team
	if err := db.Preload("CourseMatches").Preload("Courses").Preload("Teams").First(&team, &Team{TeamId: teamId}).Error; err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	c.JSON(http.StatusOK, team)
}

func updateTeam(c *gin.Context) {
	var team Team

	if err := c.BindJSON(&team); err != nil {
		log.Println(err)
		c.Status(http.StatusBadRequest)
		return
	}

	if err := db.Model(&team).Omit("SignupId", "CourseMatches").Updates(&team).Error; err != nil {
		log.Println(err)
		c.Status(http.StatusNotFound)
		return
	}

	c.JSON(http.StatusOK, team)
}

func deleteTeam(c *gin.Context) {
	teamId := c.Param("teamId")

	if err := db.Delete(&Team{}, &Team{TeamId: teamId}).Error; err != nil {
		log.Println(err)
		c.Status(http.StatusNotFound)
		return
	}

	c.Status(http.StatusOK)
}

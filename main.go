package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

var db *gorm.DB
var err error

// Load struct
type Load struct {
	PK         string    `json:"pk" gorm:"primary_key"`
	ID         uint      `json:"id,string" gorm:"index"`
	CustomerID uint      `json:"customer_id,string" gorm:"index"`
	LoadAmount float64   `json:"load_amount,string"`
	Time       time.Time `json:"time" gorm:"index"`
	Accepted   bool      `json:"accepted" gorm:"index"`
	Ignored    bool
}

// OutputLoad struct
type OutputLoad struct {
	ID         uint `json:"id,string"`
	CustomerID uint `json:"customer_id,string"`
	Accepted   bool `json:"accepted"`
}

func main() {
	SetupServer().Run()
	defer db.Close()
}

// SetupServer is to initialize the server
func SetupServer() *gin.Engine {
	Connect()
	r := gin.Default()
	r.Use(cors.Default())
	r.POST("/validateLoads", ValidateLoads)

	return r
}

// Connect is to connect the database and migrate model
func Connect() {
	db, err = gorm.Open("sqlite3", "./MySQLite.db")
	if err != nil {
		fmt.Println(err)
	}
	db.AutoMigrate(&Load{})
}

// ValidateLoads is to validate loads and generate a newOutput.txt file from input file
func ValidateLoads(c *gin.Context) {
	var writeData string
	//use transcation
	tx := db.Begin()
	// clear the db
	tx.Model(&Load{}).Delete(&Load{})
	file, _, _ := c.Request.FormFile("file")

	defer file.Close()
	// read from text file line by line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		// load json to strct
		var load Load
		// remove $ sign
		json.Unmarshal([]byte(strings.Replace(scanner.Text(), "$", "", -1)), &load)
		//validate current load
		modifiedLoad := ValidateLoad(load, tx)
		if err := tx.Create(&modifiedLoad).Error; err != nil {
			tx.Rollback()
		}
		if modifiedLoad.Ignored != true {
			var outputLoad OutputLoad
			outputLoad.ID = modifiedLoad.ID
			outputLoad.CustomerID = modifiedLoad.CustomerID
			outputLoad.Accepted = modifiedLoad.Accepted
			out, err := json.Marshal(outputLoad)
			if err != nil {
				log.Fatal(err)
			}
			writeData += string(out) + "\n"
		}
	}
	//end transcation
	tx.Commit()
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	//write to a text file
	writeToFile(writeData)

	c.JSON(200, writeData)
}

// writeToFileis to write text to newOutput.txt on root directory
func writeToFile(text string) {
	writefile, err := os.Create("./newOutput.txt")
	if err != nil {
		fmt.Println(err)
	}
	defer writefile.Close()

	_, err = io.WriteString(writefile, text)
	if err != nil {
		fmt.Println(err)
	}
	writefile.Sync()

	return
}

// ValidateLoad will validate a single load agast the rules
func ValidateLoad(load Load, tx *gorm.DB) Load {
	dailyMaxAmount := 5000.0
	weeklyMaxAmount := 20000.0
	dailyMaxCount := 3
	var loads []Load
	tx.Where("id = ? AND customer_id = ?", load.ID, load.CustomerID).First(&loads)
	// if loads have the same id, all but the first instance can be ignored
	if len(loads) == 1 {
		load.Ignored = true
		load.Accepted = false

		return load
	}
	// not use aggregation to save db call
	tx.Where("accepted = true AND customer_id = ? AND time BETWEEN ? AND ?", load.CustomerID, time.Date(load.Time.Year(), load.Time.Month(), load.Time.Day(), 0, 0, 0, 0, load.Time.Location()), load.Time).Find(&loads)
	// rule 3: A maximum of 3 loads can be performed per day, regardless of amount
	if len(loads) >= dailyMaxCount {
		load.Accepted = false

		return load
	}
	// rule 1: A maximum of $5,000 can be loaded per day
	dailyTotal := load.LoadAmount
	for i := 0; i < len(loads); i++ {
		dailyTotal += loads[i].LoadAmount
	}
	if dailyTotal > dailyMaxAmount {
		load.Accepted = false

		return load
	}
	// rule 2: A maximum of $20,000 can be loaded per week
	var weeklyTotal float64
	row := tx.Table("loads").Select("sum(load_amount) as total").Where("accepted = true AND customer_id = ? AND time BETWEEN ? AND ?", load.CustomerID, GetMonday(load.Time), load.Time).Group("customer_id").Row()
	row.Scan(&weeklyTotal)
	if weeklyTotal+load.LoadAmount > weeklyMaxAmount {
		load.Accepted = false

		return load
	}
	load.PK = strconv.FormatUint(uint64(load.ID), 10) + "-" + strconv.FormatUint(uint64(load.CustomerID), 10)
	load.Accepted = true

	return load
}

// GetMonday is to get this monday date by a givenTime
func GetMonday(givenTime time.Time) time.Time {
	offset := int(time.Monday - givenTime.Weekday())
	// sunday
	if offset > 0 {
		offset = -6
	}

	return time.Date(givenTime.Year(), givenTime.Month(), givenTime.Day(), 0, 0, 0, 0, time.Local).AddDate(0, 0, offset)
}

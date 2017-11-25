package main

import (
	"bytes"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"image"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	_ "image/jpeg"
	_ "image/png"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func main() {
	db, err := sql.Open("mysql", "user:root@tcp(127.0.0.1:3306)/petshop")
	if err != nil {
		fmt.Print(err.Error())
	}

	defer db.Close()

	err = db.Ping()
	if err != nil {
		fmt.Print(err.Error())
	}
	type Pet struct {
		Id    int
		Name  string
		Age   int
		Photo string
	}
	type PetModel struct {
		Id        int    `json:"id"`
		Name      string `json:"name"`
		Age       int    `json:"age"`
		PhotoData string `json:"photoData"`
	}

	router := gin.Default()

	//router

	router.GET("/pet/:id", func(c *gin.Context) {
		var (
			pet    Pet
			result gin.H
		)
		id := c.Param("id")
		rows := GetDataByID(db, id)
		err = rows.Scan(&pet.Id, &pet.Name, &pet.Age, &pet.Photo)
		if err != nil {
			// If no results send null
			result = gin.H{
				"result": nil,
				"count":  0,
			}
		} else {
			result = gin.H{
				"result": pet,
				"count":  1,
			}
		}
		c.JSON(http.StatusOK, result)
	})

	router.POST("/pet", func(c *gin.Context) {
		var petModel PetModel
		c.BindJSON(&petModel)
		stmt, err := db.Prepare("call usp_pet_insert(?,?);")
		if err != nil {
			fmt.Print(err.Error())
		}
		_, err = stmt.Exec(petModel.Name, petModel.Age)

		if err != nil {
			fmt.Print(err.Error())
		}

		defer stmt.Close()
		c.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf(" %s successfully created", petModel.Name),
		})
	})

	router.GET("/pets", func(c *gin.Context) {
		var (
			pet  Pet
			pets []Pet
		)

		rows, err := db.Query("select id, name, age, photo from pet")

		if err != nil {
			fmt.Print(err.Error())
		}

		for rows.Next() {
			err = rows.Scan(&pet.Id, &pet.Name, &pet.Age, &pet.Photo)
			pets = append(pets, pet)
			if err != nil {
				fmt.Print(err.Error())
			}
		}
		defer rows.Close()
		c.JSON(http.StatusOK, gin.H{
			"result": pets,
			"count":  len(pets),
		})
	})

	router.PUT("/pet/:id", func(c *gin.Context) {

		id := c.Param("id")
		var petModel PetModel
		c.BindJSON(&petModel)
		stmt, err := db.Prepare("update pet set name= ?, age= ? where id= ?;")
		if err != nil {
			fmt.Print(err.Error())
		}

		_, err = stmt.Exec(petModel.Name, petModel.Age, id)
		if err != nil {
			fmt.Print(err.Error())
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Data berhasil di update",
		})

	})

	router.DELETE("/pet/:id", func(c *gin.Context) {
		id := c.Param("id")

		var pet Pet
		rows := GetDataByID(db, id)
		err = rows.Scan(&pet.Id, &pet.Name, &pet.Age, &pet.Photo)
		if err != nil {
			fmt.Print(err.Error())
		} else {

			//data exist
			stmt, err := db.Prepare("delete from pet where id= ?")
			if err != nil {
				fmt.Print(err.Error())
			}

			_, err = stmt.Exec(id)

			if err != nil {
				fmt.Print(err.Error())
			}
			deletePhoto(pet.Photo)
			c.JSON(http.StatusOK, gin.H{
				"message": "Data Berhasil di hapus",
			})
		}
	})

	router.POST("/pet/:id/uploadImage", func(c *gin.Context) {
		id := c.Param("id")

		var petModel PetModel
		c.BindJSON(&petModel)

		photo := "photo-" + petModel.Name

		photo, err := saveImageToDisk(photo, petModel.PhotoData)
		stmt, err := db.Prepare("update pet set photo =? where id=?")

		if err != nil {
			fmt.Print(err.Error())
		}

		_, err = stmt.Exec(photo, id)

		c.JSON(http.StatusOK, gin.H{"message": "data foto berhasil diupload"})
	})

	router.Run(":3000")
}

func GetDataByID(db *sql.DB, ID string) (row *sql.Row) {
	row = db.QueryRow("select id, name, age, photo from pet where id = ?;", ID)
	return row
}

func saveImageToDisk(fileNameBase, data string) (string, error) {

	var (
		// ErrSize         = errors.New("Invalid size!")
		ErrInvalidImage = errors.New("Invalid image!")
	)

	idx := strings.Index(data, ";base64,")
	if idx < 0 {
		return "", ErrInvalidImage
	}
	reader := base64.NewDecoder(base64.StdEncoding, strings.NewReader(data[idx+8:]))
	buff := bytes.Buffer{}
	_, err := buff.ReadFrom(reader)
	if err != nil {
		return "", err
	}

	_, fm, err := image.DecodeConfig(bytes.NewReader(buff.Bytes()))
	if err != nil {
		return "", err
	}
	folder := "upload"
	dir, err := os.Getwd()
	folder = dir + "/" + folder
	if _, err := os.Stat(folder); os.IsNotExist(err) {
		os.Mkdir(folder, os.FileMode(0700))
	}

	fileName := fileNameBase + "." + fm
	ioutil.WriteFile(fileName, buff.Bytes(), 0644)

	return fileName, err
}

func deletePhoto(fileNameBase string) {
	folder := "upload"
	dir, _ := os.Getwd()
	folder = dir + "/" + folder
	fileName := folder + "/" + fileNameBase
	os.Remove(fileName)
}

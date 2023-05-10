package main

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"
	"encoding/json"
	"io/ioutil"
	"crypto/sha256"
	"encoding/hex"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

const (
	ImgDir = "images"
)

type Response struct {
	Message string `json:"message"`
}

type Item struct {
	ID int `json:"id"`
	Name     string `json:"name"`
	Category string `json:"category"`
	Image string `json:"image_filename"`
}

type Items struct {
	Items []Item `json:"items"`
}

func root(c echo.Context) error {
	res := Response{Message: "Hello, world!"}
	return c.JSON(http.StatusOK, res)
}

func addItem(c echo.Context) error {
	// Get form data
	var newItem Item
	newItem.Name = c.FormValue("name")
	newItem.Category = c.FormValue("category")
	imagePath := c.FormValue("image")
	hash := calculateImageHash(imagePath)
	newItem.Image = hash

	// Add new item to existing items
	existingItems := loadItemsFromJSON()
	count := len(existingItems.Items)+1
	newItem.ID = count
	existingItems.Items = append(existingItems.Items, newItem)

	// Save data to JSON
	saveItemFromJSON(existingItems)

	c.Logger().Infof("Receive item: %s", newItem.Name)
	message := fmt.Sprintf("item received: %s", newItem.Name)
	res := Response{Message: message}
	
	return c.JSON(http.StatusOK, res)
}

func getImg(c echo.Context) error {
	// Create image path
	imgPath := path.Join(ImgDir, c.Param("imageFilename"))

	if !strings.HasSuffix(imgPath, ".jpg") {
		res := Response{Message: "Image path does not end with .jpg"}
		return c.JSON(http.StatusBadRequest, res)
	}
	if _, err := os.Stat(imgPath); err != nil {
		c.Logger().Debugf("Image not found: %s", imgPath)
		imgPath = path.Join(ImgDir, "default.jpg")
	}
	return c.File(imgPath)
}

func getItem(c echo.Context) error {
	// Read JSON file
	items := loadItemsFromJSON()
	c.Logger().Infof("Get items")

	return c.JSON(http.StatusOK, items)
}

func loadItemsFromJSON() Items{
	// Read JSON file
	data, _ := ioutil.ReadFile("items.json")

	// Parse JSON into Items struct
	var items Items
	_ = json.Unmarshal(data, &items)

	return items
}

func saveItemFromJSON(items Items) {
	// Save data to JSON
	data, _ := json.Marshal(items)
	_ = ioutil.WriteFile("items.json", data, 0644)
}

func calculateImageHash(filePath string) string {
	// Read image file
	imageData, _ := ioutil.ReadFile(filePath)

	// Calculate SHA256 hash
	hash := sha256.Sum256(imageData)

	// Convert hash to hexadecimal string
	hashString := hex.EncodeToString(hash[:]) + ".jpeg"

	return hashString
}

func getItemByID(c echo.Context) error {
	targetStr := c.Param("id")
	targetID, _ := strconv.Atoi(targetStr)
	items := loadItemsFromJSON()

	for _, item := range items.Items {
		if item.ID == targetID {
			return c.JSON(http.StatusOK, item)
		} 
	}

	return c.String(http.StatusOK, "Item not found \n")
}

func main() {
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Logger.SetLevel(log.INFO)

	front_url := os.Getenv("FRONT_URL")
	if front_url == "" {
		front_url = "http://localhost:3000"
	}
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{front_url},
		AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
	}))

	// Routes
	e.GET("/", root)
	e.POST("/items", addItem)
	e.GET("/image/:imageFilename", getImg)
	e.GET("/items", getItem)
	e.GET("/items/:id", getItemByID)


	// Start server
	e.Logger.Fatal(e.Start(":9000"))
}

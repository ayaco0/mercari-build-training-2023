package main

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"
	"encoding/json"
	"crypto/sha256"
	"strconv"
	"database/sql"
	"io"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	_ "github.com/mattn/go-sqlite3" 
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
	image, err := c.FormFile("image")
	if err != nil {
		return err
	}
	hash, _ := calculateImageHash(imagePath)
	newItem.Image = hash

	// Save image
	dstPath := path.Join(ImgDir, hash)
	dstFile, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	srcFile, err := image.Open()
	if err != nil {
		return err
	}
	defer srcFile.Close()
	// ioはOK？
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}

	saveItemToDB(newItem)

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
	// Load JSON/DB
	// items, _ := loadItemsFromJSON()
	items, _ := loadItemsFromDB("")
	c.Logger().Infof("Get items")

	return c.JSON(http.StatusOK, items)
}

func loadItemsFromJSON() (Items, error) {
	// Read JSON file
	data, err := os.ReadFile("items.json")
	if err != nil {
		return Items{}, err
	}
	// Parse JSON into Items struct
	var items Items
	err = json.Unmarshal(data, &items)
	if err != nil {
		return Items{}, err
	}

	return items, nil
}

func saveItemToJSON(items Items) error {
	// Save data to JSON
	data, err := json.Marshal(items)
	if err != nil {
		return err
	}
	err = os.WriteFile("items.json", data, 0644)
	if err != nil {
		return err
	}
	return nil
}

func calculateImageHash(filePath string) (string, error) {
	// Read image file
	imageData, _ := os.ReadFile(filePath)

	// Calculate SHA256 hash
	hash := sha256.Sum256(imageData)

	// Convert hash to hexadecimal string
	hashString := fmt.Sprintf("%x%s", hash, ".jpg")

	return hashString, nil
}

func getItemByID(c echo.Context) error {
	targetStr := c.Param("id")
	targetID, _ := strconv.Atoi(targetStr)
	items, _ := loadItemsFromDB("")

	for _, item := range items.Items {
		if item.ID == targetID {
			return c.JSON(http.StatusOK, item)
		} 
	}

	return c.String(http.StatusOK, "Item not found \n")
}

func saveItemToDB(item Item) error {
	// Connect DB
	db, err := sql.Open("sqlite3", "../db/mercari.sqlite3")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()	
	// Search categoryID
	var category_id int
	row := db.QueryRow("SELECT ROWID FROM category WHERE name == $1", item.Category)
	err = row.Scan(&category_id)
	// Insert item to table
	cmd := "INSERT INTO items (name, category_id, image_name) VALUES ($1, $2, $3)"
	result, err := db.Exec(cmd, item.Name, category_id, item.Image)
	if err != nil {
		log.Fatal(err)
	}
	// Return value of the lastID
	id, err := result.LastInsertId()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("LastInsertId: %d\n", id)
	return nil
}

func loadItemsFromDB(keyword string) (Items,error) {
	var items Items
	// Connect DB
	db, err := sql.Open("sqlite3", "../db/mercari.sqlite3")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	if keyword == "" {
		// Load all items
		cmd := "SELECT items.id, items.name, category.name as category, items.image_name FROM items JOIN category ON items.category_id = category.id"
		rows, err := db.Query(cmd)
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()
		for rows.Next() {
			var item Item
			err := rows.Scan(&item.ID, &item.Name, &item.Category, &item.Image)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(item.ID, item.Name, item.Category, item.Image)
			items.Items = append(items.Items, item)
		}
	} else {
		// Search items by keyword
		cmd := "SELECT items.id, items.name, category.name as category, items.image_name FROM items JOIN category ON items.category_id = category.id WHERE items.name LIKE '%'||$1||'%'"
		rows, err := db.Query(cmd, keyword)
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()
		for rows.Next() {
			var item Item
			err := rows.Scan(&item.ID, &item.Name, &item.Category, &item.Image)
			if err != nil {
				log.Fatal(err)
			}
			items.Items = append(items.Items, item)
		}
	}
	return items, nil
}

func searchItem(c echo.Context) error {
	keyword := c.QueryParam("keyword")
	items, err := loadItemsFromDB(keyword)
	if err != nil {
		return c.String(http.StatusOK, "Item not found \n")
	}
	return c.JSON(http.StatusOK, items)
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
	e.GET("/search", searchItem)

	// Start server
	e.Logger.Fatal(e.Start(":9000"))
}

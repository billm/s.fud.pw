package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"crypto/md5"
	"encoding/hex"

	"github.com/gin-gonic/gin"
	_ "github.com/heroku/x/hmetrics/onload"
	_ "github.com/lib/pq"
)

func shorten(db *sql.DB, url string) string {

	hasher := md5.New()
	hasher.Write([]byte(url))
	slug := hex.EncodeToString(hasher.Sum(nil))[0:7]

	if _, err := db.Exec("CREATE TABLE IF NOT EXISTS slugs (slug varchar, url varchar)"); err != nil {
		return err.Error()
	}

	if _, err := db.Exec("INSERT INTO slugs VALUES ($1, $2)", slug, url); err != nil {
		return err.Error()
	}

	return slug
}

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		log.Fatal("$PORT must be set")
	}

	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Error opening database: %q", err)
	}

	router := gin.New()
	router.Use(gin.Logger())
	router.LoadHTMLGlob("templates/*.tmpl.html")
	router.Static("/static", "static")

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl.html", nil)
	})
	router.GET("/r/:slug", func(c *gin.Context) {
		//c.Redirect(http.StatusMovedPermanently, "http://www.google.com/")

		rows, err := db.Query("SELECT url FROM slugs WHERE slug = ?", c.Param("slug"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, err.Error())
			return
		}
		defer rows.Close()

		for rows.Next() {
			var (
				url string
			)
			if err := rows.Scan(&url); err != nil {
				c.JSON(http.StatusInternalServerError, err.Error())
				return
			}
			c.Redirect(http.StatusTemporaryRedirect, url)
			return
		}

	})
	router.PUT("/edit/:slug", func(c *gin.Context) {
		//c.Redirect(http.StatusMovedPermanently, "http://www.google.com/")
		c.JSON(http.StatusOK, gin.H{"slug": c.Param("slug")})
	})
	router.POST("/new", func(c *gin.Context) {
		slug := shorten(db, c.PostForm("url"))

		c.JSON(http.StatusOK, gin.H{"slug": slug})
	})

	router.Run(":" + port)
}

package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"strconv"

	"crypto/md5"
	"encoding/hex"

	"github.com/gin-gonic/gin"
	_ "github.com/heroku/x/hmetrics/onload"
	_ "github.com/lib/pq"
)

// Initialize database
var db = initDb()

func initDb() *sql.DB {
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Error opening database: %q", err)
	}

	if _, err := db.Exec("CREATE TABLE IF NOT EXISTS slugs ( " +
		"slug varchar (8) NOT NULL UNIQUE, " +
		"gurl varchar NOT NULL, " +
		"gcount int NOT NULL DEFAULT 0, " +
		"burl varchar, " +
		"bcount int NOT NULL DEFAULT 0, " +
		"created_at timestamp NOT NULL DEFAULT NOW(), " +
		"expires_at timestamp NOT NULL DEFAULT NOW() + interval '10 years'" +
		")"); err != nil {
		log.Fatal(err.Error())
	}

	return db
}

func shorten(db *sql.DB, gurl string, burl string, gcount int, bcount int) string {

	hasher := md5.New()
	hasher.Write([]byte(gurl))
	slug := hex.EncodeToString(hasher.Sum(nil))[0:8]

	if _, err := db.Exec("INSERT INTO slugs(slug, gurl, burl, gcount, bcount) VALUES ($1, $2, $3, $4, $5)", slug, gurl, burl, gcount, bcount); err != nil {
		return err.Error()
	}

	return slug
}

func handleNewSlug(c *gin.Context) {
	gurl := c.PostForm("gurl")
	burl := c.PostForm("burl")
	gcount, err := strconv.Atoi(c.PostForm("gcount"))
	if err != nil {
		gcount = 0
	}
	bcount, err := strconv.Atoi(c.PostForm("bcount"))
	if err != nil {
		bcount = 0
	}

	slug := shorten(db, gurl, burl, gcount, bcount)

	c.JSON(http.StatusOK, gin.H{"slug": slug})
}

func handleGetSlug(c *gin.Context) {
	//c.Redirect(http.StatusMovedPermanently, "http://www.google.com/")

	rows, err := db.Query("SELECT gurl, burl, gcount, bcount FROM slugs WHERE slug = $1", c.Param("slug"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {
		var (
			gurl   string
			burl   string
			gcount int
			bcount int
		)
		if err := rows.Scan(&gurl, &burl, &gcount, &bcount); err != nil {
			c.JSON(http.StatusInternalServerError, err.Error())
			return
		}

		// If we still have 'good' serves left, redirect to a good url
		if 0 < gcount {
			gcount--
			db.Query("UPDATE slugs set gcount = $1 where slug = $2", gcount, c.Param("slug"))
			c.Redirect(http.StatusTemporaryRedirect, gurl)
			return
		}
		// If we still have 'bad' serves left, redirect to the bad url
		if 0 < bcount {
			bcount--
			db.Query("UPDATE slugs set bcount = $1 where slug = $2", bcount, c.Param("slug"))
			c.Redirect(http.StatusTemporaryRedirect, burl)
			return
		}

		// Fall back to always returning the good url
		c.Redirect(http.StatusTemporaryRedirect, gurl)
		return
	}

}

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		log.Fatal("$PORT must be set")
	}

	router := gin.New()
	router.Use(gin.Logger())
	router.LoadHTMLGlob("templates/*.tmpl.html")
	router.Static("/static", "static")

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl.html", nil)
	})
	router.GET("/r/:slug", handleGetSlug)

	router.PUT("/edit/:slug", func(c *gin.Context) {
		//c.Redirect(http.StatusMovedPermanently, "http://www.google.com/")
		c.JSON(http.StatusOK, gin.H{"slug": c.Param("slug")})
	})
	router.POST("/new", handleNewSlug)

	router.Run(":" + port)
}

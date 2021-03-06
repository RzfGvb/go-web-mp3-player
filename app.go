package main

import (
	"fmt"
	"io/ioutil"

	"golang.org/x/oauth2"

	"log"
	"net/http"

	"time"

	"io"

	"gopkg.in/gin-gonic/gin.v1"
)

func initApp(router *gin.Engine) {
	// APP
	g := router.Group("/")
	g.Use(checkCookie)
	g.GET("/", handleRoot)
	g.GET("/search", handleSearch)
	g.GET("/file/:id", handleFileGet)
	router.GET("/reg", handleReg)

	router.Static("/site", "./site/")
	router.Static("/static", "./site/static")
	router.Static("/jplayer", "./site/jplayer")
	router.Static("/skin/blue.monday/css", "./site/skin/blue.monday/css")
	router.Static("/skin/blue.monday/image", "./site/skin/blue.monday/image")
	router.Static("/skin/blue.monday/mustache", "./site/skin/blue.monday/mustache")

	router.LoadHTMLGlob("templates/*")
}

//--------------------------
//-- Main Website Handlers
//--------------------------

func handleRoot(c *gin.Context) {
	id, _ := c.Cookie("JAMPY_USER_ID")
	fs := getFiles(id)
	if fs == nil {
		fmt.Println()
		c.SetCookie(
			"JAMPY_USER_ID",
			"none",
			-1,
			"",
			"",
			false,
			false,
		)
		fmt.Println("REDIR")
		c.Redirect(http.StatusPermanentRedirect, "/")
	} else {
		c.HTML(200, "s2.html", gin.H{
			"files":        fs,
			"search_field": "",
			"text":         "No files found on yor drive.",
			"num_files":    len(fs),
		})
	}
}

func handleSearch(c *gin.Context) {
	fmt.Println("SEARCHING.....")
	id, _ := c.Cookie("JAMPY_USER_ID")
	name := c.Query("name")
	tag := c.Query("tag")
	fs := makeSearch(id, name, tag)
	if fs == nil {
		fmt.Println()
		c.SetCookie(
			"JAMPY_USER_ID",
			"none",
			-1,
			"",
			"",
			false,
			false,
		)
		fmt.Println("REDIR")
		c.Redirect(http.StatusPermanentRedirect, "/")
	} else {
		var req string
		if name != "" {
			req = name
		} else {
			req = "#" + tag
		}
		c.HTML(200, "s2.html", gin.H{
			"files":        fs,
			"search_field": req,
			"text":         "No songs were found.",
			"num_files":    len(fs),
		})
	}
}

func handleReg(c *gin.Context) {
	code := c.Query("code")
	fmt.Println("getting")
	resp, err := http.Get("http://" + c.Request.Host + "/api/new?code=" + code)
	fmt.Println("GOT...")
	if err != nil {
		log.Fatal(err.Error())
	}
	idb, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err.Error())
	}
	id := string(idb)
	c.SetCookie(
		"JAMPY_USER_ID",
		id,
		int(time.Hour*24*30),
		"",
		"",
		false,
		false,
	)
	fmt.Println("Cookies:")
	fmt.Println("IP:", c.ClientIP())
	c.Redirect(http.StatusPermanentRedirect, "http://"+c.Request.Host+"/")
}

func handleFileGet(c *gin.Context) {
	id := c.Param("id")
	user, err := c.Cookie("JAMPY_USER_ID")
	if err != nil {
		log.Fatal(err.Error())
	}
	serv := services[user]
	if serv == nil {
		fmt.Println("FAIL")
		c.JSON(204, gin.H{})
	}
	resp, err := serv.Files.Get(id).Download()
	if err != nil {
		fmt.Println(err.Error())
	}
	io.Copy(c.Writer, resp.Body)
}

func checkCookie(c *gin.Context) {
	_, err := c.Cookie("JAMPY_USER_ID")
	if err != nil {
		fmt.Println("=========NO COOKIE============")
		fmt.Println(c.Request.Cookies())
		fmt.Println(c.Request.Host)
		config.RedirectURL = "http://" + c.Request.Host + "/reg"
		authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline, oauth2.AccessTypeOnline)
		c.Redirect(http.StatusTemporaryRedirect, authURL)
	}
}

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"encoding/json"

	"io"

	"github.com/boltdb/bolt"
	"golang.org/x/oauth2"
	"google.golang.org/api/drive/v3"
	"gopkg.in/gin-gonic/gin.v1"
)

func initAPI(engine *gin.Engine) {
	r := engine.Group("/api")
	// API
	r.GET("/link", handleLinkApi)
	r.GET("/new", handleNewApi)
	r.POST("/files", handleFilesApi)
	r.POST("/files/:id/:tag", handleAddTagApi)
	r.DELETE("/files/:id/:tag", handleDeleteTagApi)
	r.POST("/file/:id", handleFileApi)
	r.POST("/search", handleSearchApi)
	r.GET("/alive", handleAliveApi)
}

//-----------------------
//-- API Handlers
//-----------------------

func handleLinkApi(c *gin.Context) {
	config.RedirectURL = c.Query("link")
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Fprint(c.Writer, authURL)
}

func handleNewApi(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		log.Println("No code")
	}
	var err error
	tok, err = config.Exchange(ctx, code)
	if err != nil {
		log.Printf("Unable to retrieve token from web %v", err)
		return
	}
	tb, err := json.Marshal(tok)
	if err != nil {
		log.Printf("Unable to retrieve token from web %v", err)
	}
	ioutil.WriteFile("tok.json", tb, 0600)
	client = config.Client(ctx, tok)
	service, err = drive.New(client)
	x, err := service.About.Get().Fields("user(permissionId)").Do()
	if err != nil {
		fmt.Println("err: ", err.Error())
	}
	id := x.User.PermissionId
	services[id] = service
	if err != nil {
		log.Fatalf("Unable to retrieve drive Client %v", err)
	}
	idb := []byte(id)
	db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(idb))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		b.CreateBucket([]byte("files"))

		tk, err := json.Marshal(tok)
		if err != nil {
			return fmt.Errorf("marshal json: %s", err)
		}
		err = b.Put([]byte("token"), tk)
		if err != nil {
			return fmt.Errorf("Put token: %s", err)
		}
		return nil
	})
	c.Writer.Write(idb)
}

func getFiles(user string) []*file {
	serv := services[user]
	if serv == nil {
		fmt.Println("FAIL")
		return nil
	}
	filenames := make([]*file, 0, 100)
	var f1 *file
	serv.Files.List().
		Fields("nextPageToken, files(id, name)").
		Q("mimeType='audio/mpeg'").
		Pages(ctx, func(fs *drive.FileList) error {
			for _, f := range fs.Files {
				f1 = &file{
					Id:   f.Id,
					Name: f.Name,
					Tags: []string{},
				}
				filenames = append(filenames, f1)
			}
			return nil
		})
	db.View(func(tx *bolt.Tx) error {
		root := tx.Bucket([]byte(user))
		b := root.Bucket([]byte("files"))
		if b == nil {
			return nil
		}
		for _, f := range filenames {
			v := b.Get([]byte(f.Id))
			if v == nil {
				continue
			}
			var tags []string
			tags = make([]string, 0, 10)
			err := json.Unmarshal(v, &tags)
			if err != nil {
				return err
			}
			f.Tags = tags
		}
		return nil
	})
	return filenames
}

func handleFilesApi(c *gin.Context) {
	idb, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		fmt.Println("err: ", err.Error())
	}
	id := string(idb)
	fs := getFiles(id)
	if fs == nil {
		c.JSON(204, gin.H{})
	}
	c.JSON(200, fs)
}

func makeSearch(id, name, tag string) []*file {
	fnames := getFiles(id)
	if fnames == nil {
		return nil
	}
	foundfs := make([]*file, 0, len(fnames))
	for _, f := range fnames {
		if name != "" && strings.Contains(strings.ToLower(f.Name), name) {
			foundfs = append(foundfs, f)
		}
		if tag != "" {
			for _, t := range f.Tags {
				if t == tag {
					foundfs = append(foundfs, f)
					break
				}
			}
		}
	}
	return foundfs
}

func handleSearchApi(c *gin.Context) {
	name := c.Query("name")
	tag := c.Query("tag")
	fmt.Printf("N==%s,T==%s\n", name, tag)
	idb, e := ioutil.ReadAll(c.Request.Body)
	if e != nil {
		fmt.Println("err: ", e.Error())
	}
	id := string(idb)
	fnames := makeSearch(id, name, tag)
	if fnames == nil {
		c.JSON(204, gin.H{})
	}
	c.JSON(200, fnames)
}

func handleAddTagApi(c *gin.Context) {
	fmt.Println("Handling tags")
	songb := c.Param("id")
	song := []byte(songb)
	tag := c.Param("tag")
	user, e := ioutil.ReadAll(c.Request.Body)
	if e != nil {
		fmt.Println("err: ", e.Error())
	}
	fmt.Printf("USER: %s, tag: %s, id: %s\n", string(user), tag, songb)

	err := db.Update(func(tx *bolt.Tx) error {
		root := tx.Bucket(user)
		if root == nil {
			fmt.Println("ops")
		}
		b := root.Bucket([]byte("files"))
		if b == nil {
			fmt.Println("ops2")
		}
		v := b.Get(song)
		if v == nil {
			tags := []string{tag}
			v, err := json.Marshal(tags)
			if err != nil {
				fmt.Println("22")
				return err
			}
			err = b.Put(song, v)
			if err != nil {
				fmt.Println("Didnt put")
				return err
			}
			fmt.Println("Tags: ", tags)
			return nil
		}
		var tags []string
		err := json.Unmarshal(v, &tags)
		if err != nil {
			fmt.Println("321")
			return err
		}
		for _, t := range tags {
			if t == tag {
				fmt.Println("nope")
				return nil
			}
		}
		tags = append(tags, tag)
		v, err = json.Marshal(tags)
		if err != nil {
			fmt.Println("ahao")
			return err
		}
		err = b.Put(song, v)
		if err != nil {
			fmt.Println("ewq")
			return err
		}
		fmt.Println("Tags2: ", tags)
		return nil
	})

	if err != nil {
		fmt.Println("SMTH")
		fmt.Println(err.Error())
	}
}

func handleDeleteTagApi(c *gin.Context) {
	fmt.Println("Handling tags")
	songb := c.Param("id")
	song := []byte(songb)
	tag := c.Param("tag")
	user, e := ioutil.ReadAll(c.Request.Body)
	if e != nil {
		fmt.Println("err: ", e.Error())
	}
	fmt.Printf("USER: %s, tag: %s, id: %s\n", string(user), tag, songb)

	err := db.Update(func(tx *bolt.Tx) error {
		root := tx.Bucket(user)
		if root == nil {
			fmt.Println("ops")
		}
		b := root.Bucket([]byte("files"))
		if b == nil {
			fmt.Println("ops2")
		}
		v := b.Get(song)
		if v == nil {
			return nil
		}
		var tags []string
		err := json.Unmarshal(v, &tags)
		if err != nil {
			fmt.Println("321")
			return err
		}
		for i, t := range tags {
			if t == tag {
				tags = append(tags[:i], tags[i+1:]...)
			}
		}
		v, err = json.Marshal(tags)
		if err != nil {
			fmt.Println("ahao")
			return err
		}
		err = b.Put(song, v)
		if err != nil {
			fmt.Println("ewq")
			return err
		}
		fmt.Println("Tags2: ", tags)
		return nil
	})

	if err != nil {
		fmt.Println("SMTH")
		fmt.Println(err.Error())
	}
}

func handleFileApi(c *gin.Context) {
	id := c.Param("id")
	idb, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		fmt.Println("err: ", err.Error())
	}
	user := string(idb)
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

func handleAliveApi(c *gin.Context) {
	//c.String(200, "")
	c.JSON(200, "1")
}

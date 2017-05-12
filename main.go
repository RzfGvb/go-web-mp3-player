package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"fmt"

	"os"

	"os/signal"

	"github.com/boltdb/bolt"
	"github.com/braintree/manners"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/googleapi"
	"gopkg.in/gin-gonic/gin.v1"
)

var (
	config   = getConfig()
	ctx      = context.Background()
	tok      *oauth2.Token
	client   *http.Client
	service  *drive.Service
	services = make(map[string]*drive.Service)
	router   = gin.Default()
	PORT     = os.Getenv("PORT")
	db       *bolt.DB
)

func main() {
	initAPI(router)
	initApp(router)
	var err error
	db, err = bolt.Open("my.db", os.FileMode(os.O_RDWR), nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("DEBUG")
	//db.View(func(tx *bolt.Tx) error {
	//	tok := new(oauth2.Token)
	//	tx.ForEach(func(name []byte, b *bolt.Bucket) error {
	//		bt := b.Get([]byte("token"))
	//		if bt == nil || len(bt) == 0 {
	//			return nil
	//		}
	//		err := json.Unmarshal(bt, tok)
	//		if err != nil {
	//
	//		}
	//		resetClient(tok)
	//		return nil
	//	})
	//	return nil
	//})
	go func() {
		sigchan := make(chan os.Signal, 1)
		signal.Notify(sigchan, os.Interrupt, os.Kill)
		<-sigchan
		log.Println("Shutting down...")
		db.Close()
		manners.Close()
	}()
	fmt.Println("Serving on port:", PORT)
	log.Fatal(manners.ListenAndServe(":"+PORT, router))
}

//-----------------------
//-- Structs and methods
//-----------------------

type file struct {
	Id   string   `json:"id"`
	Name string   `json:"name"`
	Link string   `json:"link"`
	Tags []string `json:"tags"`
}

type fileList []file

func (fl fileList) ToJson() []byte {
	fs, err := json.Marshal(fl)
	if err != nil {
		fmt.Println("err: ", err.Error())
	}
	return fs
}

//-------------------
//-- Just functions
//-------------------

func getConfig() *oauth2.Config {
	b, err := ioutil.ReadFile("client_secret.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	config, err := google.ConfigFromJSON(b, drive.DriveScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	return config
}

func createFilesFields(fs ...string) googleapi.Field {
	//arr := make([]string, len(fs))
	//for i, s := range fs {
	//	arr[i] = s
	//}

	return googleapi.Field(
		"files(" + strings.Join(fs, ",") + ")",
	)
}

func resetClient(tok *oauth2.Token) {
	client = config.Client(ctx, tok)
	service, err := drive.New(client)
	x, err := service.About.Get().Fields("user(permissionId)").Do()
	if err != nil {
		fmt.Println("err: ", err.Error())
	}
	id := x.User.PermissionId
	services[id] = service
}

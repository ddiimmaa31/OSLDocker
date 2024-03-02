package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	gmc "github.com/bradfitz/gomemcache/memcache"
)

var (
	MemeClient *gmc.Client
	DB         *gorm.DB
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Test"))
	})

	fmt.Println(http.ListenAndServe(os.Getenv("GOSERVER_PORT"), nil))
}

func init() {
	MemeClient = gmc.New("/server")

	dsm := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Europe/UTC", os.Getenv("HOST_NAME"), os.Getenv("USER_NAME"), os.Getenv("PASSWORD"), os.Getenv("DB_NAME"), os.Getenv("PORT"))
	var err error
	DB, err = gorm.Open(postgres.Open(dsm), &gorm.Config{})

	if err != nil {
		panic(err.Error())
	}
}

func CheckCase(key string) string {
	item, err := MemeClient.Get(key)

	if err != nil {
		return ""
	}

	if item == nil {
		return ""
	}

	MemeClient.Touch(key, int32(time.Hour*24/time.Second))

	return key
}

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

type Student struct {
	Id          int    `gorm:"primaryKey;autoIncrement"`
	Ident       string `gorm:"unique"`
	NameSurname string `gorm:"unique"`
}

const (
	home     = " <a href=\"/\">back</a>"
	register = " <a href=\"/new\">try again</a>"
)

var (
	MemeClient *gmc.Client
	DB         *gorm.DB
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		ident := r.FormValue("ident")
		if ident == "" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write([]byte(`<form action="/" method="GET">
								<input id="ident" type="text" name="ident" />
								<input type="submit" value="Check" />
		  					</form>`))
			return
		}
		if meme := CheckMeme(ident); meme != "" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write([]byte("<h>" + meme + "</h>" + home))
			return
		}
		s := GetFromDB(ident)
		if s.Ident != "" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write([]byte("<h>" + s.NameSurname + "</h>" + home))
			AddMeme(s)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(`<form action="/new" method="GET">
							<input id="ident" type="text" name="ident" />
							<input id="nameSurname" type="text" name="nameSurname" />
							<input type="submit" value="Save" />
	  					</form>`))
	})

	http.HandleFunc("/new", func(w http.ResponseWriter, r *http.Request) {
		ident := r.FormValue("ident")
		nameSurname := r.FormValue("nameSurname")
		if ident == "" || nameSurname == "" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write([]byte(`<form action="/new" method="GET">
                                <input id="ident" type="text" name="ident" />
                                <input id="nameSurname" type="text" name="nameSurname" />
                                <input type="submit" value="Save" />
                            </form>`))
			return
		}
		if err := WriteToDB(Student{Ident: ident, NameSurname: nameSurname}); err != nil {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write([]byte("<h>Wrong Iednt or NameSurname</h>" + register))
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte("<h>Student was successfully saved</h>" + home))
	})

	fmt.Println(http.ListenAndServe(os.Getenv("GOSERVER_PORT"), nil))
}

func init() {
	time.Sleep(10 * time.Second)
	MemeClient = gmc.New(os.Getenv("MEMC_PATH"))

	dsm := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Europe/Moscow",
		os.Getenv("HOST_NAME"),
		os.Getenv("USER_NAME"),
		os.Getenv("PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("POST_PORT"))

	var err error
	DB, err = gorm.Open(postgres.Open(dsm), &gorm.Config{})

	DB.AutoMigrate(Student{})

	if err != nil {
		panic(err.Error())
	}
}

func CheckMeme(key string) string {
	item, err := MemeClient.Get(key)

	if err != nil {
		return ""
	}

	if item == nil {
		return ""
	}

	MemeClient.Touch(key, int32(time.Hour*24/time.Second))

	return string(item.Value)
}

func AddMeme(student Student) error {
	err := MemeClient.Add(&gmc.Item{Key: student.Ident, Value: []byte(student.NameSurname), Expiration: int32(time.Hour * 24 / time.Second), Flags: 0})

	if err != nil {
		return err
	}

	return nil
}

func GetFromDB(key string) Student {
	student := Student{}
	r := DB.Where("ident = ?", key).First(&student)

	if r.Error != nil {
		return Student{}
	}

	return student
}

func WriteToDB(student Student) error {
	r := DB.Create(&student)

	if r.Error != nil {
		return r.Error
	}
	AddMeme(student)
	return nil
}

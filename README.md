# OSL Docker
## Развертывание контейнеров Docker и Docker Compose
### 1 Написание сервера

Сервер был написан на языке `Go` с использованием стандартной библиотеки net/http. Доступ к статическим файлам осуществлялся так же через данную библиотеку.

```golang
func main() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		ident := r.FormValue("ident")

		if ident == "" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write([]byte(index(`<form action="/" method="GET">
								...
		  					</form>`)))
			return
		}

		if meme := CheckMeme(ident); meme != "" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write([]byte(index("<p>" + meme + "</p>" + home)))
			return
		}
		s := GetFromDB(ident)
		if s.Ident != "" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write([]byte(index("<p>" + s.NameSurname + "</p>" + home)))
			AddMeme(s)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(index(`<form action="/new" method="GET">
							...
	  					</form>`)))
	})
```

В качестве БД был использован PostgreSQL, в качестве кэша - Memcached. 
Взаимодействие между БД и кэшом осуществлялось при помощи библиотек gorm и gomemcached соответственно.
Все параметры были вынесены в переменный среды и были заданы в файле[ docker-compose.yaml](docker-compose.yaml).


```golang
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
```

### 2 Описание Dockerfile

В соответствии с практикой написания Dockerfile под приложения написанные на языке Go, за основу был взят образ `golang:1.22-alpine`. 

Компиляция проекта происходит в процессе создания образа Docker контейнера. 
Сервер использует порт `:3000`. Т.к. в результате компиляции Go-приложения получаем единый исполняемый файл с именем соответствующим наименованию исполняемого `.go` файла. 
Заключительная команда - команда запуска сервера.

```Dockerfile
FROM golang:1.22-alpine

RUN mkdir -p server

WORKDIR /server

COPY . .
RUN go mod download

RUN go build -o /server

EXPOSE 3000

CMD [ "./server" ]
```

Для создания образа используем команду `docker build --no-cache -t server_go .`
* `docker build` - создание образа
* `--no-cache` - без использования кэша
* `-t server_go` - присваивание названия образу (тега)

### 3 Docker-compose

Для создания контейнера, содержащего необходимые образы, в файл `Docker-compose.yaml` были прописаны необходимые переменные:
* версия `Docker-compose`
* используемые образы
* для каждого образа - используемые переменные  
* Хитрый момент 1. Команда `depends on` - позволяет запускать образ после запуска указанных контейнеров
* Хитрый момент 2. ` MEMC_PATH: memecached:11211` данный путь прописывается в формате { *имя сервиса* } : *порт*

```yaml
version: "3.8"
services:
  server:
    image: server_go
    ports:
      - "3000:3000"
    environment:
      HOST_NAME: postgres
      USER_NAME: postgres
      PASSWORD: postgres
      DB_NAME: postgres
      POST_PORT: 5432
      GOSERVER_PORT: :3000
      MEMC_PATH: memecached:11211
    depends_on:
      - postgres
      - memecached
  postgres:
    image: postgres
    ports:
      - "5432:5432"
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
      POSTGRES_DB: postgres
  memecached:
    image: memcached
    ports: 
      - "11211:11211"
```

Для создания и запуска контейнера с указанными выше настройками используется команда ` docker compose up`.

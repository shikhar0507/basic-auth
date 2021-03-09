package main

import (
       "bufio"
       "context"
       "fmt"
       "github.com/jackc/pgx/v4/pgxpool"
       "log"
       "net/url"
       "os"
       "strings"
)

func main() {
       file,err := os.Open("data.txt")
       if err != nil {
              log.Fatal(err)
       }
       db,err := pgxpool.Connect(context.Background(),os.Getenv("DATABASE_URL"))

       scanner := bufio.NewScanner(file)
       for scanner.Scan() {
              line := scanner.Text()
              splitFirst := strings.Split(line,"(")
              splitLast := strings.Split(splitFirst[1],")")
              parsedUrl, _ := url.Parse(splitLast[0])
              fmt.Println(parsedUrl.Host,parsedUrl.String())
              _, err = db.Exec(context.Background(),"insert into domains values($1,$2)",parsedUrl.Host,parsedUrl.String())
              if err != nil {
                     log.Fatal(err)
              }
       }
}
package main

import (
	"log"
	"os"
	"fmt"
	"time"
	"regexp"
	"strconv"
	"strings"

	T "gopkg.in/telebot.v3"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

func fixa(db *sql.DB, name string, user string, km int) error {

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare("insert into achievements(name, user, ts, km) values(?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(name, user, time.Now().Unix(), km)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func getres(db *sql.DB, user string) (int, error) {

	stmt, err := db.Prepare("select sum(km) as res from achievements where user = ?")
	if err != nil {
		return -1, err
	}
	defer stmt.Close()
	var res int
	err = stmt.QueryRow(user).Scan(&res)
	return res, err
}

func main() {

	kmre := regexp.MustCompile(`[-]?\d[\d,]*[\.]?[\d{2}]*`)

	db, err := sql.Open("sqlite3", "./rur.sqlite")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	pref := T.Settings{
		Token:  os.Getenv("TOKEN"),
		Poller: &T.LongPoller{Timeout: 10 * time.Second},
	}

	b, err := T.NewBot(pref)
	if err != nil {
		log.Fatal(err)
		return
	}

	b.Handle(T.OnPhoto, func(c T.Context) error {
		
		msg := c.Message()
		user := msg.Sender.Username
		kms := kmre.FindString(msg.Caption)
		name := strings.ReplaceAll(msg.Caption, kms, "")
		km, err := strconv.Atoi(kms)
		if err != nil {
			c.Reply("Сколько сколько километров?")
			return nil
		}

		err = fixa(db, name, user, km)
		if err != nil {
			c.Reply(err)
			return err
		}

		res, err := getres(db, user)
		if err != nil {
			c.Reply(err)
			return err
		}
		
		return c.Reply(fmt.Sprintf("%s пробежал %d!", name, res))
	})

	b.Start()
}

package main

import (
	"log"

	"github.com/blendlabs/chatbus/server"
	"github.com/blendlabs/chatbus/server/db"
	"github.com/blendlabs/spiffy"
)

func configInit() error {
	return server.DefaultConfig().FromEnvironment()
}

func connectDB() error {
	spiffy.CreateDbAlias("main", spiffy.NewDbConnectionFromEnvironment())
	spiffy.SetDefaultAlias("main")

	_, err := spiffy.DefaultDb().Open()
	if err != nil {
		return err
	}

	spiffy.DefaultDb().Connection.SetMaxIdleConns(50)
	return nil
}

func main() {
	err := configInit()
	if err != nil {
		log.Fatal(err)
	}

	err = connectDB()
	if err != nil {
		log.Fatal(err)
	}

	err = db.Migrate()
	if err != nil {
		log.Fatal(err)
	}

	app, err := server.New()
	if err != nil {
		log.Fatal(err)
	}

	log.Fatal(app.Start())
}

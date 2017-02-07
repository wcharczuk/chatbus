package model

import (
	"log"
	"os"
	"testing"

	"github.com/blendlabs/spiffy"
)

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

func TestMain(m *testing.M) {
	err := connectDB()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	os.Exit(m.Run())
}

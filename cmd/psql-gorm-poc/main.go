package main

import (
	"time"

	"github.com/mattn/go-colorable"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gitlab.com/blender/flamenco-ng-poc/internal/appinfo"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	output := zerolog.ConsoleWriter{Out: colorable.NewColorableStdout(), TimeFormat: time.RFC3339}
	log.Logger = log.Output(output)
	log.Info().Str("version", appinfo.ApplicationVersion).Msgf("starting %v GORM test", appinfo.ApplicationName)

	dsn := "host=localhost user=flamenco password=flamenco dbname=flamenco TimeZone=Europe/Amsterdam"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Panic().Err(err).Msg("failed to connect database")
	}

	// Migrate the schema
	if err := db.AutoMigrate(&Dude{}); err != nil {
		log.Panic().Err(err).Msg("failed to automigrate database")
	}

	var dude Dude

	db.Transaction(func(tx *gorm.DB) error {
		// Find pre-existing
		findResult := tx.First(&dude, "metadata ->> 'name' = ?", "Daš D°°D")
		switch findResult.Error {
		case gorm.ErrRecordNotFound:
			// Create if not found
			dude = Dude{
				Name:     "Dude",
				Email:    "the@dude.nl",
				Metadata: make(map[string]interface{}),
			}
			dude.Metadata["name"] = "Daš D°°D"
			dude.Metadata["integer"] = 47
			dude.Metadata["float"] = 47.327

			log.Info().Interface("data", dude).Msg("the data pre-insert")

			createResult := tx.Create(&dude)
			if createResult.Error != nil {
				log.Error().Err(createResult.Error).Msg("failed to insert dude")
				return createResult.Error
			}

			log.Info().Interface("data", dude).Msg("the data post-insert")
		case nil:
			log.Info().Interface("dude", dude).Msg("the found dude")
		default:
			log.Error().Err(findResult.Error).Msg("failed to fetch dude")
			return findResult.Error
		}

		// Update
		var theInt int
		switch v := dude.Metadata["integer"].(type) {
		case float64:
			theInt = int(v)
		case int:
			theInt = int(v)
		default:
			log.Panic().Interface("value", v).Msg("unexpected type in JSON")
		}

		dude.Metadata["integer"] = theInt + 1

		tx.Model(&dude).Update("metadata", dude.Metadata)
		if tx.Error != nil {
			log.Panic().Err(tx.Error).Msg("failed to update dude")
		}

		return nil
	})

	// Fetch again
	var newDude Dude
	tx := db.First(&newDude, dude.ID)
	if tx.Error != nil {
		log.Panic().Err(tx.Error).Msg("failed to re-fetch dude")
	}
	log.Info().Interface("newDude", newDude).Msg("the updated dude")
}

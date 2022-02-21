package persistence

/* ***** BEGIN GPL LICENSE BLOCK *****
 *
 * Original Code Copyright (C) 2022 Blender Foundation.
 *
 * This file is part of Flamenco.
 *
 * Flamenco is free software: you can redistribute it and/or modify it under
 * the terms of the GNU General Public License as published by the Free Software
 * Foundation, either version 3 of the License, or (at your option) any later
 * version.
 *
 * Flamenco is distributed in the hope that it will be useful, but WITHOUT ANY
 * WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR
 * A PARTICULAR PURPOSE.  See the GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License along with
 * Flamenco.  If not, see <https://www.gnu.org/licenses/>.
 *
 * ***** END GPL LICENSE BLOCK ***** */

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"runtime"

	"github.com/rs/zerolog/log"
	"golang.org/x/term"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var errInputTooLong = errors.New("input is too long")

const adminDSN = "host=localhost user=postgres password=%s dbname=%s TimeZone=Europe/Amsterdam"

// InitialSetup uses the `postgres` admin user to set up the database.
// TODO: distinguish between production and development setups.
func InitialSetup() error {
	// Get the password of the 'postgres' user.
	adminPass, err := readPassword()
	if err != nil {
		return fmt.Errorf("unable to read password: %w", err)
	}

	// Connect to the 'postgres' database so we can create other databases.
	db, err := connectDBAsAdmin(adminPass, "postgres")
	if err != nil {
		return fmt.Errorf("unable to connect to the database: %w", err)
	}

	// TODO: get username / password / database name from some config file, user input, CLI args, whatevah.
	// Has to be used by the regular Flamenco Manager runs as well, though.
	username := "flamenco"
	userPass := "flamenco"
	// tx := db.Exec("CREATE USER flamenco PASSWORD ? NOSUPERUSER NOCREATEDB NOCREATEROLE INHERIT LOGIN", userPass)
	// if tx.Error != nil {
	// 	return fmt.Errorf("unable to create database user '%s': %w", username, tx.Error)
	// }
	{
		sqlDB, err := db.DB()
		if err != nil {
			panic(err)
		}
		_, err = sqlDB.Exec("CREATE USER flamenco WITH PASSWORD $1::string NOSUPERUSER NOCREATEDB NOCREATEROLE INHERIT LOGIN", userPass)
		if err != nil {
			panic(err)
		}
	}

	// Create the databases.
	tx := db.Debug().Exec("CREATE DATABASE flamenco OWNER ? ENCODING 'utf8'", username)
	if tx.Error != nil {
		return fmt.Errorf("unable to create database 'flamenco': %w", tx.Error)
	}
	tx = db.Exec("CREATE DATABASE flamenco-test OWNER ? ENCODING 'utf8'", username)
	if tx.Error != nil {
		return fmt.Errorf("unable to create database 'flamenco': %w", tx.Error)
	}

	// Close the connection so we can reconnect.
	sqlDB, err := db.DB()
	if err != nil {
		fmt.Printf("error closing the database connection, please report this issue: %v", err)
	} else {
		sqlDB.Close()
	}

	// Allow 'flamenco' user to completely nuke and recreate the flamenco-test database, without needing 'CREATEDB' permission.
	db, err = connectDBAsAdmin(adminPass, "flamenco-test")
	if err != nil {
		return fmt.Errorf("unable to reconnect to the database: %w", err)
	}
	tx = db.Exec("ALTER SCHEMA public OWNER TO ?", username)
	if tx.Error != nil {
		fmt.Printf("Unable to allow database user '%s' to reset the test database: %v\n", username, tx.Error)
		fmt.Println("This is not an issue, unless you want to develop Flamenco yourself.")
	}

	return nil
}

func readPassword() (string, error) {
	if pwFromEnv := os.Getenv("PSQL_ADMIN"); pwFromEnv != "" {
		log.Info().Msg("getting password from PSQL_ADMIN environment variable")
		return pwFromEnv, nil
	}

	fmt.Print("PostgreSQL admin password: ")

	var (
		line []byte
		err  error
	)

	if runtime.GOOS == "windows" {
		// term.ReadPassword() doesn't work reliably on Windows, especially when you
		// use a MingW terminal (like Git Bash). See
		// https://github.com/golang/go/issues/11914#issuecomment-613715787 for more
		// info.
		//
		// The downside is that this echoes the password to the terminal.
		buf := bufio.NewReader(os.Stdin)
		line, _, err = buf.ReadLine()
	} else {
		fd := int(os.Stdin.Fd())
		line, err = term.ReadPassword(fd)
	}
	if err != nil {
		return "", err
	}
	return string(line), nil
}

func connectDBAsAdmin(password, database string) (*gorm.DB, error) {
	dsn := fmt.Sprintf(adminDSN, password, database)
	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}

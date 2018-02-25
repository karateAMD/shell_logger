package main

import (
	"fmt"
	"encoding/json"

	"github.com/boltdb/bolt"
)

var db *bolt.DB

// PLEASE DELETE
func GetFirstCommand(str string) string {
	for i := 0; i < len(str); i++ {
		if str[i] == ' ' {
			return str[:i]
		}
	}
	return str
}

// TODO: needs to be called anytime shell logger starts running
// Creates the database if it doesn't exist, otherwise opens the database and creates a bucket for key:value pairs
func SetupDatabase() error {
	pathToDatabase := "my.db"
	var err error = nil
	db, err = bolt.Open(pathToDatabase, 0600, nil)
	if err != nil {
		fmt.Println("Failed to open database: %v", err)
		return err
	}
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("Mappings"))
		if err != nil {
			fmt.Println("Failed to create bucket: %v", err)
			return err
		}
		return nil
	})
	return err
}

func CloseDatabase() {
	db.Close()
}

// Gets the JSON Object from the database
func GetJsonValues(key []byte) ([]byte, error) {
	var jsonValue []byte
	err := db.View(func(tx *bolt.Tx) error {
		jsonValue = tx.Bucket([]byte("Mappings")).Get(key)
		return nil
	})
	if err != nil {
		fmt.Errorf("Got json retrieval error: %v", err)
	}
	return jsonValue, err
}

// Gets the list of bad commands from the database
func GetGoodCommands(key []byte) ([][]byte, error) {
	var vals [][]byte
	jsonObject, err := GetJsonValues(key)
	if err != nil {
		fmt.Println("Failed to retrieve values: ", err)
		return nil, err
	}
	json.Unmarshal(jsonObject, &vals)
	if err != nil {
		fmt.Println("Failed to unmarshal: ", err)
	}
	return vals, err
}

// Inserts the key:value pair of correctCommand:incorrectCommand into the database
func Insert(correct []byte, incorrect []byte) error {
	firstWord := []byte(GetFirstCommand(string(correct)))
	correctCommands, err := GetGoodCommands(firstWord)
	if err != nil {
		return err
	}
	if correctCommands == nil { 
		err := db.Update(func(tx *bolt.Tx) error {
			correctCommand := [1][]byte{correct}
			jsonObject, err := json.Marshal(correctCommand)
			if err != nil {
				fmt.Println("Failed to marshal: %v", err)
				return err
			}
			err = tx.Bucket([]byte("Mappings")).Put(firstWord, jsonObject)
			if err != nil {
				fmt.Println("Failed to insert values: %v", err)
				return err
			}
			return nil
			})
		return err
	} else {
		err := db.Update(func(tx *bolt.Tx) error {
			jsonObject, err := GetJsonValues(firstWord)
			if err != nil {
				return err
			}
			var vals [][]byte
			err = json.Unmarshal(jsonObject, &vals)
			if err != nil {
				fmt.Println("Failed to unmarshal object: %v", err)
				return err
			}
			vals = append(vals, correct)
			newJsonObject, err := json.Marshal(vals)
			if err != nil {
				fmt.Println("Failed to marshal object: %v", err)
				return err
			}
			err = tx.Bucket([]byte("Mappings")).Put(firstWord, newJsonObject)
			if err != nil {
				return fmt.Errorf("Could not set value: %v", err)
				return err
			}
			return nil
			})
		return err
	}
	
}

// DELETE
func main() {
	SetupDatabase()
	corr := []byte("git push origin master")
	incorr := []byte("git push origin mast")
	incorr2 := []byte("git pus origin master")
	secondCorr := []byte("fc -ln -l")
	secondIncorr := []byte("fd -ln -l")
	newCorr := []byte("git commit -m")
	newIncorr := []byte("git comit -m")
	Insert(corr, incorr)
	Insert(secondCorr, secondIncorr)
	Insert(corr, incorr2)
	Insert(newCorr, newIncorr)
	str1, _ := GetGoodCommands([]byte("git"))
	str2, _ := GetGoodCommands([]byte("fc"))
	for i := 0; i < len(str1); i++ {
		fmt.Println(string(str1[i]))
	}
	for i := 0; i < len(str2); i++ {
		fmt.Println(string(str2[i]))
	}
}
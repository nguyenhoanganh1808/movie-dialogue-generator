// db/db.go
package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func InitDB() {
	// Get connection info from environment
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	// Build connection string
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	// Connect to database
	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Test connection
	err = DB.Ping()
	if err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	log.Println("Connected to database")
}

// Character models and operations
type Character struct {
	ID     int      `json:"id"`
	Name   string   `json:"name"`
	Type   string   `json:"type"`
	Traits []string `json:"traits"`
}

func GetCharacters() ([]Character, error) {
	rows, err := DB.Query("SELECT id, name, type, traits FROM characters")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var characters []Character
	for rows.Next() {
		var c Character
		var traitsJSON []byte
		if err := rows.Scan(&c.ID, &c.Name, &c.Type, &traitsJSON); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(traitsJSON, &c.Traits); err != nil {
			return nil, err
		}
		characters = append(characters, c)
	}

	return characters, nil
}

func CreateCharacter(character Character) (int, error) {
	traitsJSON, err := json.Marshal(character.Traits)
	if err != nil {
		return 0, err
	}

	var id int
	err = DB.QueryRow(
		"INSERT INTO characters (name, type, traits) VALUES ($1, $2, $3) RETURNING id",
		character.Name, character.Type, traitsJSON,
	).Scan(&id)

	return id, err
}

// Reference Dialogue models and operations
type ReferenceDialogue struct {
	ID         int      `json:"id"`
	Source     string   `json:"source"`
	Characters []string `json:"characters"`
	Content    string   `json:"content"`
	Tags       []string `json:"tags"`
}

func GetReferenceDialogues(tag string) ([]ReferenceDialogue, error) {
	var rows *sql.Rows
	var err error

	if tag != "" {
		rows, err = DB.Query("SELECT id, source, characters, content, tags FROM reference_dialogues WHERE $1 = ANY(tags)", tag)
	} else {
		rows, err = DB.Query("SELECT id, source, characters, content, tags FROM reference_dialogues")
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dialogues []ReferenceDialogue
	for rows.Next() {
		var d ReferenceDialogue
		var charactersJSON []byte
		if err := rows.Scan(&d.ID, &d.Source, &charactersJSON, &d.Content, &d.Tags); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(charactersJSON, &d.Characters); err != nil {
			return nil, err
		}
		dialogues = append(dialogues, d)
	}

	return dialogues, nil
}

func AddReferenceDialogue(dialogue ReferenceDialogue) (int, error) {
	charactersJSON, err := json.Marshal(dialogue.Characters)
	if err != nil {
		return 0, err
	}

	var id int
	err = DB.QueryRow(
		"INSERT INTO reference_dialogues (source, characters, content, tags) VALUES ($1, $2, $3, $4) RETURNING id",
		dialogue.Source, charactersJSON, dialogue.Content, dialogue.Tags,
	).Scan(&id)

	return id, err
}

// Generated Dialogue models and operations
type GeneratedDialogue struct {
	ID        int             `json:"id"`
	Scenario  string          `json:"scenario"`
	Characters json.RawMessage `json:"characters"`
	Content   json.RawMessage `json:"content"`
	CreatedAt string          `json:"created_at"`
}

func SaveGeneratedDialogue(scenario string, characters json.RawMessage, content json.RawMessage) (int, error) {
	var id int
	err := DB.QueryRow(
		"INSERT INTO dialogues (scenario, characters, content) VALUES ($1, $2, $3) RETURNING id",
		scenario, characters, content,
	).Scan(&id)

	return id, err
}

func GetGeneratedDialogues() ([]GeneratedDialogue, error) {
	rows, err := DB.Query("SELECT id, scenario, characters, content, created_at FROM dialogues ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dialogues []GeneratedDialogue
	for rows.Next() {
		var d GeneratedDialogue
		if err := rows.Scan(&d.ID, &d.Scenario, &d.Characters, &d.Content, &d.CreatedAt); err != nil {
			return nil, err
		}
		dialogues = append(dialogues, d)
	}

	return dialogues, nil
}
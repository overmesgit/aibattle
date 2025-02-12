package main

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
)

type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type Action struct {
	Move    *MoveAction   `json:"move,omitempty"`
	Hold    *struct{}     `json:"hold,omitempty"`
	Attack1 *AttackAction `json:"attack_1,omitempty"`
	Skill1  *SkillAction  `json:"skill_1,omitempty"`
}

type MoveAction struct {
	Distance int `json:"distance"`
}

type AttackAction struct {
	Range  int `json:"range"`
	Damage int `json:"damage"`
}

type SkillAction struct {
	Effect string `json:"effect"`
	Range  int    `json:"range"`
	Value  int    `json:"value"`
}

type Unit struct {
	ID         int      `json:"id"`
	Team       int      `json:"team"`
	Type       string   `json:"type"`
	Initiative int      `json:"initiative"`
	HP         int      `json:"hp"`
	MaxHP      int      `json:"maxHp"`
	Position   Position `json:"position"`
	Actions    Action   `json:"actions"`
}

type GameState struct {
	Turn   int    `json:"turn"`
	Units  []Unit `json:"units"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type ActionTarget struct {
	Action string   `json:"action"`
	Target Position `json:"target"`
}

type ActionResponse struct {
	UnitAction []ActionTarget `json:"unit_action"`
}

// InputJSON format:
type NextTurnInput struct {
	State  GameState `json:"state"`
	UnitID int       `json:"unit_id"`
}

func distance(a, b Position) float64 {
	dx := float64(a.X - b.X)
	dy := float64(a.Y - b.Y)
	return math.Sqrt(dx*dx + dy*dy)
}

func PlayTurn(inputJSON []byte) ([]byte, error) {
	var input NextTurnInput
	if err := json.Unmarshal(inputJSON, &input); err != nil {
		return nil, err
	}
	response := GetTurnActions(input)
	fmt.Printf("input: %+v\n", input)
	fmt.Printf("response: %+v\n", response)
	return json.Marshal(response)
}

// implementation of GetTurnActions(input NextTurnInput) ActionResponse
<generated>

func HandleTurn(w http.ResponseWriter, r *http.Request) {
	var request NextTurnInput
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	inputJSON, err := json.Marshal(request)
	if err != nil {
		http.Error(w, "Error marshaling game state", http.StatusInternalServerError)
		return
	}

	response, err := PlayTurn(inputJSON)
	if err != nil {
		http.Error(w, "Error processing turn", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}

func main() {
	fmt.Println("starting game server")
	http.HandleFunc("/", HandleTurn)
	http.ListenAndServe(":8080", nil)
}

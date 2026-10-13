package services

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"strings"
	"sync/atomic"
	"time"

	paho "github.com/eclipse/paho.mqtt.golang"
)

type MessageService struct {
	database *sql.DB
}

var messageTraceID uint64

type cardRequest struct {
	CardID string `json:"card_id"`
}

type moneyRequest struct {
	CardID string `json:"card_id"`
	Amount int64  `json:"amount"`
}

type memberRequest struct {
	CardID   string `json:"card_id"`
	Username string `json:"username"`
}

func NewMessageService(database *sql.DB) *MessageService {
	return &MessageService{database: database}
}

func (service *MessageService) HandleMessage(client paho.Client, message paho.Message) {
	traceID := atomic.AddUint64(&messageTraceID, 1)
	startedAt := time.Now()
	log.Printf("[trace=%d] received topic=%s payload=%s", traceID, message.Topic(), message.Payload())

	var response interface{}
	var err error

	switch message.Topic() {
	case "/check_role":
		response, err = service.checkRole(message.Payload())
	case "/pay":
		response, err = service.changeBalance(message.Payload(), false)
	case "/retieve_member":
		response, err = service.retrieveMember(message.Payload())
	case "/retieve_leaderboard":
		response, err = service.retrieveLeaderboard()
	case "/add_money":
		response, err = service.changeBalance(message.Payload(), true)
	case "/create_guest":
		response, err = service.createGuest(message.Payload())
	case "/create_member":
		response, err = service.createMember(message.Payload())
	case "/remove_card":
		response, err = service.removeCard(message.Payload())
	default:
		return
	}

	if err != nil {
		log.Printf("[trace=%d] action failed topic=%s error=%v", traceID, message.Topic(), err)
		response = map[string]string{"error": err.Error()}
	} else {
		log.Printf("[trace=%d] action completed topic=%s", traceID, message.Topic())
	}
	payload, err := json.Marshal(response)
	if err != nil {
		log.Printf("[trace=%d] response marshal failed topic=%s error=%v", traceID, message.Topic(), err)
		return
	}

	token := client.Publish(message.Topic()+"/response", 1, false, payload)
	token.Wait()
	if err := token.Error(); err != nil {
		log.Printf("[trace=%d] response publish failed topic=%s error=%v", traceID, message.Topic()+"/response", err)
		return
	}
	log.Printf("[trace=%d] response sent topic=%s payload=%s duration=%s", traceID, message.Topic()+"/response", payload, time.Since(startedAt))
}

func (service *MessageService) checkRole(payload []byte) (map[string]string, error) {
	request, err := parseCardRequest(payload)
	if err != nil {
		return nil, err
	}

	var exists bool
	if err := service.database.QueryRow("SELECT EXISTS (SELECT 1 FROM guests WHERE card_id = $1)", request.CardID).Scan(&exists); err != nil {
		return nil, err
	}
	if exists {
		return map[string]string{"card_id": request.CardID, "role": "guest"}, nil
	}
	if err := service.database.QueryRow("SELECT EXISTS (SELECT 1 FROM members WHERE card_id = $1)", request.CardID).Scan(&exists); err != nil {
		return nil, err
	}
	if exists {
		return map[string]string{"card_id": request.CardID, "role": "member"}, nil
	}

	return nil, errors.New("invalid card")
}

func (service *MessageService) changeBalance(payload []byte, addMoney bool) (map[string]interface{}, error) {
	request, err := parseMoneyRequest(payload)
	if err != nil {
		return nil, err
	}
	if request.Amount <= 0 {
		return nil, errors.New("amount must be greater than 0")
	}

	delta := request.Amount
	if !addMoney {
		delta = -request.Amount
	}
	result, err := service.database.Exec("UPDATE guests SET balance = balance + $1 WHERE card_id = $2 AND balance + $1 >= 0", delta, request.CardID)
	if err != nil {
		return nil, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	if affected == 0 {
		result, err = service.database.Exec("UPDATE members SET balance = balance + $1 WHERE card_id = $2 AND balance + $1 >= 0", delta, request.CardID)
		if err != nil {
			return nil, err
		}
		affected, err = result.RowsAffected()
		if err != nil {
			return nil, err
		}
	}
	if affected == 0 {
		if addMoney {
			return nil, errors.New("invalid card")
		}
		return nil, errors.New("invalid card or insufficient balance")
	}

	return map[string]interface{}{"card_id": request.CardID, "amount": request.Amount, "operation": operationName(addMoney)}, nil
}

func (service *MessageService) retrieveMember(payload []byte) (interface{}, error) {
	request, err := parseCardRequest(payload)
	if err != nil {
		return nil, err
	}

	var member struct {
		CardID   string `json:"card_id"`
		Username string `json:"username"`
		Balance  int64  `json:"balance"`
		Points   int64  `json:"points"`
	}
	if err := service.database.QueryRow("SELECT card_id, username, balance, points FROM members WHERE card_id = $1", request.CardID).Scan(&member.CardID, &member.Username, &member.Balance, &member.Points); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("member not found")
		}
		return nil, err
	}
	return member, nil
}

func (service *MessageService) createGuest(payload []byte) (map[string]string, error) {
	request, err := parseCardRequest(payload)
	if err != nil {
		return nil, err
	}

	var exists bool
	if err := service.database.QueryRow("SELECT EXISTS (SELECT 1 FROM guests WHERE card_id = $1) OR EXISTS (SELECT 1 FROM members WHERE card_id = $1)", request.CardID).Scan(&exists); err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("card already exists")
	}
	if _, err := service.database.Exec("INSERT INTO guests (card_id) VALUES ($1)", request.CardID); err != nil {
		return nil, err
	}
	return map[string]string{"card_id": request.CardID, "role": "guest", "status": "created"}, nil
}

func (service *MessageService) createMember(payload []byte) (map[string]string, error) {
	var request memberRequest
	if err := json.Unmarshal(payload, &request); err != nil || strings.TrimSpace(request.CardID) == "" || strings.TrimSpace(request.Username) == "" {
		return nil, errors.New("card_id and username are required")
	}

	var exists bool
	if err := service.database.QueryRow("SELECT EXISTS (SELECT 1 FROM guests WHERE card_id = $1) OR EXISTS (SELECT 1 FROM members WHERE card_id = $1)", request.CardID).Scan(&exists); err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("card already exists")
	}
	if err := service.database.QueryRow("SELECT EXISTS (SELECT 1 FROM members WHERE username = $1)", request.Username).Scan(&exists); err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("username already exists")
	}
	if _, err := service.database.Exec("INSERT INTO members (card_id, username) VALUES ($1, $2)", request.CardID, request.Username); err != nil {
		return nil, err
	}
	return map[string]string{"card_id": request.CardID, "role": "member", "status": "created"}, nil
}

func (service *MessageService) removeCard(payload []byte) (map[string]string, error) {
	request, err := parseCardRequest(payload)
	if err != nil {
		return nil, err
	}

	transaction, err := service.database.Begin()
	if err != nil {
		return nil, err
	}
	defer transaction.Rollback()

	guestResult, err := transaction.Exec("DELETE FROM guests WHERE card_id = $1", request.CardID)
	if err != nil {
		return nil, err
	}
	memberResult, err := transaction.Exec("DELETE FROM members WHERE card_id = $1", request.CardID)
	if err != nil {
		return nil, err
	}
	guestCount, err := guestResult.RowsAffected()
	if err != nil {
		return nil, err
	}
	memberCount, err := memberResult.RowsAffected()
	if err != nil {
		return nil, err
	}
	if guestCount+memberCount == 0 {
		return nil, errors.New("invalid card")
	}
	if err := transaction.Commit(); err != nil {
		return nil, err
	}
	return map[string]string{"card_id": request.CardID, "status": "removed"}, nil
}

func (service *MessageService) retrieveLeaderboard() ([]map[string]interface{}, error) {
	rows, err := service.database.Query("SELECT card_id, username, points FROM members ORDER BY points DESC, card_id ASC LIMIT 10")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	leaderboard := make([]map[string]interface{}, 0, 10)
	for rows.Next() {
		var cardID, username string
		var points int64
		if err := rows.Scan(&cardID, &username, &points); err != nil {
			return nil, err
		}
		leaderboard = append(leaderboard, map[string]interface{}{"card_id": cardID, "username": username, "points": points})
	}
	return leaderboard, rows.Err()
}

func parseCardRequest(payload []byte) (cardRequest, error) {
	var request cardRequest
	if err := json.Unmarshal(payload, &request); err != nil || strings.TrimSpace(request.CardID) == "" {
		return cardRequest{}, errors.New("card_id is required")
	}
	return request, nil
}

func parseMoneyRequest(payload []byte) (moneyRequest, error) {
	var request moneyRequest
	if err := json.Unmarshal(payload, &request); err != nil || strings.TrimSpace(request.CardID) == "" {
		return moneyRequest{}, errors.New("card_id and amount are required")
	}
	return request, nil
}

func operationName(addMoney bool) string {
	if addMoney {
		return "add_money"
	}
	return "pay"
}

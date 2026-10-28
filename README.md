# MAD Tetris Backend

MQTT backend for an RFID-card Tetris arcade. Players tap a card; an MQTT broker
(HiveMQ) relays requests; this Go service handles them against PostgreSQL and
publishes a response on `<topic>/response`.

## Concepts

- **Guest** — card with a `balance` only (pay to play).
- **Member** — card with `balance`, a unique `username`, and a high-score `points`.
- Every request is JSON. Every reply is published to the same topic suffixed with
  `/response`.

## MQTT Topics

| Topic                  | Payload                                 | Does                                  |
|------------------------|-----------------------------------------|---------------------------------------|
| `/check_role`          | `{"card_id":"..."}`                      | Returns `guest` / `member` / error    |
| `/pay`                 | `{"card_id":"...","amount":10}`          | Subtract balance                      |
| `/add_money`           | `{"card_id":"...","amount":50}`          | Add balance                           |
| `/add_score`           | `{"card_id":"...","points":1200}`        | Save high score (member, best-only)   |
| `/retieve_member`      | `{"card_id":"..."}`                      | Member profile                        |
| `/retieve_leaderboard` | `{}`                                     | Top 10 members by points              |
| `/create_guest`        | `{"card_id":"..."}`                      | New guest                             |
| `/create_member`       | `{"card_id":"...","username":"..."}`     | New member                            |
| `/remove_card`         | `{"card_id":"..."}`                      | Delete guest or member                |
| `/check_balance`       | `{"card_id":"..."}`                      | Returns current balance + role        |

## Stack

Go 1.25 · PostgreSQL 16 · Eclipse Paho MQTT · HiveMQ broker (TLS).

## Setup

```bash
cp .env.example .env            # fill in MQTT + Postgres values
docker compose up -d            # start PostgreSQL
# apply schema
docker compose exec -T postgres psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" \
  < db/migrations/001_create_card_tables.up.sql
go run .                        # start the backend
```

### Env (`.env`)

```
MQTT_BROKER=tls://your-hivemq-cluster:8883
MQTT_CLIENT_ID=your-client-id
MQTT_USERNAME=your-username
MQTT_PASSWORD=your-password
POSTGRES_DB=tetris
POSTGRES_USER=tetris
POSTGRES_PASSWORD=change-me
POSTGRES_PORT=5432
POSTGRES_HOST=localhost
```

## Mock a Member

### A. Direct SQL seed (balance + points preset)

```bash
docker compose exec postgres psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c \
  "INSERT INTO members (card_id, username, balance, points) \
   VALUES ('MOCK001', 'mockuser', 1000, 5000);"
```

### B. Via MQTT (goes through the real handler; balance/points start at 0)

```bash
mosquitto_pub -h your-hivemq-cluster -p 8883 \
  -u "$MQTT_USERNAME" -P "$MQTT_PASSWORD" --cafile /etc/ssl/cert.pem \
  -t /create_member \
  -m '{"card_id":"MOCK001","username":"mockuser"}'
```

Then top up / score the mock card:

```bash
mosquitto_pub ... -t /add_money -m '{"card_id":"MOCK001","amount":1000}'
mosquitto_pub ... -t /add_score -m '{"card_id":"MOCK001","points":5000}'
```

Verify:

```bash
docker compose exec postgres psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" \
  -c "SELECT * FROM members;"
```

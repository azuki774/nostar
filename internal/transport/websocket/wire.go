package websocket

import (
	"encoding/json"
	"fmt"
)

type WireMessage struct {
	Type           string
	SubscriptionID string
	Event          json.RawMessage
	Filters        []json.RawMessage
}

func (w *WireMessage) UnmarshalJSON(data []byte) error {
	// まず配列として受ける
	var arr []json.RawMessage
	if err := json.Unmarshal(data, &arr); err != nil {
		return err
	}
	if len(arr) == 0 {
		return fmt.Errorf("empty wire message: %s", string(data))
	}

	// 0 番目: "EVENT" / "REQ" / "CLOSE"
	if err := json.Unmarshal(arr[0], &w.Type); err != nil {
		return fmt.Errorf("invalid type: %w", err)
	}

	switch w.Type {
	case "EVENT":
		// ["EVENT", <event>]
		if len(arr) != 2 {
			return fmt.Errorf("invalid EVENT message: %s", string(data))
		}
		w.Event = arr[1]

	case "REQ":
		// ["REQ", <subscription_id>, <filter>, <filter>...]
		if len(arr) < 3 {
			return fmt.Errorf("invalid REQ message: %s", string(data))
		}
		if err := json.Unmarshal(arr[1], &w.SubscriptionID); err != nil {
			return fmt.Errorf("invalid REQ subscription id: %w", err)
		}
		w.Filters = arr[2:]

	case "CLOSE":
		// ["CLOSE", <subscription_id>]
		if len(arr) != 2 {
			return fmt.Errorf("invalid CLOSE message: %s", string(data))
		}
		if err := json.Unmarshal(arr[1], &w.SubscriptionID); err != nil {
			return fmt.Errorf("invalid CLOSE subscription id: %w", err)
		}

	default:
		return fmt.Errorf("unknown wire message type: %q", w.Type)
	}

	return nil
}

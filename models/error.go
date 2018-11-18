package models

import (
	"encoding/json"
)

// Error error
// swagger:model Error
type Error struct {

	// Текстовое описание ошибки.
	// В процессе проверки API никаких проверок на содерижимое данного описание не делается.
	//
	// Read Only: true
	Message string `json:"message,omitempty"`
}

// MarshalBinary interface implementation
func (m *Error) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}

	return json.Marshal(m)
}

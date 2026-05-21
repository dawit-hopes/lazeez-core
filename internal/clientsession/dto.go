package clientsession

import "time"

type CreateSessionInput struct {
	Reference string `json:"reference"`
}

type ClientSessionResponse struct {
	SessionKey     string    `json:"session_key"`
	TableReference string    `json:"table_reference"`
	TableID        string    `json:"table_id"`
	TableName      string    `json:"table_name"`
	BranchID       string    `json:"branch_id"`
	ExpiresAt      time.Time `json:"expires_at"`
}

func (s *ClientSession) ToResponse(tableName string) *ClientSessionResponse {
	return &ClientSessionResponse{
		SessionKey:     s.SessionKey,
		TableReference: s.TableReference,
		TableID:        s.TableID,
		TableName:      tableName,
		BranchID:       s.BranchID,
		ExpiresAt:      s.ExpiresAt,
	}
}

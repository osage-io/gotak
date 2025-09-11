package chat

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dfedick/gotak/pkg/logger"
)

func TestNewService_Success(t *testing.T) {
	mockDB, _, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	sqlxDB := sqlx.NewDb(mockDB, "postgres")
	log := logger.NewDefault()

	service := NewService(sqlxDB, log)
	assert.NotNil(t, service)
	assert.Equal(t, sqlxDB, service.db)
	assert.Equal(t, log, service.logger)
}

// Test enum validation functions
func TestChatRoomType_Validate(t *testing.T) {
	tests := []struct {
		roomType ChatRoomType
		valid    bool
	}{
		{RoomTypeGroup, true},
		{RoomTypePrivate, true},
		{RoomTypeTactical, true},
		{RoomTypeEmergency, true},
		{ChatRoomType("invalid"), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.roomType), func(t *testing.T) {
			err := tt.roomType.Validate()
			if tt.valid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestClassification_Validate(t *testing.T) {
	tests := []struct {
		classification Classification
		valid          bool
	}{
		{ClassificationUnclassified, true},
		{ClassificationRestricted, true},
		{ClassificationConfidential, true},
		{ClassificationSecret, true},
		{ClassificationTopSecret, true},
		{Classification("invalid"), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.classification), func(t *testing.T) {
			err := tt.classification.Validate()
			if tt.valid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestMessageType_Validate(t *testing.T) {
	tests := []struct {
		messageType MessageType
		valid       bool
	}{
		{MessageTypeText, true},
		{MessageTypeSystem, true},
		{MessageTypePosition, true},
		{MessageTypeEmergency, true},
		{MessageTypeTacticalReport, true},
		{MessageType("invalid"), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.messageType), func(t *testing.T) {
			err := tt.messageType.Validate()
			if tt.valid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestMessagePriority_Validate(t *testing.T) {
	tests := []struct {
		priority MessagePriority
		valid    bool
	}{
		{PriorityLow, true},
		{PriorityNormal, true},
		{PriorityHigh, true},
		{PriorityUrgent, true},
		{PriorityEmergency, true},
		{MessagePriority("invalid"), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.priority), func(t *testing.T) {
			err := tt.priority.Validate()
			if tt.valid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestParticipantRole_Validate(t *testing.T) {
	tests := []struct {
		role  ParticipantRole
		valid bool
	}{
		{RoleAdmin, true},
		{RoleModerator, true},
		{RoleMember, true},
		{RoleObserver, true},
		{ParticipantRole("invalid"), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.role), func(t *testing.T) {
			err := tt.role.Validate()
			if tt.valid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestReactionType_Validate(t *testing.T) {
	tests := []struct {
		reaction ReactionType
		valid    bool
	}{
		{ReactionRoger, true},
		{ReactionWilco, true},
		{ReactionNegative, true},
		{ReactionLike, true},
		{ReactionImportant, true},
		{ReactionQuestion, true},
		{ReactionType("invalid"), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.reaction), func(t *testing.T) {
			err := tt.reaction.Validate()
			if tt.valid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestJSONB_ValueAndScan(t *testing.T) {
	// Test Value method
	jsonb := JSONB{"test": "value", "number": 42}
	value, err := jsonb.Value()
	assert.NoError(t, err)
	assert.NotNil(t, value)

	// Test nil JSONB
	var nilJSONB JSONB
	nilValue, err := nilJSONB.Value()
	assert.NoError(t, err)
	assert.Nil(t, nilValue)

	// Test Scan method
	var scannedJSONB JSONB
	jsonBytes := []byte(`{"scanned": true, "data": "test"}`)
	err = scannedJSONB.Scan(jsonBytes)
	assert.NoError(t, err)
	assert.Equal(t, true, scannedJSONB["scanned"])
	assert.Equal(t, "test", scannedJSONB["data"])

	// Test Scan with nil
	var nilScannedJSONB JSONB
	err = nilScannedJSONB.Scan(nil)
	assert.NoError(t, err)
	assert.Nil(t, nilScannedJSONB)
}

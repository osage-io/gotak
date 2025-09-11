-- Rollback Chat System Migration
-- Removes all chat-related tables and data

-- Drop indexes first
DROP INDEX IF EXISTS idx_message_reactions_type;
DROP INDEX IF EXISTS idx_message_reactions_message_id;

DROP INDEX IF EXISTS idx_message_acknowledgments_user_id;
DROP INDEX IF EXISTS idx_message_acknowledgments_message_id;

DROP INDEX IF EXISTS idx_chat_messages_active;
DROP INDEX IF EXISTS idx_chat_messages_thread;
DROP INDEX IF EXISTS idx_chat_messages_cot_uid;
DROP INDEX IF EXISTS idx_chat_messages_location;
DROP INDEX IF EXISTS idx_chat_messages_classification;
DROP INDEX IF EXISTS idx_chat_messages_type_priority;
DROP INDEX IF EXISTS idx_chat_messages_created_at;
DROP INDEX IF EXISTS idx_chat_messages_sender_id;
DROP INDEX IF EXISTS idx_chat_messages_room_id;

DROP INDEX IF EXISTS idx_chat_room_participants_active;
DROP INDEX IF EXISTS idx_chat_room_participants_user_id;
DROP INDEX IF EXISTS idx_chat_room_participants_room_id;

DROP INDEX IF EXISTS idx_chat_rooms_classification;
DROP INDEX IF EXISTS idx_chat_rooms_active;
DROP INDEX IF EXISTS idx_chat_rooms_type;

-- Drop tables in reverse dependency order
DROP TABLE IF EXISTS message_reactions;
DROP TABLE IF EXISTS message_acknowledgments;
DROP TABLE IF EXISTS chat_messages;
DROP TABLE IF EXISTS chat_room_participants;
DROP TABLE IF EXISTS chat_rooms;

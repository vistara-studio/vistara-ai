-- Create user bookmarks table
CREATE TABLE user_bookmarks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(100) NOT NULL,
    destination_id VARCHAR(50) NOT NULL REFERENCES destinations(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(user_id, destination_id)
);

-- Create indexes for bookmarks
CREATE INDEX idx_user_bookmarks_user_id ON user_bookmarks (user_id);
CREATE INDEX idx_user_bookmarks_destination_id ON user_bookmarks (destination_id);
CREATE INDEX idx_user_bookmarks_created_at ON user_bookmarks (created_at DESC);

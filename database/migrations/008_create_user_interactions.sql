-- Create user interactions table for tracking behavior and recommendations
CREATE TABLE user_interactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(100) NOT NULL,
    destination_id VARCHAR(50) REFERENCES destinations(id) ON DELETE SET NULL,
    itinerary_id UUID REFERENCES travel_itineraries(id) ON DELETE SET NULL,
    interaction_type VARCHAR(50) NOT NULL, -- viewed, saved, shared, reviewed, added_to_itinerary
    metadata JSONB, -- additional data like rating, search_query, etc.
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for user interactions
CREATE INDEX idx_user_interactions_user_id ON user_interactions (user_id);
CREATE INDEX idx_user_interactions_destination_id ON user_interactions (destination_id);
CREATE INDEX idx_user_interactions_itinerary_id ON user_interactions (itinerary_id);
CREATE INDEX idx_user_interactions_type ON user_interactions (interaction_type);
CREATE INDEX idx_user_interactions_created_at ON user_interactions (created_at DESC);

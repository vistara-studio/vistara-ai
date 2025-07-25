-- Create reviews table for destination reviews
CREATE TABLE reviews (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    destination_id VARCHAR(50) NOT NULL REFERENCES destinations(id) ON DELETE CASCADE,
    user_id VARCHAR(100) NOT NULL,
    user_name VARCHAR(255),
    user_avatar TEXT,
    is_verified BOOLEAN DEFAULT false,
    rating INTEGER NOT NULL CHECK (rating >= 1 AND rating <= 5),
    comment TEXT,
    helpful_count INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for reviews
CREATE INDEX idx_reviews_destination_id ON reviews (destination_id);
CREATE INDEX idx_reviews_user_id ON reviews (user_id);
CREATE INDEX idx_reviews_rating ON reviews (rating);
CREATE INDEX idx_reviews_created_at ON reviews (created_at DESC);

-- Add trigger for updated_at
CREATE TRIGGER update_reviews_updated_at BEFORE UPDATE
    ON reviews FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

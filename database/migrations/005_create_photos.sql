-- Create destination photos table for photo galleries
CREATE TABLE destination_photos (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    destination_id VARCHAR(50) NOT NULL REFERENCES destinations(id) ON DELETE CASCADE,
    photo_url TEXT NOT NULL,
    caption TEXT,
    photographer VARCHAR(255),
    is_featured BOOLEAN DEFAULT false,
    sort_order INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for photos
CREATE INDEX idx_destination_photos_destination_id ON destination_photos (destination_id);
CREATE INDEX idx_destination_photos_featured ON destination_photos (is_featured);
CREATE INDEX idx_destination_photos_sort_order ON destination_photos (sort_order);

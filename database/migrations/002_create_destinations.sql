-- Create destinations table with PostGIS support
CREATE TABLE destinations (
    id VARCHAR(50) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    short_description VARCHAR(500),
    location GEOGRAPHY(POINT, 4326), -- PostGIS geography point
    main_image_url TEXT,
    images TEXT[], -- Array of image URLs
    category VARCHAR(100),
    tags TEXT[], -- Array of tags for recommendations
    price_amount DECIMAL(12, 2),
    price_currency VARCHAR(10) DEFAULT 'IDR',
    price_unit VARCHAR(50) DEFAULT 'per_person', -- per_person, per_group, per_hour, etc.
    booking_required BOOLEAN DEFAULT false,
    booking_url TEXT,
    open_hours TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create spatial index for location queries
CREATE INDEX idx_destinations_location ON destinations USING GIST (location);

-- Create indexes for frequently queried fields
CREATE INDEX idx_destinations_category ON destinations (category);
CREATE INDEX idx_destinations_price ON destinations (price_amount);
CREATE INDEX idx_destinations_created_at ON destinations (created_at DESC);
CREATE INDEX idx_destinations_tags ON destinations USING GIN (tags);

-- Add trigger for updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_destinations_updated_at BEFORE UPDATE
    ON destinations FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

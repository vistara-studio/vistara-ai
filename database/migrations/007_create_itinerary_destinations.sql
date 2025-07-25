-- Create itinerary destinations table (many-to-many relationship)
CREATE TABLE itinerary_destinations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    itinerary_id UUID NOT NULL REFERENCES travel_itineraries(id) ON DELETE CASCADE,
    destination_id VARCHAR(50) REFERENCES destinations(id) ON DELETE SET NULL,
    day_number INTEGER NOT NULL,
    visit_order INTEGER NOT NULL,
    planned_time TIME,
    duration_minutes INTEGER,
    notes TEXT,
    is_completed BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(itinerary_id, day_number, visit_order)
);

-- Create indexes for itinerary destinations
CREATE INDEX idx_itinerary_destinations_itinerary_id ON itinerary_destinations (itinerary_id);
CREATE INDEX idx_itinerary_destinations_destination_id ON itinerary_destinations (destination_id);
CREATE INDEX idx_itinerary_destinations_day_order ON itinerary_destinations (day_number, visit_order);

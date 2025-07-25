-- Create travel itineraries table for Smart Planner
CREATE TABLE travel_itineraries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(100) NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    destination_name VARCHAR(255),
    start_date DATE,
    end_date DATE,
    duration_days INTEGER,
    total_budget DECIMAL(15, 2),
    budget_currency VARCHAR(10) DEFAULT 'IDR',
    interests TEXT[], -- user interests like culture, nature, etc.
    itinerary_data JSONB, -- detailed itinerary plan from AI
    status VARCHAR(50) DEFAULT 'active', -- active, completed, cancelled
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for itineraries
CREATE INDEX idx_travel_itineraries_user_id ON travel_itineraries (user_id);
CREATE INDEX idx_travel_itineraries_status ON travel_itineraries (status);
CREATE INDEX idx_travel_itineraries_created_at ON travel_itineraries (created_at DESC);
CREATE INDEX idx_travel_itineraries_start_date ON travel_itineraries (start_date);

-- Add trigger for updated_at
CREATE TRIGGER update_travel_itineraries_updated_at BEFORE UPDATE
    ON travel_itineraries FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Destination Explorer Database Schema
-- This schema supports the destination explorer feature with manuscripts, narration, and recommendations

-- Enable PostGIS extension for geospatial queries
CREATE EXTENSION IF NOT EXISTS postgis;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Table for tourist destinations
CREATE TABLE destinations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    short_description VARCHAR(500), -- For card display
    location GEOGRAPHY(POINT, 4326), -- For geospatial queries (latitude, longitude)
    main_image_url VARCHAR(500),
    images TEXT[], -- Array of image URLs for gallery
    category VARCHAR(100), -- Cultural, Historical, Religious, Nature, etc.
    tags TEXT[], -- Array of tags for categorization (e.g., ['history', 'culture', 'ancient'])
    price_amount DECIMAL(10,2), -- Ticket price
    price_currency VARCHAR(10) DEFAULT 'IDR',
    price_unit VARCHAR(50), -- per person, per group, etc.
    booking_required BOOLEAN DEFAULT false,
    booking_url VARCHAR(500),
    open_hours VARCHAR(200),
    last_entry VARCHAR(100),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Index for geospatial queries
CREATE INDEX idx_destinations_location ON destinations USING GIST(location);
-- Index for tag searches
CREATE INDEX idx_destinations_tags ON destinations USING GIN(tags);
-- Index for text search
CREATE INDEX idx_destinations_name_desc ON destinations USING GIN(to_tsvector('english', name || ' ' || description));

-- Table for manuscripts linked to destinations
CREATE TABLE manuscripts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    destination_id UUID REFERENCES destinations(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    cultural_story TEXT NOT NULL, -- Text to be converted into audio narration
    original_script_text TEXT, -- Original text in ancient script
    translation_text TEXT, -- Translation into English/Indonesian
    manuscript_image_url VARCHAR(500), -- URL to the digital manuscript image
    audio_narration_url VARCHAR(500), -- URL to generated audio narration
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Index for destination lookup
CREATE INDEX idx_manuscripts_destination ON manuscripts(destination_id);

-- Table for user reviews
CREATE TABLE reviews (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    destination_id UUID REFERENCES destinations(id) ON DELETE CASCADE,
    user_id UUID NOT NULL, -- Reference to the users table in Service-Users
    user_name VARCHAR(255),
    user_avatar VARCHAR(500),
    rating INT CHECK (rating >= 1 AND rating <= 5),
    comment TEXT,
    images TEXT[], -- Array of review photo URLs
    is_verified BOOLEAN DEFAULT false, -- Verified visitor
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes for reviews
CREATE INDEX idx_reviews_destination ON reviews(destination_id);
CREATE INDEX idx_reviews_user ON reviews(user_id);
CREATE INDEX idx_reviews_rating ON reviews(rating);
CREATE INDEX idx_reviews_created_at ON reviews(created_at);

-- Table for user itineraries
CREATE TABLE itineraries (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,
    name VARCHAR(255) DEFAULT 'My Cultural Journey',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Index for user itineraries
CREATE INDEX idx_itineraries_user ON itineraries(user_id);

-- Join table for destinations within an itinerary
CREATE TABLE itinerary_destinations (
    itinerary_id UUID REFERENCES itineraries(id) ON DELETE CASCADE,
    destination_id UUID REFERENCES destinations(id) ON DELETE CASCADE,
    visit_date DATE,
    notes TEXT,
    added_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (itinerary_id, destination_id)
);

-- Table for tracking user interactions (for recommendation engine)
CREATE TABLE user_interactions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,
    destination_id UUID REFERENCES destinations(id) ON DELETE CASCADE,
    action_type VARCHAR(50) NOT NULL, -- 'viewed', 'saved', 'reviewed', 'visited'
    rating INT CHECK (rating >= 1 AND rating <= 5), -- Only for 'reviewed' actions
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes for user interactions (important for recommendation engine)
CREATE INDEX idx_user_interactions_user ON user_interactions(user_id);
CREATE INDEX idx_user_interactions_destination ON user_interactions(destination_id);
CREATE INDEX idx_user_interactions_action ON user_interactions(action_type);
CREATE INDEX idx_user_interactions_created_at ON user_interactions(created_at);
CREATE INDEX idx_user_interactions_composite ON user_interactions(user_id, destination_id, action_type);

-- Table for storing generated audio narrations metadata
CREATE TABLE narrations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    manuscript_id UUID REFERENCES manuscripts(id) ON DELETE CASCADE,
    audio_url VARCHAR(500) NOT NULL,
    language_code VARCHAR(10) NOT NULL, -- 'id-ID', 'en-US', etc.
    voice_gender VARCHAR(10), -- 'male', 'female'
    voice_style VARCHAR(20), -- 'dramatic', 'calm', 'friendly'
    duration_seconds DECIMAL(10,2), -- Audio duration in seconds
    file_size_bytes BIGINT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Index for narrations
CREATE INDEX idx_narrations_manuscript ON narrations(manuscript_id);
CREATE INDEX idx_narrations_language ON narrations(language_code);

-- Table for user bookmarks
CREATE TABLE user_bookmarks (
    user_id UUID NOT NULL,
    destination_id UUID REFERENCES destinations(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (user_id, destination_id)
);

-- Index for bookmarks
CREATE INDEX idx_user_bookmarks_user ON user_bookmarks(user_id);
CREATE INDEX idx_user_bookmarks_destination ON user_bookmarks(destination_id);

-- Table for destination photos/gallery
CREATE TABLE destination_photos (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    destination_id UUID REFERENCES destinations(id) ON DELETE CASCADE,
    photo_url VARCHAR(500) NOT NULL,
    caption VARCHAR(255),
    photographer VARCHAR(255),
    is_main BOOLEAN DEFAULT false,
    order_index INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Index for photos
CREATE INDEX idx_destination_photos_destination ON destination_photos(destination_id);
CREATE INDEX idx_destination_photos_main ON destination_photos(is_main);
CREATE INDEX idx_destination_photos_order ON destination_photos(destination_id, order_index);

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Triggers to auto-update updated_at
CREATE TRIGGER update_destinations_updated_at BEFORE UPDATE ON destinations FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_manuscripts_updated_at BEFORE UPDATE ON manuscripts FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_itineraries_updated_at BEFORE UPDATE ON itineraries FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Seed data based on the provided cultural destinations
INSERT INTO destinations (id, name, description, short_description, location, main_image_url, images, category, tags, price_amount, price_currency, booking_required, open_hours) VALUES
(
    '123e4567-e89b-12d3-a456-426614174001',
    'Trowulan, Mojokerto',
    'The capital city of the 14th-century Majapahit Empire, where the traces of the archipelago''s glory are preserved.',
    'Ancient capital of the mighty Majapahit Empire',
    ST_SetSRID(ST_MakePoint(112.3833, -7.5500), 4326)::geography,
    'https://storage.vistara.com/destinations/trowulan-main.jpg',
    ARRAY[
        'https://storage.vistara.com/destinations/trowulan-1.jpg',
        'https://storage.vistara.com/destinations/trowulan-2.jpg',
        'https://storage.vistara.com/destinations/trowulan-3.jpg',
        'https://storage.vistara.com/destinations/trowulan-4.jpg'
    ],
    'Historical',
    ARRAY['history', 'ancient-kingdom', 'majapahit', 'archaeology', 'culture'],
    5000,
    'IDR',
    true,
    '08:00 - 17:00'
),
(
    '123e4567-e89b-12d3-a456-426614174002',
    'Cirebon',
    'The city where Sunan Gunung Jati spread Islam through a cultural approach.',
    'Historic Islamic sultanate with rich cultural heritage',
    ST_SetSRID(ST_MakePoint(108.5667, -6.7167), 4326)::geography,
    'https://storage.vistara.com/destinations/cirebon-main.jpg',
    ARRAY[
        'https://storage.vistara.com/destinations/cirebon-1.jpg',
        'https://storage.vistara.com/destinations/cirebon-2.jpg',
        'https://storage.vistara.com/destinations/cirebon-3.jpg'
    ],
    'Religious',
    ARRAY['history', 'islamic-heritage', 'sultanate', 'culture', 'religious'],
    3000,
    'IDR',
    false,
    '24 hours'
),
(
    '123e4567-e89b-12d3-a456-426614174003',
    'Luwu, South Sulawesi',
    'Home to the world''s longest literary epic, I La Galigo.',
    'Birthplace of the epic I La Galigo literature',
    ST_SetSRID(ST_MakePoint(120.2000, -3.0000), 4326)::geography,
    'https://storage.vistara.com/destinations/luwu-main.jpg',
    ARRAY[
        'https://storage.vistara.com/destinations/luwu-1.jpg',
        'https://storage.vistara.com/destinations/luwu-2.jpg'
    ],
    'Cultural',
    ARRAY['literature', 'epic', 'bugis-culture', 'maritime', 'heritage'],
    0,
    'IDR',
    false,
    'Various locations'
),
(
    '123e4567-e89b-12d3-a456-426614174004',
    'Surakarta (Solo)',
    'The City of Javanese Literature, birthplace of the great work, Serat Centhini.',
    'Royal city and center of Javanese culture',
    ST_SetSRID(ST_MakePoint(110.8167, -7.5667), 4326)::geography,
    'https://storage.vistara.com/destinations/solo-main.jpg',
    ARRAY[
        'https://storage.vistara.com/destinations/solo-1.jpg',
        'https://storage.vistara.com/destinations/solo-2.jpg',
        'https://storage.vistara.com/destinations/solo-3.jpg',
        'https://storage.vistara.com/destinations/solo-4.jpg',
        'https://storage.vistara.com/destinations/solo-5.jpg'
    ],
    'Cultural',
    ARRAY['literature', 'javanese-culture', 'royal-city', 'traditional-arts', 'philosophy'],
    10000,
    'IDR',
    true,
    '08:00 - 16:00'
),
(
    '123e4567-e89b-12d3-a456-426614174005',
    'Samudra Pasai, North Aceh',
    'A witness to the early history of the spread of Islam in the archipelago.',
    'First Islamic sultanate in Southeast Asia',
    ST_SetSRID(ST_MakePoint(97.2333, 5.1667), 4326)::geography,
    'https://storage.vistara.com/destinations/pasai-main.jpg',
    ARRAY[
        'https://storage.vistara.com/destinations/pasai-1.jpg',
        'https://storage.vistara.com/destinations/pasai-2.jpg'
    ],
    'Historical',
    ARRAY['history', 'islamic-heritage', 'early-islam', 'sultanate', 'maritime'],
    2000,
    'IDR',
    false,
    '08:00 - 18:00'
);

-- Insert corresponding manuscripts
INSERT INTO manuscripts (id, destination_id, title, cultural_story, original_script_text, translation_text, manuscript_image_url) VALUES
(
    '223e4567-e89b-12d3-a456-426614174001',
    '123e4567-e89b-12d3-a456-426614174001',
    'Nagarakretagama',
    'In Trowulan, the heart of the 14th-century Majapahit Kingdom, lies the legacy of the archipelago''s golden age. The Nagarakretagama manuscript, written by Mpu Prapanca, chronicles King Hayam Wuruk''s grand tour across Majapahit''s territories from Sumatra, Kalimantan, Bali, to Nusa Tenggara.',
    'I bhumi Nusantara ika nata rakawi sang amukti palapa Sang Maha Patih Gajah Mada.',
    'The entire Nusantara region was unified by the great poet and the executor of the Palapa Oath, Prime Minister Gajah Mada.',
    'https://storage.vistara.com/manuscripts/nagarakretagama.jpg'
),
(
    '223e4567-e89b-12d3-a456-426614174002',
    '123e4567-e89b-12d3-a456-426614174002',
    'Babad Cirebon',
    'Sunan Gunung Jati spread Islam in Cirebon using cultural and diplomatic approaches. The Babad Cirebon chronicles the establishment of the city, its palaces, and its interactions with China and the Middle East.',
    'Ingsun Sunan Gunung Jati, kang anedeg ing nagara Cirebon, maringake piwulang marang bangsa.',
    'I, Sunan Gunung Jati, who founded the state of Cirebon, provide teachings to the people.',
    'https://storage.vistara.com/manuscripts/babad-cirebon.jpg'
),
(
    '223e4567-e89b-12d3-a456-426614174003',
    '123e4567-e89b-12d3-a456-426614174003',
    'I La Galigo',
    'I La Galigo is the world''s longest Bugis epic, telling the story of Sawerigading, a heavenly prince who sailed the seas in search of love and destiny. Written in the Lontara script, this manuscript describes the cosmological values and life philosophy of the Bugis people.',
    'Ia muané Sawerigading, malajangngi ri bolae, naé mappatemme ri lauté.',
    'This is Sawerigading, a brave man who left the palace and sailed the vast ocean.',
    'https://storage.vistara.com/manuscripts/la-galigo.jpg'
),
(
    '223e4567-e89b-12d3-a456-426614174004',
    '123e4567-e89b-12d3-a456-426614174004',
    'Serat Centhini',
    'Serat Centhini is a masterpiece of Javanese literature that records the spiritual journey of Mas Cebolang. In the form of a song (tembang), this manuscript contains philosophy, religious teachings, culinary arts, and Javanese culture from both an inner and logical perspective.',
    'Cebolang lumaku ngalor kidul, nyekseni lakuning urip lan ngudi kawruh sejati.',
    'Cebolang travels north and south, witnessing life and seeking true knowledge.',
    'https://storage.vistara.com/manuscripts/serat-centhini.jpg'
),
(
    '223e4567-e89b-12d3-a456-426614174005',
    '123e4567-e89b-12d3-a456-426614174005',
    'Hikayat Raja-Raja Pasai',
    'In this ancient port city, Merah Silu received a revelation in a dream to embrace Islam, later becoming Sultan Malik Al-Saleh. The Hikayat Raja-Raja Pasai tells the early history of the spread of Islam in the archipelago, long before the Sultanates of Malacca and Demak.',
    'Maka disuruh oleh nabi bermimpi pada Merah Silu: hendaklah engkau masuk Islam.',
    'Then, in a dream, the prophet commanded Merah Silu: you shall convert to Islam.',
    'https://storage.vistara.com/manuscripts/hikayat-pasai.jpg'
);

-- Insert sample reviews with enhanced data
INSERT INTO reviews (destination_id, user_id, user_name, user_avatar, rating, comment, images, is_verified) VALUES
('123e4567-e89b-12d3-a456-426614174001', 'user-001', 'Andi Pratama', 'https://storage.vistara.com/avatars/user-001.jpg', 5, 'Amazing historical site! The Majapahit ruins are truly magnificent.', ARRAY['https://storage.vistara.com/reviews/trowulan-user1-1.jpg'], true),
('123e4567-e89b-12d3-a456-426614174001', 'user-002', 'Sarah Dewi', 'https://storage.vistara.com/avatars/user-002.jpg', 4, 'Great cultural experience, learned so much about Indonesian history.', NULL, false),
('123e4567-e89b-12d3-a456-426614174002', 'user-003', 'Bambang', 'https://storage.vistara.com/avatars/user-003.jpg', 5, 'Beautiful architecture and rich Islamic heritage.', ARRAY['https://storage.vistara.com/reviews/cirebon-user3-1.jpg', 'https://storage.vistara.com/reviews/cirebon-user3-2.jpg'], true),
('123e4567-e89b-12d3-a456-426614174003', 'user-004', 'Rizka Imannia', 'https://storage.vistara.com/avatars/user-004.jpg', 4, 'The I La Galigo story is fascinating, truly epic literature.', NULL, false),
('123e4567-e89b-12d3-a456-426614174004', 'user-005', 'Charles', 'https://storage.vistara.com/avatars/user-005.jpg', 5, 'Solo is the heart of Javanese culture, absolutely loved it!', ARRAY['https://storage.vistara.com/reviews/solo-user5-1.jpg'], true);

-- Insert sample user interactions for recommendation engine
INSERT INTO user_interactions (user_id, destination_id, action_type, rating) VALUES
('user-001', '123e4567-e89b-12d3-a456-426614174001', 'viewed', NULL),
('user-001', '123e4567-e89b-12d3-a456-426614174001', 'saved', NULL),
('user-001', '123e4567-e89b-12d3-a456-426614174001', 'reviewed', 5),
('user-002', '123e4567-e89b-12d3-a456-426614174001', 'viewed', NULL),
('user-002', '123e4567-e89b-12d3-a456-426614174002', 'saved', NULL),
('user-003', '123e4567-e89b-12d3-a456-426614174002', 'reviewed', 5),
('user-004', '123e4567-e89b-12d3-a456-426614174003', 'visited', NULL),
('user-005', '123e4567-e89b-12d3-a456-426614174004', 'reviewed', 5);

-- View for destination search with review aggregation
CREATE VIEW destination_search_view AS
SELECT 
    d.id,
    d.name,
    d.description,
    ST_Y(d.location::geometry) as latitude,
    ST_X(d.location::geometry) as longitude,
    d.main_image_url,
    d.tags,
    d.created_at,
    COALESCE(AVG(r.rating), 0) as avg_rating,
    COUNT(r.id) as review_count,
    COALESCE(COUNT(DISTINCT ui.user_id) FILTER (WHERE ui.action_type = 'viewed'), 0) as view_count,
    COALESCE(COUNT(DISTINCT ui.user_id) FILTER (WHERE ui.action_type = 'saved'), 0) as save_count
FROM destinations d
LEFT JOIN reviews r ON d.id = r.destination_id
LEFT JOIN user_interactions ui ON d.id = ui.destination_id
GROUP BY d.id, d.name, d.description, d.location, d.main_image_url, d.tags, d.created_at;

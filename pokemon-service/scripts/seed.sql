-- Data contoh untuk local development. Jalankan via: make seed
INSERT INTO pokemons (id, pokedex_number, name, primary_type, secondary_type, hp, attack, defense, speed) VALUES
    (gen_random_uuid(), 1,  'bulbasaur',  'grass',    'poison', 45, 49, 49, 45),
    (gen_random_uuid(), 4,  'charmander', 'fire',     NULL,     39, 52, 43, 65),
    (gen_random_uuid(), 6,  'charizard',  'fire',     'flying', 78, 84, 78, 100),
    (gen_random_uuid(), 7,  'squirtle',   'water',    NULL,     44, 48, 65, 43),
    (gen_random_uuid(), 25, 'pikachu',    'electric', NULL,     35, 55, 40, 90),
    (gen_random_uuid(), 94, 'gengar',     'ghost',    'poison', 60, 65, 60, 110)
ON CONFLICT DO NOTHING;

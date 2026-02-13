CREATE TABLE IF NOT EXISTS meal_food_mappings (
    meal_id INTEGER NOT NULL,
    food_id INTEGER NOT NULL,
    PRIMARY KEY (meal_id, food_id),
    FOREIGN KEY (meal_id) REFERENCES meals(id),
    FOREIGN KEY (food_id) REFERENCES foods(id)
);
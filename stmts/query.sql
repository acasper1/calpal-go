-- name: GetMeals :many
SELECT * FROM meals;

-- name: GetMealById :one
SELECT * FROM meals WHERE id = ?;

-- name: GetFoods :many
SELECT * FROM foods;

-- name: GetFoodById :one
SELECT * FROM foods WHERE id = ?;

-- name: CreateMeal :one
INSERT INTO meals (name) VALUES (?)
RETURNING *;

-- name: CreateFood :one
INSERT INTO foods (name, calories) VALUES (?, ?)
RETURNING *;

-- name: CreateMealFoodMapping :one
INSERT INTO meal_food_mappings (meal_id, food_id) VALUES (?, ?)
RETURNING *;

-- name: GetAllFoodsAndMeals :many
SELECT m.name as meal_name, f.name as food_name, f.calories
FROM meals m
JOIN meal_food_mappings mf ON m.id = mf.meal_id
JOIN foods f ON mf.food_id = f.id;

-- name: GetAllFoodsAndMealsByDate :many
SELECT m.name as meal_name, f.name as food_name, f.calories
FROM meals m
JOIN meal_food_mappings mf ON m.id = mf.meal_id
JOIN foods f ON mf.food_id = f.id
WHERE DATE(m.created_at) = DATE(?);

-- name: UpdateFood :one
UPDATE foods 
SET name = ?, calories = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING *;

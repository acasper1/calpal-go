package stmts

const InsertMeals string = `INSERT INTO meals (name) VALUES ($1);`

const GetMeals string = `SELECT name FROM meals;`

const GetMealsAll string = `SELECT * FROM meals;`

const InsertFoods string = `INSERT INTO foods (name, calories) VALUES ($1, $2);`

const GetFoods string = `SELECT name, calories FROM foods;`

const GetFoodsAll string = `SELECT * FROM foods;`

const InsertMealFoodMapping string = `INSERT INTO meal_food_mappings (meal_id, food_id) VALUES ($1, $2);`

const GetAllFoodsAndMeals string = `
	SELECT m.name, f.name, f.calories
	FROM meals m
	JOIN meal_food_mappings mf ON m.id = mf.meal_id
	JOIN foods f ON mf.food_id = f.id;`

const GetMealsWithFoodsByDate string = `
	SELECT m.name, f.name, f.calories
	FROM meals m
	JOIN meal_food_mappings mf ON m.id = mf.meal_id
	JOIN foods f ON mf.food_id = f.id
	WHERE DATE(m.created_at) = DATE($1);`

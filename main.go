package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"calpal-go/migrations"
	"calpal-go/stmts"

	// consider replacing with modern sqlite implementation
	_ "github.com/glebarez/go-sqlite"
)

type MealPageData struct {
	PageTitle string
	Meals     []MealFood
}

type Food struct {
	name     string
	calories int16
}

type MealFood struct {
	MealName string
	FoodName string
	Calories int16
}

var db *sql.DB

func MealsHandler(w http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case http.MethodGet:
		GetMeals(w, request)
	case http.MethodPost:
		AddMeal(w, request)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(`Method Not Allowed`))
	}
}

func FoodsHandler(w http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case http.MethodGet:
		GetFoods(w, request)
	case http.MethodPost:
		AddFood(w, request)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(`Method Not Allowed`))
	}
}

func GetMeals(w http.ResponseWriter, request *http.Request) {
	files := []string{
		"./templates/base.html",
		"./templates/index.html",
	}
	tmpl := template.Must(template.ParseFiles(files...))
	if db == nil {
		log.Fatal("Database connection closed!")
	}

	rows, err := db.Query(stmts.GetAllFoodsAndMeals)
	if err != nil {
		log.Fatal(err)
	}

	var meals []MealFood
	for rows.Next() {
		var meal MealFood
		if err = rows.Scan(&meal.MealName, &meal.FoodName, &meal.Calories); err != nil {
			log.Print(err)
		}
		meals = append(meals, meal)
	}

	data := MealPageData{
		PageTitle: "My Meals",
		Meals:     meals,
	}

	err = tmpl.ExecuteTemplate(w, "base", data)
	if err != nil {
		log.Printf("Failed to execute template: %s\n", err)
	}

}

func AddMeal(w http.ResponseWriter, request *http.Request) {
	var mealName string
	var foodName string
	var calories int
	var err error
	files := []string{
		"./templates/base.html",
		"./templates/index.html",
	}

	mealName = request.FormValue("meal-name")
	foodName = request.FormValue("food")
	calories, err = strconv.Atoi(request.FormValue("calories"))
	if err != nil {
		log.Fatal(err)
	}

	// Insert new food record
	var stmt *sql.Stmt
	stmt, err = db.Prepare(stmts.InsertFoods)
	if err != nil {
		log.Fatal(err)
	}
	var res sql.Result
	res, err = stmt.Exec(foodName, calories)
	if err != nil {
		log.Fatal(err)
	}

	// Add meal and meal_food_mapping records
	foodId, _ := res.LastInsertId()

	stmt, err = db.Prepare(stmts.InsertMeals)
	if err != nil {
		log.Fatal(err)
	}
	res, err = stmt.Exec(mealName)
	if err != nil {
		log.Fatal(err)
	}
	mealId, _ := res.LastInsertId()

	stmt, err = db.Prepare(stmts.InsertMealFoodMapping)
	if err != nil {
		log.Fatal(err)
	}
	_, err = stmt.Exec(mealId, foodId)
	if err != nil {
		log.Fatal(err)
	}

	// re-render the page with the new meal added
	tmpl := template.Must(template.ParseFiles(files...))
	tmpl.ExecuteTemplate(w, "meal", MealFood{
		MealName: mealName,
		FoodName: foodName,
		Calories: int16(calories),
	})
}

func GetFoods(w http.ResponseWriter, request *http.Request) []Food {
	rows, err := db.Query(stmts.GetFoods)
	if err != nil {
		log.Fatal(err)
	}

	var foods []Food
	for rows.Next() {
		var food Food
		if err = rows.Scan(&food.name, &food.calories); err != nil {
			log.Print(err)
		}
		foods = append(foods, food)
	}

	return foods
}

func AddFood(w http.ResponseWriter, request *http.Request) {

}

func insertTestData(db *sql.DB) {
	// Test data
	meals := []string{
		"Breakfast",
		"Lunch",
		"Dinner",
	}

	foods := []Food{
		{"Apple", 95},
		{"Banana", 105},
		{"Chicken Breast", 165},
		{"Rice (1 cup)", 206},
	}

	// Insert test meal data into database
	stmt, err := db.Prepare(stmts.InsertMeals)
	if err != nil {
		log.Fatal(err)
	}

	for _, m := range meals {
		_, err := stmt.Exec(m)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Insert test food data into database
	stmt, err = db.Prepare(stmts.InsertFoods)
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range foods {
		_, err := stmt.Exec(f.name, f.calories)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Insert test meal-food mappings into database
	// TODO use sqlite's sqlite_sequence table to get the last inserted id instead of hardcoding ids
	// Doing some basic math (subtracting the number of foods inserted from the max id) will get the id range for a transaction.
	stmt, err = db.Prepare(stmts.InsertMealFoodMapping)
	if err != nil {
		log.Fatal(err)
	}

	// Breakfast: Apple, Banana
	_, err = stmt.Exec(1, 1) // Breakfast - Apple
	if err != nil {
		log.Fatal(err)
	}
	_, err = stmt.Exec(2, 2) // Lunch - Banana
	if err != nil {
		log.Fatal(err)
	}

	// Dinner: Chicken Breast, Rice
	_, err = stmt.Exec(3, 3) // Dinner - Chicken Breast
	if err != nil {
		log.Fatal(err)
	}
	_, err = stmt.Exec(3, 4) // Dinner - Rice
	if err != nil {
		log.Fatal(err)
	}
	stmt.Close()
	// End test data insertion
}

func main() {
	// run db migrations on server start
	var err error // this prevents re-declaring db variable -- use the global instead
	db, err = sql.Open("sqlite", ":memory:")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	migrations.RunMigration(db)
	insertTestData(db)

	// Register routes and handlers
	http.HandleFunc("/meals/", MealsHandler)
	http.HandleFunc("/food/", FoodsHandler)

	http.ListenAndServe(":8080", nil)
}

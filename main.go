package main

import (
	"context"
	"database/sql"
	_ "embed"
	"html/template"
	"log"
	"net/http"
	"strconv"

	query "github.com/acasper1/calpal-go/stmts"

	_ "modernc.org/sqlite"
)

type MealPageData struct {
	PageTitle string
	Meals     []query.GetAllFoodsAndMealsRow
}

type FoodPageData struct {
	Foods []query.Food
}

//go:embed stmts/schema.sql
var ddl string
var db *sql.DB
var q query.Queries
var ctx context.Context

func MealsHandler(w http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case http.MethodGet:
		GetMeals(w, request)
	case http.MethodPost:
		AddMeal(w, request)
	case http.MethodPut:
		UpdateMeal(w, request)
	case http.MethodDelete:
		DeleteMeal(w, request)
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
	case http.MethodPut:
		UpdateFood(w, request)
	case http.MethodDelete:
		DeleteFood(w, request)
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

	meals, err := q.GetAllFoodsAndMeals(ctx)
	if err != nil {
		log.Fatal(err)
	}

	data := MealPageData{
		PageTitle: "All Meals",
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
	food, err := q.CreateFood(ctx, query.CreateFoodParams{Name: foodName, Calories: 0})
	if err != nil {
		log.Fatal(err)
	}

	// Add meal and meal_food_mapping records
	meal, err := q.CreateMeal(ctx, mealName)
	if err != nil {
		log.Fatal(err)
	}

	_, err = q.CreateMealFoodMapping(ctx, query.CreateMealFoodMappingParams{MealID: meal.ID, FoodID: food.ID})

	// re-render the page with the new meal added
	tmpl := template.Must(template.ParseFiles(files...))
	tmpl.ExecuteTemplate(w, "meal", MealFood{
		MealName: mealName,
		FoodName: foodName,
		Calories: int16(calories),
	})
}

func UpdateMeal(w http.ResponseWriter, request *http.Request) {}

func DeleteMeal(w http.ResponseWriter, request *http.Request) {}

func GetFoods(w http.ResponseWriter, request *http.Request) {
	files := []string{
		"./templates/base.html",
		"./templates/foods.html",
	}
	rows, err := db.Query(query.GetFoods)
	if err != nil {
		log.Fatal(err)
	}

	var foods []Food
	for rows.Next() {
		var food Food
		if err = rows.Scan(&food.FoodName, &food.Calories); err != nil {
			log.Print(err)
		}
		foods = append(foods, food)
	}

	pageData := FoodPageData{
		Foods: foods,
	}
	tmpl := template.Must(template.ParseFiles(files...))
	err = tmpl.ExecuteTemplate(w, "base", pageData)
	if err != nil {
		log.Printf("Failed to execute template: %s\n", err)
	}
}

func AddFood(w http.ResponseWriter, request *http.Request) {
	var foodName string
	var calories int
	var err error
	files := []string{
		"./templates/base.html",
		"./templates/foods.html",
	}

	foodName = request.FormValue("food-name")
	calories, err = strconv.Atoi(request.FormValue("calories"))
	if err != nil {
		log.Fatal(err)
	}

	var stmt *sql.Stmt
	stmt, err = db.Prepare(query.InsertFoods)
	if err != nil {
		log.Fatal(err)
	}
	_, err = stmt.Exec(foodName, calories)
	if err != nil {
		log.Fatal(err)
	}

	tmpl := template.Must(template.ParseFiles(files...))
	tmpl.ExecuteTemplate(w, "food", Food{
		FoodName: foodName,
		Calories: int16(calories),
	})
}

func UpdateFood(w http.ResponseWriter, request *http.Request) {
	var newFoodName string
	var newCalories int
	var err error
	files := []string{
		"./templates/base.html",
		"./templates/foods.html",
	}

	newFoodName = request.FormValue("food-name")
	newCalories, err = strconv.Atoi(request.FormValue("calories"))
	if err != nil {
		log.Fatal(err)
	}

	var stmt *sql.Stmt
	stmt, err = db.Prepare(query.GetFood)

	// First, get old record, change whatever values need changing in DAO struct, then update db record
	if err != nil {
		log.Fatal(err)
	}
	_, err = stmt.Exec(newFoodName, newCalories)
	if err != nil {
		log.Fatal(err)
	}

	tmpl := template.Must(template.ParseFiles(files...))
	tmpl.ExecuteTemplate(w, "food", Food{
		FoodName: newFoodName,
		Calories: int16(newCalories),
	})
}

func DeleteFood(w http.ResponseWriter, request *http.Request) {}

func run() error {
	ctx = context.Background()

	// run db migrations on server start
	var err error
	db, err = sql.Open("sqlite", ":memory:")
	if err != nil {
		return err
	}

	if _, err := db.ExecContext(ctx, ddl); err != nil {
		return err
	}

	q = *query.New(db)

	// Register routes and handlers
	http.HandleFunc("/meals/", MealsHandler)
	http.HandleFunc("/foods/", FoodsHandler)

	http.ListenAndServe(":8080", nil)
	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

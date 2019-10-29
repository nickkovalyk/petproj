package routing

import (
	"net/http"

	"gitlab.com/i4s-edu/petstore-kovalyk/db/mappers"

	"gitlab.com/i4s-edu/petstore-kovalyk/api/routing/handlers"

	"github.com/go-chi/chi"
	"github.com/jmoiron/sqlx"

	"gitlab.com/i4s-edu/petstore-kovalyk/api/routing/middlewares"
)

func NewRouter(db *sqlx.DB) http.Handler {
	r := chi.NewRouter()
	middlewares.SetMiddlewares(r)

	pet := handlers.Pet{
		PetMapper: mappers.PetMapper{DB: db}}
	store := handlers.Store{
		PetMapper:   mappers.PetMapper{DB: db},
		OrderMapper: mappers.OrderMapper{DB: db}}
	user := handlers.User{
		UserMapper: mappers.UserMapper{DB: db}}

	r.Route("/pet", func(r chi.Router) {
		r.Post("/", pet.Create)
		r.Put("/", pet.Update)
		r.Get("/{id}", pet.GetByID)
		r.Get("/findByStatus", pet.FindByStatus)
		r.Get("/findByTags", pet.FindByTags)
		r.Post("/{id}", pet.UpdateByID)
		r.Delete("/{id}", pet.Delete)
		r.Post("/{id}/uploadImage", pet.UploadImage)
	})
	r.Route("/store", func(r chi.Router) {
		r.Get("/inventory", store.GetInventory)
		r.Post("/order", store.CreateOrder)
		r.Get("/order/{id}", store.GetByID)
		r.Delete("/order/{id}", store.Delete)
	})
	r.Route("/user", func(r chi.Router) {
		r.Post("/", user.Create)
		r.Post("/createWithArray", user.CreateWithList)
		r.Post("/createWithList", user.CreateWithList)
		r.Get("/login", user.Login)
		r.Get("/logout", user.Logout)
		r.Get("/{username}", user.GetByUsername)
		r.Put("/{username}", user.Update)
		r.Delete("/{username}", user.Delete)

	})

	return r
}

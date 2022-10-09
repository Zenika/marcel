package api

import (
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"

	"github.com/allez-chauffe/marcel/pkg/api/auth"
	"github.com/allez-chauffe/marcel/pkg/api/clients"
	"github.com/allez-chauffe/marcel/pkg/api/medias"
	"github.com/allez-chauffe/marcel/pkg/api/plugins"
	"github.com/allez-chauffe/marcel/pkg/api/users"
	"github.com/allez-chauffe/marcel/pkg/config"
	"github.com/allez-chauffe/marcel/pkg/db"
	"github.com/allez-chauffe/marcel/pkg/module"
)

// Module creates API module
func Module() *module.Module {
	var clientsService *clients.Service
	var mediasService *medias.Service
	var usersService *users.Service
	var authService *auth.Service

	return &module.Module{
		Name: "API",
		Start: func(_ module.Context, next module.NextFunc) (module.StopFunc, error) {
			if err := db.Open(); err != nil {
				return nil, err
			}

			var stop = func() error {
				return db.Close()
			}

			if err := db.Users().EnsureOneUser(); err != nil {
				return stop, err
			}

			usersService = users.NewService()
			authService = auth.NewService()
			clientsService = clients.NewService()
			mediasService = medias.NewService(clientsService)

			plugins.Initialize()

			return stop, next()
		},
		HTTP: module.HTTP{
			BasePath: config.Default().API().BasePath(),
			Setup: func(_ module.Context, _ string, r *mux.Router) {
			    r.Use(LoggingMiddleware)
				r.Use(auth.Middleware)
				if !config.Default().API().Auth().Secure() {
					log.Warnln("Secure mode is disabled")
				}

				medias := r.PathPrefix("/medias").Subrouter()
				medias.HandleFunc("/", mediasService.GetAllHandler).Methods("GET")
				medias.HandleFunc("/", mediasService.CreateHandler).Methods("POST")
				medias.HandleFunc("/", mediasService.SaveHandler).Methods("PUT")

				media := medias.PathPrefix("/{idMedia:[0-9]*}").Subrouter()
				media.HandleFunc("/", mediasService.GetHandler).Methods("GET")
				media.HandleFunc("/", mediasService.DeleteHandler).Methods("DELETE")
				media.HandleFunc("/activate", mediasService.ActivateHandler).Methods("GET")
				media.HandleFunc("/deactivate", mediasService.DeactivateHandler).Methods("GET")
				media.HandleFunc("/plugins/{eltName}/{instanceId}/{filePath:.*}", mediasService.GetPluginFilesHandler).Methods("GET")

				clients := r.PathPrefix("/clients").Subrouter()
				clients.HandleFunc("/", clientsService.GetAllHandler).Methods("GET")
				clients.HandleFunc("/", clientsService.CreateHandler).Methods("POST")
				clients.HandleFunc("/", clientsService.UpdateHandler).Methods("PUT")
				clients.HandleFunc("/", clientsService.DeleteAllHandler).Methods("DELETE")

				client := clients.PathPrefix("/{clientID}").Subrouter()
				client.HandleFunc("/", clientsService.GetHandler).Methods("GET")
				client.HandleFunc("/", clientsService.DeleteHandler).Methods("DELETE")
				client.HandleFunc("/ws", clientsService.WSConnectionHandler)

				pluginsRouter := r.PathPrefix("/plugins").Subrouter()
				pluginsRouter.HandleFunc("/", plugins.GetAllHandler).Methods("GET")
				pluginsRouter.HandleFunc("/", plugins.AddHandler).Methods("POST")
				pluginsRouter.HandleFunc("/{eltName}", plugins.GetHandler).Methods("GET")
				pluginsRouter.HandleFunc("/{eltName}", plugins.UpdateHandler).Methods("PUT")
				pluginsRouter.HandleFunc("/{eltName}", plugins.DeleteHandler).Methods("DELETE")

				auth := r.PathPrefix("/auth").Subrouter()
				auth.HandleFunc("/login", authService.LoginHandler).Methods("POST")
				auth.HandleFunc("/logout", authService.LogoutHandler).Methods("PUT")
				auth.HandleFunc("/validate", authService.ValidateHandler).Methods("GET")
				auth.HandleFunc("/validate/admin", authService.ValidateAdminHandler).Methods("GET")

				users := auth.PathPrefix("/users").Subrouter()
				users.HandleFunc("/", usersService.CreateUserHandler).Methods("POST")
				users.HandleFunc("/", usersService.GetUsersHandler).Methods("GET")

				user := users.PathPrefix("/{userID}").Subrouter()
				user.HandleFunc("", usersService.DeleteUserHandler).Methods("DELETE")
				user.HandleFunc("", usersService.UpdateUserHandler).Methods("PUT")
			},
		},
	}
}

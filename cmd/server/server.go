package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/antoine-granier/urlshortener/internal/api"
	"github.com/antoine-granier/urlshortener/internal/repository"

	cmd2 "github.com/antoine-granier/urlshortener/cmd"
	"github.com/antoine-granier/urlshortener/internal/models"
	"github.com/antoine-granier/urlshortener/internal/monitor"
	"github.com/antoine-granier/urlshortener/internal/services"
	"github.com/antoine-granier/urlshortener/internal/workers"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"gorm.io/driver/sqlite" // Driver SQLite pour GORM
	"gorm.io/gorm"
)

// RunServerCmd représente la commande 'run-server' de Cobra.
// C'est le point d'entrée pour lancer le serveur de l'application.
var RunServerCmd = &cobra.Command{
	Use:   "run-server",
	Short: "Lance le serveur API de raccourcissement d'URLs et les processus de fond.",
	Long: `Cette commande initialise la base de données, configure les APIs,
démarre les workers asynchrones pour les clics et le moniteur d'URLs,
puis lance le serveur HTTP.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Charger la configuration globale
		cfg := cmd2.Cfg
		if cfg == nil {
			log.Fatalf("Configuration non initialisée : cmd.Cfg est nil")
		}

		// Initialiser la connexion à la base de données SQLite avec GORM
		db, err := gorm.Open(sqlite.Open(cfg.Database.Name), &gorm.Config{})
		if err != nil {
			log.Fatalf("Erreur de connexion à la BDD : %v", err)
		}

		// Migrations automatiques
		if err := db.AutoMigrate(&models.Link{}, &models.Click{}); err != nil {
			log.Fatalf("Erreur lors des migrations : %v", err)
		}

		// Initialiser les repositories
		linkRepo := repository.NewLinkRepository(db)
		clickRepo := repository.NewClickRepository(db)
		log.Println("Repositories initialisés.")

		// Initialiser les services métiers
		linkSvc := services.NewLinkService(linkRepo)
		log.Println("Services métiers initialisés.")

		// Initialiser le channel ClickEventsChannel et lancer les workers
		bufferSize := cfg.Analytics.BufferSize
		numWorkers := 3
		clickChan := make(chan models.ClickEvent, bufferSize)
		workers.StartClickWorkers(numWorkers, clickChan, clickRepo)

		log.Printf(
			"Channel d'événements de clic initialisé avec un buffer de %d. %d worker(s) de clics démarré(s).",
			bufferSize, numWorkers,
		)

		// Initialiser et lancer le moniteur d'URLs
		interval := time.Duration(cfg.Monitor.IntervalMinutes) * time.Minute
		urlMonitor := monitor.NewUrlMonitor(linkRepo, interval)
		go urlMonitor.Start()
		log.Printf("Moniteur d'URLs démarré avec un intervalle de %v.", interval)

		// Configurer le routeur Gin et les handlers API
		router := gin.Default()
		api.SetupRoutes(router, linkSvc)
		log.Println("Routes API configurées.")

		// Créer le serveur HTTP Gin
		serverAddr := fmt.Sprintf(":%d", cfg.Server.Port)
		srv := &http.Server{
			Addr:    serverAddr,
			Handler: router,
		}

		// Démarrer le serveur Gin dans une goroutine anonyme pour ne pas bloquer.
		go func() {
			log.Printf("Démarrage du serveur sur %s…", serverAddr)
			if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Fatalf("Erreur du serveur HTTP : %v", err)
			}
		}()

		// Gère l'arrêt propre du serveur (graceful shutdown)
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		// Bloquer jusqu'à ce qu'un signal d'arrêt soit reçu
		<-quit
		log.Println("Signal d'arrêt reçu. Arrêt du serveur...")

		log.Println("Arrêt en cours... Donnez un peu de temps aux workers pour finir.")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Fatalf("Erreur lors du shutdown : %v", err)
		}

		log.Println("Serveur arrêté proprement.")
	},
}

func init() {
	cmd2.RootCmd.AddCommand(RunServerCmd)
}

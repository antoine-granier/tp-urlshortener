package api

import (
	"log"
	"net/http"

	"github.com/axellelanca/urlshortener/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"gorm.io/gorm" // Pour gérer gorm.ErrRecordNotFound
)

// TODO Créer une variable ClickEventsChannel qui est un chan de type ClickEvent
// ClickEventsChannel est le channel global (ou injecté) utilisé pour envoyer les événements de clic
// aux workers asynchrones. Il est bufferisé pour ne pas bloquer les requêtes de redirection.

type ClickEvent struct {
	clicked int
}

// SetupRoutes configure toutes les routes de l'API Gin et injecte les dépendances nécessaires
func SetupRoutes(router *gin.Engine, linkService *services.LinkService) {
	// Le channel est initialisé ici.
	if ClickEventsChannel == nil {
		// TODO Créer le channel ici (make), il doit être bufférisé
		// La taille du buffer doit être configurable via Viper (cfg.Analytics.BufferSize)
		bufferSize := viper.GetInt("cfg.analytics.buffer_size")
		ClickEventsChannel := make(chan ClickEvent, bufferSize)
	}

	// router := gin.Default()
	// router.Use(GetLinkStatsHandler(linkService))

	// TODO : Route de Health Check , /health
	router.GET("/health", HealthCheckHandler())

	api := router.Group("/api/v1")
	{
		// TODO : Routes de l'API
		// Doivent être au format /api/v1/
		// POST /links
		router.POST("/links", func(c *gin.Context) {

		})

		// GET /links/:shortCode/stats
		router.GET("/links/:shortCode/stats", GetLinkStatsHandler())

		// Route de Redirection (au niveau racine pour les short codes)
		router.GET("/:shortCode", RedirectHandler(linkService))
	}
}

// HealthCheckHandler gère la route /health pour vérifier l'état du service.
func HealthCheckHandler(c *gin.Context) {
	return
	// TODO  Retourner simplement du JSON avec un StatusOK, {"status": "ok"}
}

// CreateLinkRequest représente le corps de la requête JSON pour la création d'un lien.
type CreateLinkRequest struct {
	LongURL string `json:"long_url" binding:"required,url"` // 'binding:required' pour validation, 'url' pour format URL
}

// CreateShortLinkHandler gère la création d'une URL courte.
func CreateShortLinkHandler(linkService *services.LinkService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreateLinkRequest
		// TODO : Tente de lier le JSON de la requête à la structure CreateLinkRequest.
		// Gin gère la validation 'binding'.

		// TODO: Appeler le LinkService (CreateLink pour créer le nouveau lien.

		// Retourne le code court et l'URL longue dans la réponse JSON.
		// TODO Choisir le bon code HTTP
		c.JSON(XXX, gin.H{
			"short_code":     link.ShortCode,
			"long_url":       link.LongURL,
			"full_short_url": "http://localhost:8080/" + link.ShortCode, // TODO: Utiliser cfg.Server.BaseURL ici
		})
	}
}

// RedirectHandler gère la redirection d'une URL courte vers l'URL longue et l'enregistrement asynchrone des clics.
func RedirectHandler(linkService *services.LinkService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Récupère le shortCode de l'URL avec c.Param
		shortCode := c.Param("shortCode")

		// TODO 2: Récupérer l'URL longue associée au shortCode depuis le linkService (GetLinkByShortCode)
		link, err := services.GetLinkByShortCode(shortCode)
		if err != nil {
			// Si le lien n'est pas trouvé, retourner HTTP 404 Not Found.
			// Utiliser errors.Is et l'erreur Gorm
			// Utilisez errors.Is(err, gorm.ErrRecordNotFound) en production si l'erreur est wrappée
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Link not found"})
				return
			}
			// Gérer d'autres erreurs potentielles de la base de données ou du service
			log.Printf("Error retrieving link for %s: %v", shortCode, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		// TODO 3: Créer un ClickEvent avec les informations pertinentes.
		clickEvent := ClickEvent{
			clicked: 1,
		}

		// TODO 4: Envoyer le ClickEvent dans le ClickEventsChannel avec le Multiplexage.
		// Utilise un `select` avec un `default` pour éviter de bloquer si le channel est plein.
		// Pour le default, juste un message à afficher :
		select {
		case ClickEventsChannel <- clickEvent:
			// Si l'événement est envoyé avec succès, logguer l'événement
			log.Printf("ClickEvent for %s successfully sent to the channel.", shortCode)
		default:
			// Si le canal est plein, afficher un message d'avertissement
			log.Printf("Warning: ClickEventsChannel is full, dropping click event for %s.", shortCode)
		}

		// log.Printf("Warning: ClickEventsChannel is full, dropping click event for %s.", shortCode)

		// TODO 5: Effectuer la redirection HTTP 302 (StatusFound) vers l'URL longue.

	}
}

// GetLinkStatsHandler gère la récupération des statistiques pour un lien spécifique.
func GetLinkStatsHandler(linkService *services.LinkService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO Récupère le shortCode de l'URL avec c.Param
		shortCode := c.Param("shortCode")

		// TODO 6: Appeler le LinkService pour obtenir le lien et le nombre total de clics.
		link, err := services.GetLinkByShortCode(shortCode)
		if err != nil {
			// Gérer le cas où le lien n'est pas trouvé (Gorm ErrRecordNotFound)
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Link not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
			return
		}
		totalClicks, err := linkService.GetTotalClicks(shortCode)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve total clicks"})
			return
		}
		// Gérer le cas où le lien n'est pas trouvé.
		// toujours avec l'erreur Gorm ErrRecordNotFound
		// Gérer d'autres erreurs

		// Retourne les statistiques dans la réponse JSON.
		c.JSON(http.StatusOK, gin.H{
			"short_code":   link.ShortCode,
			"long_url":     link.LongURL,
			"total_clicks": totalClicks,
		})
	}
}

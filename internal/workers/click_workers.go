package workers

import (
	"log"

	"github.com/antoine-granier/urlshortener/internal/models"
	"github.com/antoine-granier/urlshortener/internal/repository" // Nécessaire pour interagir avec le ClickRepository
)

// StartClickWorkers lance un pool de goroutines "workers" pour traiter les événements de clic.
// Chaque worker lira depuis le même 'clickEventsChan' et utilisera le 'clickRepo' pour la persistance.
func StartClickWorkers(workerCount int, clickEventsChan <-chan models.ClickEvent, clickRepo repository.ClickRepository) {
	log.Printf("Starting %d click worker(s)...", workerCount)
	for i := 0; i < workerCount; i++ {
		go clickWorker(clickEventsChan, clickRepo)
	}
}

// clickWorker est la fonction exécutée par chaque goroutine worker.
// Elle tourne indéfiniment, lisant les événements de clic dès qu'ils sont disponibles dans le channel.
func clickWorker(clickEventsChan <-chan models.ClickEvent, clickRepo repository.ClickRepository) {
	for event := range clickEventsChan { // Boucle qui lit les événements du channel
		// TODO 1: Convertir le 'ClickEvent' (reçu du channel) en un modèle 'models.Click'.
		click := &models.Click{
			LinkID: event.LinkID,
			// CreatedAt sera géré automatiquement par GORM si laissé à zéro
		}

		// TODO 2: Persister le clic en base de données via le 'clickRepo'.
		if err := clickRepo.CreateClick(click); err != nil {
			// En cas d'erreur, on logge l'échec
			log.Printf(
				"ERROR: Failed to save click for LinkID %d (UserAgent: %s, IP: %s): %v",
				event.LinkID, event.UserAgent, event.IPAddress, err,
			)
		} else {
			// Log optionnel pour confirmer l'enregistrement
			log.Printf("Click recorded successfully for LinkID %d", event.LinkID)
		}
	}
}

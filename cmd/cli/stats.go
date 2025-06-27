package cli

import (
	"fmt"
	"log"
	"os"

	cmd2 "github.com/axellelanca/urlshortener/cmd"
	"github.com/axellelanca/urlshortener/internal/repository"
	"github.com/axellelanca/urlshortener/internal/services"
	"github.com/spf13/cobra"

	"gorm.io/driver/sqlite" // Driver SQLite pour GORM
	"gorm.io/gorm"
)

//variable shortCodeFlag qui stockera la valeur du flag --code
var shortCodeFlag string

// StatsCmd représente la commande 'stats'
var StatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Affiche les statistiques (nombre de clics) pour un lien court.",
	Long: `Cette commande permet de récupérer et d'afficher le nombre total de clics
pour une URL courte spécifique en utilisant son code.

Exemple:
  url-shortener stats --code="xyz123"`,
	Run: func(cmd *cobra.Command, args []string) {
		// Valider que le flag --code a été fourni.
		// os.Exit(1) si erreur
		if shortCodeFlag == "" {
			fmt.Fprintln(os.Stderr, "Erreur : le flag --code est requis")
			os.Exit(1)
		}

		// Charger la configuration globale
		cfg := cmd2.Cfg
		if cfg == nil {
			log.Fatal("Configuration non initialisée")
		}

		// Initialiser la connexion à la base de données SQLite avec GORM.
		db, err := gorm.Open(sqlite.Open(cfg.Database.Name), &gorm.Config{})
		if err != nil {
			log.Fatalf("Erreur de connexion à la BDD : %v", err)
		}
		sqlDB, err := db.DB()
		if err != nil {
			log.Fatalf("FATAL: Échec de l'obtention de la DB SQL : %v", err)
		}
		//S'assurer que la connexion est fermée à la fin de l'exécution de la commande
		defer sqlDB.Close()

		// Initialiser les repositories et services nécessaires
		linkRepo := repository.NewLinkRepository(db)
		linkService := services.NewLinkService(linkRepo)

		// Appeler GetLinkStats pour récupérer le lien et ses statistiques.
		link, err := linkService.GetLinkByCode(shortCodeFlag)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				fmt.Fprintf(os.Stderr, "Aucun lien trouvé pour le code '%s'\n", shortCodeFlag)
				os.Exit(1)
			}
			log.Fatalf("Erreur lors de la récupération du lien : %v", err)
		}

		// Récupérer les statistiques
		totalClicks, err := linkService.GetStats(shortCodeFlag)
		if err != nil {
			log.Fatalf("Erreur lors de la récupération des stats : %v", err)
		}

		// Afficher le résultat
		fmt.Printf("Statistiques pour le code court: %s\n", link.Code)
		fmt.Printf("URL longue: %s\n", link.LongURL)
		fmt.Printf("Total de clics: %d\n", totalClicks)
	},
}

func init() {
	// Définir et marquer le flag --code comme requis
	StatsCmd.Flags().StringVarP(&shortCodeFlag, "code", "c", "", "Code court à interroger")
	StatsCmd.MarkFlagRequired("code")

	// Ajouter la commande à RootCmd
	cmd2.RootCmd.AddCommand(StatsCmd)
}

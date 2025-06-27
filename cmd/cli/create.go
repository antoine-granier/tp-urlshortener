package cli

import (
	"fmt"
	"log"
	"net/url" // Pour valider le format de l'URL
	"os"

	cmd2 "github.com/antoine-granier/urlshortener/cmd"
	"github.com/antoine-granier/urlshortener/internal/repository"
	"github.com/antoine-granier/urlshortener/internal/services"
	"github.com/spf13/cobra"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Faire une variable longURLFlag qui stockera la valeur du flag --url
var longURLFlag string

// CreateCmd représente la commande 'create'
var CreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Crée une URL courte à partir d'une URL longue.",
	Long: `Cette commande raccourcit une URL longue fournie et affiche le code court généré.

Exemple:
  url-shortener create --url="https://www.google.com/search?q=go+lang"`,
	Run: func(cmd *cobra.Command, args []string) {
		// Valider que le flag --url a été fourni
		if longURLFlag == "" {
			fmt.Fprintln(os.Stderr, "Erreur : le flag --url est requis")
			os.Exit(1)
		}

		// Validation basique du format de l'URL
		if _, err := url.ParseRequestURI(longURLFlag); err != nil {
			log.Fatalf("URL invalide : %v", err)
		}

		// Charger la configuration globale
		cfg := cmd2.Cfg
		if cfg == nil {
			log.Fatal("Configuration non initialisée")
		}

		// Initialiser la connexion à la base de données SQLite
		db, err := gorm.Open(sqlite.Open(cfg.Database.Name), &gorm.Config{})
		if err != nil {
			log.Fatalf("Erreur de connexion à la BDD : %v", err)
		}
		sqlDB, err := db.DB()
		if err != nil {
			log.Fatalf("Échec de l'obtention de la DB SQL : %v", err)
		}
		defer sqlDB.Close()

		// Initialiser les repositories et services nécessaires
		linkRepo := repository.NewLinkRepository(db)
		linkSvc := services.NewLinkService(linkRepo)

		// Créer le lien court
		link, err := linkSvc.CreateLink(longURLFlag)
		if err != nil {
			log.Fatalf("Erreur lors de la création du lien : %v", err)
		}

		// Afficher le résultat
		fullShortURL := fmt.Sprintf("%s/%s", cfg.Server.BaseURL, link.ShortCode)
		fmt.Println("URL courte créée avec succès:")
		fmt.Printf("Code: %s\n", link.ShortCode)
		fmt.Printf("URL complète: %s\n", fullShortURL)
	},
}

func init() {
	// Définir et marquer le flag --url comme requis
	CreateCmd.Flags().StringVarP(&longURLFlag, "url", "u", "", "URL à raccourcir")
	CreateCmd.MarkFlagRequired("url")

	// Ajouter la commande à RootCmd
	cmd2.RootCmd.AddCommand(CreateCmd)
}

package main

import (
	_ "github.com/antoine-granier/urlshortener/cmd/cli"    // Importe le package 'cli' pour que ses init() soient exécutés
	_ "github.com/antoine-granier/urlshortener/cmd/server" // Importe le package 'server' pour que ses init() soient exécutés
	- "github.com/antoine-granier/urlshortener/cmd"
)

func main() {
	// Exécute la commande racine de Cobra.
	cmd.Execute()
}

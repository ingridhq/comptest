package dbutil

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

// MustMakeMigrationsWithSeedSource builds a temporary directory with existing migrations and a seed SQL files,
// and returns the directory location in a format of a valid migrations source path for `migrate` tool.
//
// This is necessary since `migrate` requires all migration files to be in the same location.
func MustMakeMigrationsWithSeedSource(migrationsFS fs.FS, seedFS fs.FS) string {
	// Create temporary directory to store "production" migrations and seed migrations.
	tmpDir, _ := os.MkdirTemp("", "test-migrations")

	// Copy "production" migrations to a temporary directory with the same names.
	mff := mustReadDirectory(migrationsFS)
	for _, mf := range mff {
		outName := filepath.Join(tmpDir, mf.Name())
		mustCopyFile(migrationsFS, mf, outName)
	}

	// Generate a future timestamp for seed migrations to follow `migrate` name pattern.
	migrationTime := "20990101000000"

	// Copy seed migrations to the temporary directory with modified names.
	sff := mustReadDirectory(seedFS)
	for _, sf := range sff {
		outName := filepath.Join(tmpDir, fmt.Sprintf("%s_%s", migrationTime, sf.Name()))
		mustCopyFile(seedFS, sf, outName)
	}

	// Return a directory location in a `migrate` migrations source format.
	return fmt.Sprintf("file://%s/", tmpDir)
}

func mustCopyFile(migrationsFS fs.FS, fe fs.DirEntry, outName string) {
	bb := mustReadFile(migrationsFS, fe)
	mustWriteFile(outName, bb)
}

func mustReadDirectory(migrationsFS fs.FS) []fs.DirEntry {
	ff, err := fs.ReadDir(migrationsFS, ".")
	if err != nil {
		log.Fatalf("Failed to read directory: %v", err)
	}
	return ff
}

func mustReadFile(migrationsFS fs.FS, fe fs.DirEntry) []byte {
	bb, err := fs.ReadFile(migrationsFS, fe.Name())
	if err != nil {
		log.Fatalf("Failed to read file %q: %v", fe.Name(), err)
	}
	return bb
}
func mustWriteFile(outName string, bb []byte) {
	if err := ioutil.WriteFile(outName, bb, 0644); err != nil {
		log.Fatalf("Failed to write file %q: %v", outName, err)
	}
}

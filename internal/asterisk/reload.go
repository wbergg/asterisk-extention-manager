package asterisk

import (
	"database/sql"
	"fmt"
	"log"
	"os/exec"

	"github.com/wbergg/asterisk-extention-manager/internal/config"
)

func Reload(cfg *config.Config) error {
	cmd := exec.Command(cfg.AsteriskCmd, "-rx", "pjsip reload")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("asterisk reload: %w (output: %s)", err, string(output))
	}
	log.Printf("Asterisk reload output: %s", string(output))
	return nil
}

func SyncAndReload(db *sql.DB, cfg *config.Config) error {
	if err := WriteConfig(db, cfg); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	if err := Reload(cfg); err != nil {
		log.Printf("WARNING: asterisk reload failed: %v", err)
		return err
	}
	return nil
}

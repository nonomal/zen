package cfg

import (
	"context"
	"log"
	"os"
	"os/exec"

	"github.com/blang/semver"
	"github.com/rhysd/go-github-selfupdate/selfupdate"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// Version is the current version of the binary. Set at compile time using ldflags (see .github/workflows/build.yml).
var Version = "development"

// noSelfUpdate is set to "true" for builds distributed to package managers to prevent auto-updating.
// Set at compile time using ldflags (see .github/workflows/build.yml).
var noSelfUpdate = "false"

func (c *Config) GetVersion() string {
	return Version
}

// SelfUpdate checks for updates and prompts the user to update if there is one.
func SelfUpdate(ctx context.Context) {
	shouldUpdate, release := checkForUpdates()
	if !shouldUpdate {
		return
	}

	action, err := runtime.MessageDialog(ctx, runtime.MessageDialogOptions{
		Title:         "Would you like to update Zen?",
		Message:       release.ReleaseNotes,
		Buttons:       []string{"Yes", "No"},
		Type:          runtime.QuestionDialog,
		DefaultButton: "Yes",
		CancelButton:  "No",
	})
	if err != nil {
		log.Printf("error occurred while showing update dialog: %v", err)
		return
	}
	if action == "No" {
		return
	}

	v, _ := semver.ParseTolerant(Version)
	_, err = selfupdate.UpdateSelf(v, "anfragment/zen")
	if err != nil {
		log.Printf("error occurred while updating binary: %v", err)
		return
	}

	action, err = runtime.MessageDialog(ctx, runtime.MessageDialogOptions{
		Title:         "Zen has been updated",
		Message:       "Zen has been updated to the latest version. Would you like to restart it now?",
		Buttons:       []string{"Yes", "No"},
		Type:          runtime.QuestionDialog,
		DefaultButton: "Yes",
		CancelButton:  "No",
	})
	if err != nil {
		log.Printf("error occurred while showing restart dialog: %v", err)
	}
	if action == "Yes" {
		cmd := exec.Command(os.Args[0], os.Args[1:]...) // #nosec G204
		if err := cmd.Start(); err != nil {
			log.Printf("error occurred while restarting: %v", err)
			return
		}
		runtime.Quit(ctx)
	}
}

func checkForUpdates() (bool, *selfupdate.Release) {
	log.Println("checking for updates")
	if noSelfUpdate == "true" || Version == "development" {
		log.Println("self-update disabled")
		return false, nil
	}

	latest, found, err := selfupdate.DetectLatest("anfragment/zen")
	if err != nil {
		log.Printf("error occurred while detecting version: %v", err)
		return false, nil
	}

	v, _ := semver.ParseTolerant(Version) // ParseTolerant allows for the non-standard "v" prefix
	if !found || latest.Version.LTE(v) {
		log.Println("current version is up-to-date")
		return false, nil
	}

	return true, latest
}

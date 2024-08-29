package publish

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// PublishToGitHub publishes the generated content to the Hugo site repository and pushes to both repositories.
func PublishToGitHub() error {
	// Step 1: Ensure the submodule is correctly initialized and updated
	err := updateSubmoduleToMaster("./website")
	if err != nil {
		return fmt.Errorf("failed to update submodule to master: %w", err)
	}

	// Step 2: Move the generated content to the Hugo content directory
	generatedDir := "generated"
	targetDir := "./website/content/posts/"

	// Ensure the generated directory exists before attempting to move files
	if _, err := os.Stat(generatedDir); os.IsNotExist(err) {
		return fmt.Errorf("generated directory does not exist: %w", err)
	}

	// Ensure the target directory exists
	err = EnsureDir(targetDir)
	if err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// Move files from the generated directory to the Hugo content directory
	err = moveFiles(generatedDir, targetDir)
	if err != nil {
		return fmt.Errorf("failed to move files: %w", err)
	}

	// Step 3: Check if Hugo is installed
	_, err = exec.LookPath("hugo")
	if err != nil {
		return fmt.Errorf("Hugo is not installed. Please install Hugo before running this script.")
	}

	// Step 4: Build the Hugo site in the correct directory
	cmd := exec.Command("hugo")
	cmd.Dir = "./website" // Ensure this is pointing to the correct directory
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to build Hugo site: %w", err)
	}

	// Step 5: Commit and push changes to the main repository
	err = gitCommitAndPush(".", "Add new blog post")
	if err != nil {
		return fmt.Errorf("failed to push to main repository: %w", err)
	}

	// Step 6: Perform a fresh pull and force push to the Hugo site repository's public directory (master branch)
	err = gitFreshPullAndPushToMaster("./website/public", "Deploy new site version")
	if err != nil {
		return fmt.Errorf("failed to push to Hugo site deployment branch: %w", err)
	}

	return nil
}

// updateSubmoduleToMaster initializes, updates, checks out the master branch, and forcefully resets the submodule to match the remote master branch.
func updateSubmoduleToMaster(dir string) error {
	// Initialize and update the submodule
	cmd := exec.Command("git", "submodule", "update", "--init", "--recursive")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to update submodule: %w", err)
	}

	// Change directory to the submodule
	cmd = exec.Command("git", "checkout", "master")
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to checkout master branch in submodule: %w", err)
	}

	// Forcefully reset the master branch to match the remote
	cmd = exec.Command("git", "reset", "--hard", "origin/master")
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to forcefully reset submodule to origin/master: %w", err)
	}

	return nil
}

// moveFiles moves all files from srcDir to dstDir.
func moveFiles(srcDir, dstDir string) error {
	files, err := os.ReadDir(srcDir)
	if err != nil {
		return fmt.Errorf("failed to read source directory: %w", err)
	}

	for _, file := range files {
		srcPath := filepath.Join(srcDir, file.Name())
		dstPath := filepath.Join(dstDir, file.Name())

		err := os.Rename(srcPath, dstPath)
		if err != nil {
			return fmt.Errorf("failed to move file %s: %w", file.Name(), err)
		}
	}

	return nil
}

// EnsureDir ensures that a directory exists, creating it if necessary.
func EnsureDir(dirName string) error {
	err := os.MkdirAll(dirName, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dirName, err)
	}
	return nil
}

// gitCommitAndPush commits and pushes changes in the given directory with the provided commit message.
func gitCommitAndPush(dir, commitMessage string) error {
	cmd := exec.Command("git", "add", ".")
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to add changes: %w", err)
	}

	cmd = exec.Command("git", "commit", "-m", commitMessage)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		// If there's nothing to commit, just move on
		if strings.Contains(err.Error(), "nothing to commit") {
			fmt.Println("Nothing to commit.")
		} else {
			return fmt.Errorf("failed to commit changes: %w", err)
		}
	}

	// Get the current branch
	cmd = exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = dir
	currentBranch, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}

	// Trim the newline from the branch name
	branchName := string(currentBranch)
	branchName = branchName[:len(branchName)-1]

	cmd = exec.Command("git", "push", "origin", branchName)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to push changes: %w", err)
	}

	return nil
}

// gitFreshPullAndPushToMaster performs a fresh pull of the submodule, builds the site, and force-pushes the updates.
func gitFreshPullAndPushToMaster(dir, commitMessage string) error {
	// Step 1: Remove the existing submodule directory
	err := os.RemoveAll(dir)
	if err != nil {
		return fmt.Errorf("failed to remove submodule directory: %w", err)
	}

	// Step 2: Clone the submodule afresh
	cmd := exec.Command("git", "submodule", "update", "--init", "--recursive")
	cmd.Dir = filepath.Dir(dir) // Go to the parent directory of the submodule
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to clone submodule: %w", err)
	}

	// Step 3: Build the Hugo site (ensure this is in the correct directory)
	cmd = exec.Command("hugo")
	cmd.Dir = "./website" // This must point to the directory where Hugo's config files are located
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to build Hugo site: %w", err)
	}

	// Step 4: Add and commit the changes
	cmd = exec.Command("git", "add", ".")
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to add changes: %w", err)
	}

	cmd = exec.Command("git", "commit", "-m", commitMessage)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		// If there's nothing to commit, just move on
		if strings.Contains(err.Error(), "nothing to commit") {
			fmt.Println("Nothing to commit.")
		} else {
			return fmt.Errorf("failed to commit changes: %w", err)
		}
	}

	// Step 5: Force push the changes to the remote master branch
	cmd = exec.Command("git", "push", "origin", "master", "--force")
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to force push changes: %w", err)
	}

	return nil
}

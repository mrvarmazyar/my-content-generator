package publish

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// createShortenedSlug creates a shorter version of the slug suitable for branch names
func createShortenedSlug(slug string) string {
	// Shorten the slug to the first 50 characters, and replace spaces with hyphens
	if len(slug) > 50 {
		slug = slug[:50]
	}
	return strings.ReplaceAll(slug, " ", "-")
}

// PublishToGitHub publishes the generated content by creating a new branch, pushing it, and creating a PR.
func PublishToGitHub(title, slug string) error {
	// Step 1: Ensure the submodule is correctly initialized and updated
	err := updateSubmoduleToMaster("./website")
	if err != nil {
		return fmt.Errorf("failed to update submodule to master: %w", err)
	}

	// Step 2: Move the generated content to the Hugo content directory
	generatedDir := "generated"
	targetDir := "./website/content/posts/"

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

	// Step 5: Commit the changes in the content/posts/ directory to master
	err = gitCommitAndPush("./website", "Add new blog post")
	if err != nil {
		return fmt.Errorf("failed to commit changes: %w", err)
	}

	// Step 6: Create a new branch from the latest master after committing the changes
	shortenedSlug := createShortenedSlug(slug)
	err = gitCreateAndPushBranch("./website", shortenedSlug)
	if err != nil {
		return fmt.Errorf("failed to create and push new branch: %w", err)
	}

	// Step 7: Create a pull request to merge the new branch into master
	err = createPullRequest(shortenedSlug, title)
	if err != nil {
		return fmt.Errorf("failed to create pull request: %w", err)
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

	// Stash any uncommitted changes to prevent issues during checkout
	cmd = exec.Command("git", "stash")
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil && !strings.Contains(err.Error(), "No local changes to save") {
		return fmt.Errorf("failed to stash changes in submodule: %w", err)
	}

	// Checkout the master branch
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

	// Apply the stashed changes back if necessary
	cmd = exec.Command("git", "stash", "pop")
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil && !strings.Contains(err.Error(), "No stash entries found.") {
		return fmt.Errorf("failed to apply stashed changes: %w", err)
	}

	return nil
}

// gitCommitAndPush stages and commits changes in the given directory
func gitCommitAndPush(dir, commitMessage string) error {
	// Stage all changes
	cmd := exec.Command("git", "add", ".")
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to add changes: %w", err)
	}

	// Commit the changes
	cmd = exec.Command("git", "commit", "-m", commitMessage)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		// If there's nothing to commit, log a message but continue
		if strings.Contains(err.Error(), "nothing to commit") {
			fmt.Println("Nothing to commit.")
			return nil
		}
		return fmt.Errorf("failed to commit changes: %w", err)
	}

	// Push the changes to master
	cmd = exec.Command("git", "push", "origin", "master")
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to push changes to master: %w", err)
	}

	return nil
}

// gitCreateAndPushBranch creates and pushes a new branch after committing changes to master
func gitCreateAndPushBranch(dir, branchName string) error {
	// Create and switch to the new branch
	cmd := exec.Command("git", "checkout", "-b", branchName)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to create and switch to new branch: %w", err)
	}

	// Push the new branch to the remote repository
	cmd = exec.Command("git", "push", "origin", branchName)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to push new branch: %w", err)
	}

	return nil
}

// createPullRequest creates a pull request from the new branch to the master branch.
func createPullRequest(branchName, title string) error {
	// Assuming we have the GitHub CLI (gh) installed and authenticated.
	cmd := exec.Command("gh", "pr", "create", "--base", "master", "--head", branchName, "--title", title, "--body", "Automated PR created by content generator.")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to create pull request: %w", err)
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

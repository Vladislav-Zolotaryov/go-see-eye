package main

import (
	"errors"
	"os"
	"os/exec"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
	"github.com/gofiber/fiber/v3/middleware/filesystem"
)

const main_repository_path = "repositories/"

func initializeRepository(repository_name string) (string, error) {
	repository_path := main_repository_path + repository_name

	if _, err := os.Stat(repository_path); os.IsNotExist(err) {
		err := os.MkdirAll(repository_path, 0755)
		if err != nil {
			panic("Failed to create directory " + err.Error())
		}
	}

	cmd := exec.Command("git", "init", "--bare", "--shared=group")
	cmd.Dir = repository_path
	stdout, err := cmd.Output()
	log.Info(string(stdout))

	if err != nil {
		log.Info(err.Error())
		return "", errors.New("failed to create a repositry")
	}

	cmd = exec.Command("git", "update-server-info")
	cmd.Dir = repository_path
	stdout, err = cmd.Output()
	if err != nil {
		log.Info(err.Error())
		return "", errors.New("failed to update server info in a repositry")
	}
	log.Info(string(stdout))

	hook_file := repository_path + "/hooks/post-update"

	file, errs := os.Create(hook_file)
	if errs != nil {
		log.Info("Failed to create file:", errs)
	}
	defer file.Close()

	_, errs = file.WriteString(
		"exec git update-server-info\nexec curl -X PUT http://localhost:3000/repository/" + repository_name + "/commit-hook")
	if errs != nil {
		log.Info("Failed to write to file:", errs)
	}

	err2 := os.Chmod(hook_file, 0755)
	if err2 != nil {
		log.Info(err2)
	}

	return string(stdout), nil
}

func main() {
	app := fiber.New()

	repository_fs := os.DirFS("./" + main_repository_path)

	app.Use(filesystem.New(filesystem.Config{
		Root:   repository_fs,
		Browse: true,
	}))

	app.Put("/repository/:name", func(c fiber.Ctx) error {
		_, err := initializeRepository(c.Params("name"))
		return err
	})

	app.Put("/repository/:name/commit-hook", func(c fiber.Ctx) error {
		log.Info("A commit has been pushed!")
		return nil
	})

	log.Fatal(app.Listen(":3000"))
}
